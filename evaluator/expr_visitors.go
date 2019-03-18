package evaluator

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/10gen/sqlproxy/internal/option"
)

// joinOnExpression holds the SQLColumnExpr value
// used within a join stage matcher.
type joinOnExpression struct {
	exprCollector *sqlColExprCollector
}

func (v *joinOnExpression) visit(n Node) (Node, error) {
	switch typedN := n.(type) {
	case *JoinStage:
		_, err := v.exprCollector.visit(typedN.matcher)
		if err != nil {
			return nil, err
		}
		return typedN, nil
	default:
		return walk(v, n)
	}
}

// sqlColExprCounter is a used to hold and count all
// SQLColumnExpr values found during a Node visit.
type sqlColExprCounter struct {
	counts map[string]int
	exprs  []SQLColumnExpr
}

// newSQLColumnExprCounter returns a new sqlColExprCounter.
func newSQLColumnExprCounter() *sqlColExprCounter {
	return &sqlColExprCounter{
		counts: make(map[string]int),
	}
}

func (c *sqlColExprCounter) add(e SQLColumnExpr) {
	s := e.String()
	if _, ok := c.counts[s]; ok {
		c.counts[s]++
	} else {
		c.counts[s] = 1
		c.exprs = append(c.exprs, e)
	}
}

func (c *sqlColExprCounter) copyExprs() []SQLColumnExpr {
	exprs := make([]SQLColumnExpr, len(c.exprs))
	copy(exprs, c.exprs)
	return exprs
}

func (c *sqlColExprCounter) remove(e SQLColumnExpr) {
	s := e.String()
	for i, expr := range c.exprs {
		if strings.EqualFold(s, expr.String()) {
			c.counts[s]--
			if c.counts[s] == 0 {
				delete(c.counts, s)
				c.exprs = append(c.exprs[:i], c.exprs[i+1:]...)
			}
			return
		}
	}
}

// sqlColExprCollector is used to track which SQLColumnExpr values
// have been visited and the select context in which they were found.
type sqlColExprCollector struct {
	selectIDs         []int
	referencedColumns *sqlColExprCounter
	removeMode        bool
}

// newSQLColExprCollector returns a new sqlColExprCollector.
func newSQLColExprCollector(selectIDs []int) *sqlColExprCollector {
	return &sqlColExprCollector{
		selectIDs:         selectIDs,
		referencedColumns: newSQLColumnExprCounter(),
	}
}

// Remove visits and removes the SQLColumnExpr
// values within e from the expression collector.
func (c *sqlColExprCollector) Remove(e SQLExpr) {
	c.removeMode = true
	c.Add(e)
	c.removeMode = false
}

// RemoveAll visits and removes the SQLColumnExpr values
// within each element of the slice, e, from the
// expression collector.
func (c *sqlColExprCollector) RemoveAll(e []SQLExpr) {
	c.removeMode = true
	c.AddAll(e)
	c.removeMode = false
}

// AddAll visits and adds the SQLColumnExpr values
// within each element of the slice, e, to the
// expression collector.
func (c *sqlColExprCollector) AddAll(exprs []SQLExpr) {
	for _, e := range exprs {
		c.Add(e)
	}
}

// Add visits and adds the SQLColumnExpr values
// within e to the expression collector.
func (c *sqlColExprCollector) Add(e SQLExpr) {
	_, err := c.visit(e)
	// This err was previously ignored.
	if err != nil {
		panic(err)
	}
}

func (c *sqlColExprCollector) visit(n Node) (Node, error) {
	switch typedN := n.(type) {
	case SQLColumnExpr:
		if containsInt(c.selectIDs, typedN.selectID) {
			if c.removeMode {
				c.referencedColumns.remove(typedN)
			} else {
				c.referencedColumns.add(typedN)
			}
		}
		return typedN, nil
	case *SQLSubqueryExpr:
		if typedN.correlated {
			return walk(c, n)
		}
		return n, nil
	default:
		return walk(c, n)
	}
}

// getJoinOnExpressions gets all column expressions referenced in
// a join 'on' clause in the given plan stage.
func (c *sqlColExprCollector) getJoinOnExpressions(ps PlanStage) {
	v := &joinOnExpression{c}
	_, err := v.visit(ps)
	if err != nil {
		panic(err) // This error should always be nil.
	}
}

func explainQuery(plan PlanStage, pCfg *PushdownConfig) ([]*ExplainRecord, error) {

	visitor := newExplainVisitor()
	_, err := visitor.visit(plan)
	if err != nil {
		return nil, err
	}

	var pushdownFailures map[PlanStage][]PushdownFailure

	_, err = PushdownPlan(pCfg, plan)
	pde, ok := err.(PushdownError)
	if err != nil && ok {
		pushdownFailures = pde.Failures()
	}

	explain := []*ExplainRecord{}
	for stg, rec := range visitor.records {
		failures, ok := pushdownFailures[stg]
		if ok {
			rec.PushdownFailures = failures
		}
		explain = append(explain, rec)
	}

	sort.Slice(explain, func(i, j int) bool {
		return explain[i].ID < explain[j].ID
	})

	return explain, nil
}

// explainVisitor will visit each stage of the explain plan.
type explainVisitor struct {
	// records contains the ExplainRecord generated for each stage in the visited plan.
	records map[PlanStage]*ExplainRecord
	// currentStageID keeps track of the number of stages visited to generate unique IDs
	currentStageID int
}

func newExplainVisitor() *explainVisitor {
	return &explainVisitor{
		records:        make(map[PlanStage]*ExplainRecord),
		currentStageID: 0,
	}
}

// ExplainRecord contains explain plan data for one stage in a query plan.
type ExplainRecord struct {
	ID               int
	StageType        string
	Columns          string
	Sources          []int
	Database         option.String
	Tables           option.String
	Aliases          option.String
	Collections      option.String
	Pipeline         option.String
	PipelineExplain  option.String
	PushdownFailures []PushdownFailure
}

// NewExplainRecord returns a new ExplainRecord with the provided fields.
func NewExplainRecord(id int, stageType string, columns string, sources []int, database option.String, tables option.String, aliases option.String, collections option.String, pipeline option.String, pipelineExplain option.String, failures []PushdownFailure) *ExplainRecord {
	return &ExplainRecord{
		ID:               id,
		StageType:        stageType,
		Columns:          columns,
		Sources:          sources,
		Database:         database,
		Tables:           tables,
		Aliases:          aliases,
		Collections:      collections,
		Pipeline:         pipeline,
		PipelineExplain:  pipelineExplain,
		PushdownFailures: failures,
	}
}

func (v *explainVisitor) visit(n Node) (Node, error) {

	switch typedN := n.(type) {
	case *MongoSourceStage:
		v.currentStageID++
		curr := v.currentStageID

		rec := v.generateExplainRecord(typedN, curr)
		v.records[typedN] = rec
	case PlanStage:
		v.currentStageID++
		curr := v.currentStageID

		_, err := walk(v, n)
		if err != nil {
			return nil, err
		}

		rec := v.generateExplainRecord(typedN, curr)
		v.records[typedN] = rec
	}

	return n, nil
}

func jsonList(strs []string) string {
	return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
}

// generateStageRecord will create a row for the explain plan table from a plan stage.
func (v *explainVisitor) generateExplainRecord(stage PlanStage, curr int) *ExplainRecord {

	var stageType string
	var sourceNodes []int
	var database, tables, aliases, collections, pipeline option.String

	// get stage name
	switch typedN := stage.(type) {
	case *UnionStage:
		if typedN.kind == UnionAll {
			stageType = "UnionAll"
		} else {
			stageType = "UnionDistinct"
		}
	case *JoinStage:
		stageType = string(typedN.kind)
	default:
		if t := reflect.TypeOf(stage); t.Kind() == reflect.Ptr {
			stageType = t.Elem().Name()
		} else {
			stageType = t.Name()
		}
	}

	// get source nodes
	sourceNodes = getSources(stage, v)

	// get mongosource-specific fields
	ms, ok := stage.(*MongoSourceStage)
	if ok {
		sourceNodes = nil
		database = option.SomeString(ms.dbName)
		tables = option.SomeString(jsonList(ms.tableNames))
		aliases = option.SomeString(jsonList(ms.aliasNames))
		collections = option.SomeString(jsonList(ms.collectionNames))
		pipeline = option.SomeString(ms.PipelineString())
	}

	return NewExplainRecord(
		curr,                            // ID
		stageType,                       // StageType
		getPlanColumns(stage.Columns()), // Columns
		sourceNodes,                     // Sources
		database,                        // Database
		tables,                          // Tables
		aliases,                         // Aliases
		collections,                     // Collections
		pipeline,                        // Pipeline
		option.NoneString(),             // PipelineExplain
		nil,                             // PushdownFailures
	)
}

func getSources(stage PlanStage, v *explainVisitor) []int {
	children := stage.Children()
	ret := make([]int, 0, len(children))
	for _, child := range children {
		if planStage, ok := child.(PlanStage); ok {
			ret = append(ret, v.records[planStage].ID)
		}
	}
	return ret
}

type selectIDGatherer struct {
	selectIDs []int
}

func gatherSelectIDs(n Node) []int {
	v := &selectIDGatherer{}
	_, err := v.visit(n)
	if err != nil {
		panic(fmt.Errorf("selectIDGatherer returned unexpected error: %v", err))
	}
	return v.selectIDs
}

func (v *selectIDGatherer) visit(n Node) (Node, error) {
	n, err := walk(v, n)
	if err != nil {
		return nil, err
	}

	switch typedN := n.(type) {
	case SQLColumnExpr:
		v.selectIDs = append(v.selectIDs, typedN.selectID)
	}

	return n, nil
}

type databaseNameFinder struct {
	databaseName         string
	hasMultipleDatabases bool
}

// getDatabaseName returns the name of the database from where the SQLColumnExpr values in n are
// derived. It returns the empty string if the values come from more than one database or the
// dual database.
func getDatabaseName(n Node) string {
	v := &databaseNameFinder{}
	_, err := v.visit(n)
	if err != nil {
		panic(fmt.Errorf("databaseNameFinder returned unexpected error: %v", err))
	}
	if v.hasMultipleDatabases {
		return ""
	}
	return v.databaseName
}

func (v *databaseNameFinder) visit(n Node) (Node, error) {
	if v.hasMultipleDatabases {
		return n, nil
	}

	n, err := walk(v, n)
	if err != nil {
		return nil, err
	}

	switch typedN := n.(type) {
	case SQLColumnExpr:
		if v.databaseName == "" {
			v.databaseName = typedN.databaseName
		} else if typedN.databaseName != v.databaseName {
			v.databaseName = ""
			v.hasMultipleDatabases = true
		}
	}
	return n, nil
}
