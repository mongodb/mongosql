package evaluator

import (
	"bytes"
	"fmt"

	"github.com/10gen/sqlproxy/internal/astutil"
)

// PrettyPrintCommand takes a command and prints it out.
func PrettyPrintCommand(c Command) string {
	return prettyPrintNode(c)
}

// PrettyPrintPlan takes a plan and recursively prints its source.
func PrettyPrintPlan(p PlanStage) string {
	return prettyPrintNode(p)
}

func prettyPrintNode(n Node) string {
	b := bytes.NewBufferString("")

	prettyPrint(b, n, 0)

	return b.String()
}

func prettyPrint(b *bytes.Buffer, n Node, d int) {

	astutil.PrintTabs(b, d)

	switch typedN := n.(type) {
	case *BSONSourceStage:
		b.WriteString("↳ BSONSource(")
		for i, c := range typedN.Columns() {
			if i != 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%v", c.Name))
		}
		b.WriteString(")")
	case *CountStage:
		b.WriteString(fmt.Sprintf("↳ Count: %s (db: %s, collection: %s)",
			typedN.mongoSource.tableNames[0],
			typedN.mongoSource.dbName,
			typedN.mongoSource.collectionNames[0]))
		if typedN.mongoSource.aliasNames[0] != typedN.mongoSource.tableNames[0] {
			b.WriteString(fmt.Sprintf(" as '%v'", typedN.mongoSource.aliasNames[0]))
		}
	case *DropTableCommand:
		if typedN.ifExists {
			b.WriteString(fmt.Sprintf("↳ DropTable if exists (%s)", typedN.tableName))
		} else {
			b.WriteString(fmt.Sprintf("↳ DropTable (%s)", typedN.tableName))
		}
	case *DropDatabaseCommand:
		if typedN.ifExists {
			b.WriteString(fmt.Sprintf("↳ DropDatabase if exists (%s)", typedN.dbName))
		} else {
			b.WriteString(fmt.Sprintf("↳ DropDatabase (%s)", typedN.dbName))
		}
	case *CreateTableCommand:
		if typedN.ifNotExists {
			b.WriteString(fmt.Sprintf("↳ CreateTable if not exists (%s)", typedN.table.SQLName()))
		} else {
			b.WriteString(fmt.Sprintf("↳ CreateTable (%s)", typedN.table.SQLName()))
		}
	case *CreateDatabaseCommand:
		if typedN.ifNotExists {
			b.WriteString(fmt.Sprintf("↳ CreateDatabase if not exists (%s)", typedN.dbName))
		} else {
			b.WriteString(fmt.Sprintf("↳ CreateDatabase (%s)", typedN.dbName))
		}
	case *InsertCommand:
		b.WriteString(fmt.Sprintf("↳ Insert (%s.%s)", typedN.dbName, typedN.tableName))
	case *DynamicSourceStage:
		b.WriteString(fmt.Sprintf("↳ DynamicSource (%s)", typedN.aliasName))
	case *DualStage:
		b.WriteString("↳ Dual")
	case *EmptyStage:
		b.WriteString("↳ Empty")
	case *ExplainStage:
		b.WriteString("↳ Explain")
	case *FilterStage:
		b.WriteString(fmt.Sprintf("↳ Filter (%v):\n", typedN.matcher))

		prettyPrint(b, typedN.source, d+1)
	case *FlushCommand:
		switch typedN.kind {
		case FlushLogs:
			b.WriteString(fmt.Sprintf("↳ Flush Logs"))
		case FlushSample:
			b.WriteString(fmt.Sprintf("↳ Flush Sample"))
		default:
			b.WriteString(fmt.Sprintf("↳ Flush <unknown>"))
		}
	case *GroupByStage:
		b.WriteString("↳ GroupBy(")
		for i, key := range typedN.keys {
			if i != 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%v", key.String()))
		}
		b.WriteString("):\n")

		prettyPrint(b, typedN.source, d+1)
	case *JoinStage:
		b.WriteString("↳ Join:\n")

		prettyPrint(b, typedN.left, d+1)
		astutil.PrintTabs(b, d+1)

		b.WriteString(fmt.Sprintf("%v\n", typedN.kind))
		prettyPrint(b, typedN.right, d+1)

		if typedN.matcher != nil {
			astutil.PrintTabs(b, d+1)
			b.WriteString(fmt.Sprintf("on %v\n", typedN.matcher.String()))
		}
	case *KillCommand:
		scope := "connection"
		if typedN.Scope == KillQuery {
			scope = "query"
		}
		b.WriteString(fmt.Sprintf("↳ Kill %v %v", scope, typedN.ID))
	case *LimitStage:
		b.WriteString(fmt.Sprintf("↳ Limit(offset: %v, limit: %v):\n", typedN.offset, typedN.limit))

		prettyPrint(b, typedN.source, d+1)
	case *MongoSourceStage:
		b.WriteString(fmt.Sprintf("↳ MongoSource: '%v' (db: '%v', collection: '%v')",
			typedN.tableNames,
			typedN.dbName,
			typedN.collectionNames))

		if typedN.aliasNames[0] != "" {
			b.WriteString(fmt.Sprintf(" as '%v'", typedN.aliasNames))
		}

		if len(typedN.pipeline.Stages) > 0 {
			b.WriteString(":\n")
			prettyPipeline, err := astutil.PipelineJSON(typedN.pipeline, d+1, true)
			if err != nil { // marshaling as json failed, fall back to Sprintf
				prettyPipeline = astutil.PipelineString(typedN.pipeline, d+1)
			}
			b.Write(prettyPipeline)
		}
	case *OrderByStage:
		b.WriteString("↳ OrderBy(")

		for i, t := range typedN.terms {
			if i != 0 {
				b.WriteString(", ")
			}

			dir := "ASC"
			if !t.ascending {
				dir = "DESC"
			}

			b.WriteString(fmt.Sprintf("%v %v", t.expr.String(), dir))
		}

		b.WriteString("):\n")
		prettyPrint(b, typedN.source, d+1)
	case *ProjectStage:
		b.WriteString("↳ Project(")

		for i, c := range typedN.projectedColumns {
			if i != 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%v", c.Name))
		}

		b.WriteString("):\n")
		prettyPrint(b, typedN.source, d+1)
	case *RowGeneratorStage:
		b.WriteString(fmt.Sprintf("↳ RowGeneratorStage("))
		b.WriteString(typedN.rowCountColumn.Name)
		b.WriteString("):\n")
		prettyPrint(b, typedN.source, d+1)
	case *SetCommand:
		b.WriteString("↳ Set:\n")
		for i, e := range typedN.assignments {
			astutil.PrintTabs(b, d+1)
			b.WriteString(e.String())
			if i != len(typedN.assignments)-1 {
				b.WriteString("\n")
			}
		}
	case *SQLSubqueryExpr:
		b.WriteString("(subquery)")
	case *SubquerySourceStage:
		if typedN.fromCTE {
			b.WriteString("↳ CTESubquery(" + typedN.aliasName + "):\n")
		} else {
			b.WriteString("↳ Subquery(" + typedN.aliasName + "):\n")
		}
		prettyPrint(b, typedN.source, d+1)
	case *UnionStage:
		kind := "distinct"
		if typedN.kind == UnionAll {
			kind = "all"
		}
		b.WriteString(fmt.Sprintf("↳ Union (%s):\n", kind))

		prettyPrint(b, typedN.left, d+1)
		astutil.PrintTabs(b, d+1)

		b.WriteString("\n")
		prettyPrint(b, typedN.right, d+1)
	case *UseCommand:
		b.WriteString(fmt.Sprintf("↳ UseCommand (%s):\n", typedN.db))
	default:
		panic(fmt.Sprintf("unsupported print operator: %T", typedN))
	}
}
