package evaluator

import (
	"fmt"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
)

type pipelineGatherer struct {
	pipelines [][]bson.D
}

func (v *pipelineGatherer) visit(n node) (node, error) {
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

func CreateProjectedColumnFromSQLExpr(selectID int, columnName string, expr SQLExpr) ProjectedColumn {
	column := &Column{
		SelectID: selectID,
		Name:     columnName,
		SQLType:  expr.Type(),
	}

	if sqlColExpr, ok := expr.(SQLColumnExpr); ok {
		column.MongoType = sqlColExpr.columnType.MongoType
	}

	return ProjectedColumn{Column: column, Expr: expr}
}

func CreateTestVariables(info *mongodb.Info) *variable.Container {
	gbl := variable.NewGlobalContainer(nil)
	gbl.MongoDBInfo = info
	ctn := variable.NewSessionContainer(gbl)
	ctn.MongoDBInfo = info
	return ctn
}

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

func GetCatalogFromSchema(schema *schema.Schema, variables *variable.Container) *catalog.Catalog {
	c, err := catalog.Build(schema, variables)
	if err != nil {
		panic(fmt.Sprintf("unable to build catalog: %v", err))
	}
	return c
}

// GetMongoDBInfo returns Info without looking up the information in MongoDB by setting
// all privileges to the specified privileges.
func GetMongoDBInfo(versionArray []uint8, sch *schema.Schema, privileges mongodb.Privilege) *mongodb.Info {
	if len(versionArray) == 0 {
		versionArray = []uint8{3, 4, 0}
	}

	versionString := ""

	for _, entry := range versionArray {
		versionString = fmt.Sprintf("%v.", entry)
	}

	i := &mongodb.Info{
		Privileges:   privileges,
		Databases:    make(map[mongodb.DatabaseName]*mongodb.DatabaseInfo),
		Version:      versionString[1:],
		VersionArray: versionArray,
	}

	for _, db := range sch.Databases {
		dbInfo := &mongodb.DatabaseInfo{
			Privileges:  privileges,
			Name:        mongodb.DatabaseName(db.Name),
			Collections: make(map[mongodb.CollectionName]*mongodb.CollectionInfo),
		}

		i.Databases[dbInfo.Name] = dbInfo

		for _, col := range db.Tables {
			if _, ok := dbInfo.Collections[mongodb.CollectionName(col.Name)]; ok {
				continue
			}

			colInfo := &mongodb.CollectionInfo{
				Privileges: privileges,
				Name:       mongodb.CollectionName(col.Name),
			}

			dbInfo.Collections[colInfo.Name] = colInfo
		}
	}

	return i
}

func GetSQLExpr(schema *schema.Schema, dbName, tableName, sql string) (SQLExpr, error) {
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

func GetProjectProjectedColumnExpr(plan PlanStage) SQLExpr {
	return (plan.(*ProjectStage)).projectedColumns[0].Expr
}

func GetNodePipeline(n node) [][]bson.D {
	pg := &pipelineGatherer{}
	pg.visit(n)
	return pg.pipelines
}

type subqueryFinder struct {
	count         int
	firstSubquery *SQLSubqueryExpr
}

func (v *subqueryFinder) visit(n node) (node, error) {
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

func GetSubqueryPlan(optimized node) PlanStage {
	finder := &subqueryFinder{}
	finder.visit(optimized)
	return finder.firstSubquery.plan
}

type cacheStageCounter struct {
	count int
}

func (v *cacheStageCounter) visit(n node) (node, error) {
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

func GetCacheStateCount(optimized node) int {
	cacheCounter := &cacheStageCounter{}
	cacheCounter.visit(optimized)
	return cacheCounter.count
}

type SourceStageReplacer struct {
	Data            []bson.D
	Existing        int
	Replaced        int
	LastSourceStage *BSONSourceStage
}

func (v *SourceStageReplacer) visit(n node) (node, error) {
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
		bs := NewBSONSourceStage(typedN.selectIDs[0], typedN.tableNames[0], typedN.collation, v.Data[0:1])
		v.Data = v.Data[1:]
		v.Replaced++
		n = bs
	}

	return n, nil
}

func (v *SourceStageReplacer) VisitStage(n node) (node, error) {
	return v.visit(n)
}
