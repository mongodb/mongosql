package evaluator

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/mongodb/mongo-tools/common/bsonutil"
	"gopkg.in/mgo.v2/bson"
)

// PrettyPrintPlan takes a plan and recursively prints its source.
func PrettyPrintPlan(p PlanStage) string {

	b := bytes.NewBufferString("")

	prettyPrintPlan(b, p, 0)

	return b.String()
}

func prettyPrintPlan(b *bytes.Buffer, p PlanStage, d int) {

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

	switch typedE := p.(type) {
	case *DualStage:
		b.WriteString("↳ Dual")
	case *EmptyStage:
		b.WriteString("↳ Empty")
	case *FilterStage:
		b.WriteString(fmt.Sprintf("↳ Filter (%v):\n", typedE.matcher))

		prettyPrintPlan(b, typedE.source, d+1)
	case *GroupByStage:
		b.WriteString("↳ GroupBy(")
		for i, key := range typedE.keys {
			if i != 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%v", key.String()))
		}
		b.WriteString("):\n")

		prettyPrintPlan(b, typedE.source, d+1)
	case *JoinStage:
		b.WriteString("↳ Join:\n")

		prettyPrintPlan(b, typedE.left, d+1)
		printTabs(b, d+1)

		b.WriteString(fmt.Sprintf("%v\n", typedE.kind))
		prettyPrintPlan(b, typedE.right, d+1)

		if typedE.matcher != nil {
			printTabs(b, d+1)
			b.WriteString(fmt.Sprintf("on %v\n", typedE.matcher.String()))
		}
	case *LimitStage:
		b.WriteString(fmt.Sprintf("↳ Limit(offset: %v, limit: %v):\n", typedE.offset, typedE.limit))

		prettyPrintPlan(b, typedE.source, d+1)
	case *OrderByStage:
		b.WriteString("↳ OrderBy(")

		for i, t := range typedE.terms {
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
		prettyPrintPlan(b, typedE.source, d+1)
	case *ProjectStage:
		b.WriteString("↳ Project(")

		for i, c := range typedE.projectedColumns {
			if i != 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%v", c.Name))
		}

		b.WriteString("):\n")
		prettyPrintPlan(b, typedE.source, d+1)
	case *SchemaDataSourceStage:
		b.WriteString("↳ SchemaDataSource:")
	case *MongoSourceStage:
		b.WriteString(fmt.Sprintf("↳ MongoSource: '%v' (db: '%v', collection: '%v')", typedE.tableName, typedE.dbName, typedE.collectionName))

		if typedE.aliasName != "" {
			b.WriteString(fmt.Sprintf(" as '%v'", typedE.aliasName))
		}

		b.WriteString(":\n")
		prettyPipeline, err := pipelineJSON(typedE.pipeline, d+1)
		if err != nil { // marshaling as json failed, fall back to Sprintf
			prettyPipeline = pipelineString(typedE.pipeline, d+1)
		}
		b.Write(prettyPipeline)
	case *BSONSourceStage:
		b.WriteString("↳ BSONSource:\n")
	default:
		panic(fmt.Sprintf("unsupported print operator: %T", typedE))
	}
}
