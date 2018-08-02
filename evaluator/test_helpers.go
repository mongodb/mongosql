package evaluator

import (
	"fmt"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/drdl"
	"github.com/10gen/sqlproxy/variable"
)

type pipelineGatherer struct {
	pipelines [][]bson.D
}

func (v *pipelineGatherer) visit(n Node) (Node, error) {
	n, err := walk(v, n)
	if err != nil {
		return nil, err
	}

	switch typedN := n.(type) {
	case *MongoSourceStage:
		if len(typedN.pipeline) > 0 {
			pipeline := make([]bson.D, len(typedN.pipeline))
			copy(pipeline, typedN.pipeline)
			v.pipelines = append(v.pipelines, pipeline)
		}
	}

	return n, nil
}

// CreateProjectedColumnFromSQLExpr creates a projected column for a selectID,
// column name and a SQLExpr.
func CreateProjectedColumnFromSQLExpr(selectID int,
	columnName string,
	expr SQLExpr) ProjectedColumn {
	column := &Column{
		SelectID: selectID,
		Name:     columnName,
		ColumnType: *NewColumnType(
			expr.EvalType(),
			schema.MongoNone,
		),
	}

	column.Database = getDatabaseName(expr)
	if sqlColExpr, ok := expr.(SQLColumnExpr); ok {
		column.MongoType = sqlColExpr.columnType.MongoType
	}

	return ProjectedColumn{Column: column, Expr: expr}
}

// CreateTestVariables creates a container from a mongoDB config for testing.
func CreateTestVariables(info *mongodb.Info) *variable.Container {
	gbl := variable.NewGlobalContainer(nil)
	gbl.MongoDBInfo = info
	ctn := variable.NewSessionContainer(gbl)
	ctn.MongoDBInfo = info
	return ctn
}

// GetBinaryExprLeaves gets the left and right leaves of binary SQLExprs.
func GetBinaryExprLeaves(expr SQLExpr) (SQLExpr, SQLExpr) {
	switch typedE := expr.(type) {
	case *SQLAndExpr:
		return typedE.left, typedE.right
	case *SQLAddExpr:
		return typedE.left, typedE.right
	case *SQLSubtractExpr:
		return typedE.left, typedE.right
	case *SQLMultiplyExpr:
		return typedE.left, typedE.right
	case *SQLDivideExpr:
		return typedE.left, typedE.right
	case *SQLEqualsExpr:
		return typedE.left, typedE.right
	case *SQLLessThanExpr:
		return typedE.left, typedE.right
	case *SQLGreaterThanExpr:
		return typedE.left, typedE.right
	case *SQLLessThanOrEqualExpr:
		return typedE.left, typedE.right
	case *SQLGreaterThanOrEqualExpr:
		return typedE.left, typedE.right
	case *SQLLikeExpr:
		return typedE.left, typedE.right
	case *SQLSubqueryExpr:
		return nil, &SQLTupleExpr{typedE.Exprs()}
	//case *SQLSubqueryCmpExpr:
	// return typedE.left, &SQLTupleExpr{typedE.value.exprs}
	case *SQLInExpr:
		return typedE.left, typedE.right
	}
	return nil, nil
}

// GetCatalogFromSchema builds a catalog for a schema and container.
func GetCatalogFromSchema(schema *schema.Schema, variables *variable.Container) *catalog.Catalog {
	c, err := catalog.Build(schema, variables)
	if err != nil {
		panic(fmt.Sprintf("unable to build catalog: %v", err))
	}
	return c
}

// GetMongoDBInfo returns Info without looking up the information in MongoDB by setting
// all privileges to the specified privileges.
func GetMongoDBInfo(versionArray []uint8,
	sch *schema.Schema,
	privileges mongodb.Privilege) *mongodb.Info {
	if len(versionArray) == 0 {
		versionArray = []uint8{3, 4, 0}
	}

	versionString := ""

	for _, entry := range versionArray {
		versionString = fmt.Sprintf("%v.", entry)
	}

	i := &mongodb.Info{
		ClusterPrivileges: privileges,
		Databases:         make(map[mongodb.DatabaseName]*mongodb.DatabaseInfo),
		Version:           versionString[1:],
		VersionArray:      versionArray,
	}

	for _, db := range sch.Databases() {
		dbInfo := &mongodb.DatabaseInfo{
			Privileges:  privileges,
			Name:        mongodb.DatabaseName(db.Name()),
			Collections: make(map[mongodb.CollectionName]*mongodb.CollectionInfo),
		}

		i.Databases[dbInfo.Name] = dbInfo

		for _, col := range db.Tables() {
			if _, ok := dbInfo.Collections[mongodb.CollectionName(col.MongoName())]; ok {
				continue
			}

			colInfo := &mongodb.CollectionInfo{
				Privileges: privileges,
				Name:       mongodb.CollectionName(col.MongoName()),
			}

			dbInfo.Collections[colInfo.Name] = colInfo
		}
	}

	return i
}

// GetSQLExpr translates a SQL statement into a SQLExpr.
func GetSQLExpr(schema *schema.Schema,
	dbName,
	tableName,
	sql string) (SQLExpr,
	error) {
	statement, err := parser.Parse("select " + sql + " from " + tableName)
	if err != nil {
		return nil, err
	}

	selectStatement := statement.(parser.SelectStatement)
	info := GetMongoDBInfo(nil, schema, mongodb.AllPrivileges)
	vars := CreateTestVariables(info)
	catalog := GetCatalogFromSchema(schema, vars)
	actualPlan, err := AlgebrizeQuery(selectStatement, dbName, vars, catalog)
	if err != nil {
		return nil, err
	}

	// Depending on the "sql" expression we are getting, the algebrizer could have put it in
	// either the ProjectStage (for non-aggregate expressions) or a GroupByStage (for aggregate
	// expressions). We don't know which one the user is asking for, so we'll assume the
	// GroupByStage if it exists, otherwise the ProjectStage.
	project := actualPlan.(*ProjectStage)
	expr := project.projectedColumns[0].Expr

	group, ok := project.source.(*GroupByStage)
	if ok {
		expr = group.projectedColumns[0].Expr
	}

	if conv, ok := expr.(*SQLConvertExpr); ok {
		expr = conv.expr
	}

	return expr, nil
}

// GetProjectProjectedColumnExpr returns the SQLExpr for the first projected
// column in a ProjectStage.
func GetProjectProjectedColumnExpr(plan PlanStage) SQLExpr {
	return (plan.(*ProjectStage)).projectedColumns[0].Expr
}

// GetNodePipeline walks a node and returns all the aggregation pipelines found.
func GetNodePipeline(n Node) [][]bson.D {
	pg := &pipelineGatherer{}
	_, err := pg.visit(n)
	// This err was previously ignored.
	if err != nil {
		panic(err)
	}
	return pg.pipelines
}

type subqueryFinder struct {
	count         int
	firstSubquery *SQLSubqueryExpr
}

func (v *subqueryFinder) visit(n Node) (Node, error) {
	n, err := walk(v, n)
	if err != nil {
		return nil, err
	}

	switch typedN := n.(type) {
	case *SQLSubqueryExpr:
		v.count++
		v.firstSubquery = typedN
	}

	return n, nil
}

// GetSubqueryPlan walks a node and returns the PlanStage of the first subquery found.
func GetSubqueryPlan(optimized Node) PlanStage {
	finder := &subqueryFinder{}
	_, err := finder.visit(optimized)
	// This err was previously ignored.
	if err != nil {
		panic(err)
	}
	return finder.firstSubquery.plan
}

type cacheStageCounter struct {
	count int
}

func (v *cacheStageCounter) visit(n Node) (Node, error) {
	n, err := walk(v, n)
	if err != nil {
		return nil, err
	}

	switch n.(type) {
	case *CacheStage:
		v.count++
	}
	return n, nil
}

// GetCacheStateCount walks a node and counts the number of CacheStages found.
func GetCacheStateCount(optimized Node) int {
	cacheCounter := &cacheStageCounter{}
	_, err := cacheCounter.visit(optimized)
	// This err was previously ignored.
	if err != nil {
		panic(err)
	}
	return cacheCounter.count
}

// MustLoadSchema loads a schema from the provided DRDL bytes, and panics if any
// error is encountered.
func MustLoadSchema(schemaBytes []byte) *schema.Schema {
	drdlSchema, err := drdl.NewFromBytes(schemaBytes)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}

	testSchema, err := schema.NewFromDRDL(log.GlobalLogger(), drdlSchema)
	if err != nil {
		panic(fmt.Sprintf("Error loading schema: %v", err))
	}

	return testSchema
}

// SourceStageReplacer is a walker that replaces MongoSourceStages with
// BSONSourceStages for testing.
type SourceStageReplacer struct {
	Data            []bson.D
	Existing        int
	Replaced        int
	LastSourceStage *BSONSourceStage
}

func (v *SourceStageReplacer) visit(n Node) (Node, error) {
	n, err := walk(v, n)
	if err != nil {
		return nil, err
	}

	switch typedN := n.(type) {
	case *BSONSourceStage:
		v.Existing++
		if v.LastSourceStage == nil {
			v.LastSourceStage = typedN
		}
	case *MongoSourceStage:
		bs := NewBSONSourceStage(typedN.selectIDs[0],
			typedN.tableNames[0],
			typedN.collation,
			v.Data[0:1])
		v.Data = v.Data[1:]
		v.Replaced++
		n = bs
	}

	return n, nil
}

// VisitStage walks a node for SourceStageReplacer.
func (v *SourceStageReplacer) VisitStage(n Node) (Node, error) {
	return v.visit(n)
}
