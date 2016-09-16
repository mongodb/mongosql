package evaluator

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/mongodb/mongo-tools/common/bsonutil"
	"gopkg.in/mgo.v2/bson"
)

// PrettyPrintCommand takes a command and prints it out.
func PrettyPrintCommand(c command) string {
	return prettyPrintNode(c)
}

// PrettyPrintPlan takes a plan and recursively prints its source.
func PrettyPrintPlan(p PlanStage) string {
	return prettyPrintNode(p)
}

func prettyPrintNode(n node) string {
	b := bytes.NewBufferString("")

	prettyPrint(b, n, 0)

	return b.String()
}

func prettyPrint(b *bytes.Buffer, n node, d int) {

	printTabs := func(b *bytes.Buffer, d int) {
		for i := 0; i < d; i++ {
			b.WriteString("\t")
		}
	}

	pipelineJSON := func(stages []bson.D, depth int) ([]byte, error) {
		buf := bytes.Buffer{}

		for i, s := range stages {
			converted, err := bsonutil.GetBSONValueAsJSON(s)
			if err != nil {
				return nil, err
			}
			b, err := json.Marshal(converted)
			if err != nil {
				return nil, err
			}
			printTabs(&buf, depth)
			buf.Write(b)
			if i != len(stages)-1 {
				buf.WriteString(",\n")
			}
		}
		return buf.Bytes(), nil
	}

	pipelineString := func(stages []bson.D, depth int) []byte {
		buf := bytes.Buffer{}
		for i, stage := range stages {
			printTabs(&buf, depth)
			buf.WriteString(fmt.Sprintf("  stage %v: '%v'\n", i+1, stage))
		}
		return buf.Bytes()
	}

	printTabs(b, d)

	switch typedN := n.(type) {
	case *BSONSourceStage:
		b.WriteString("↳ BSONSource:\n")
	case *DualStage:
		b.WriteString("↳ Dual")
	case *CacheStage:
		b.WriteString("↳ Cache\n")
		prettyPrint(b, typedN.source, d+1)
	case *EmptyStage:
		b.WriteString("↳ Empty")
	case *FilterStage:
		b.WriteString(fmt.Sprintf("↳ Filter (%v):\n", typedN.matcher))

		prettyPrint(b, typedN.source, d+1)
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
		printTabs(b, d+1)

		b.WriteString(fmt.Sprintf("%v\n", typedN.kind))
		prettyPrint(b, typedN.right, d+1)

		if typedN.matcher != nil {
			printTabs(b, d+1)
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
		b.WriteString(fmt.Sprintf("↳ MongoSource: '%v' (db: '%v', collection: '%v')", typedN.tableNames, typedN.dbName, typedN.collectionNames))

		if typedN.aliasNames[0] != "" {
			b.WriteString(fmt.Sprintf(" as '%v'", typedN.aliasNames))
		}

		if len(typedN.pipeline) > 0 {
			b.WriteString(":\n")
			prettyPipeline, err := pipelineJSON(typedN.pipeline, d+1)
			if err != nil { // marshaling as json failed, fall back to Sprintf
				prettyPipeline = pipelineString(typedN.pipeline, d+1)
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
	case *SchemaDataSourceStage:
		b.WriteString("↳ SchemaDataSource:")
	case *SetCommand:
		b.WriteString("↳ Set:\n")
		for _, e := range typedN.assignments {
			printTabs(b, d+1)
			b.WriteString(e.String())
			b.WriteString("\n")
		}
	case *SubquerySourceStage:
		b.WriteString("↳ Subquery(" + typedN.aliasName + "):\n")
		prettyPrint(b, typedN.source, d+1)
	default:
		panic(fmt.Sprintf("unsupported print operator: %T", typedN))
	}
}
