package evaluator

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/10gen/sqlproxy/internal/option"
)

// PlanStats contains some statistics about a query plan.
type PlanStats struct {
	FullyPushedDown bool
	Explain         []*ExplainRecord
}

func getPlanStats(plan PlanStage, pushdownFailures map[PlanStage][]PushdownFailure) (*PlanStats, error) {
	pushedDown := IsFullyPushedDown(plan) == nil

	explain, err := explainQuery(plan, pushdownFailures)
	if err != nil {
		return nil, err
	}

	stats := &PlanStats{
		FullyPushedDown: pushedDown,
		Explain:         explain,
	}
	return stats, nil
}

func explainQuery(plan PlanStage, pushdownFailures map[PlanStage][]PushdownFailure) ([]*ExplainRecord, error) {

	visitor := newExplainVisitor()
	_, err := visitor.visit(plan)
	if err != nil {
		return nil, err
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
	// currentStageID keeps track of the number of stages visited to generate unique IDs.
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
		stageType = "JoinStage (" + string(typedN.kind) + ")"
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
