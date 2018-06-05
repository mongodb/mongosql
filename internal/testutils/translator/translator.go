package translator

import (
	"context"
	"fmt"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/catalog"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/sample"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
)

// Translator is a type for translating MySQL queries to MongoDB aggregation
// pipelines based on a schema.Schema and mongodb.Info.
type Translator struct {
	info   *mongodb.Info
	schema *schema.Schema
}

// NewTranslator creates a new Translator by fetching and translating the latest
// schema stored in the sampleSource database.
func NewTranslator(o *config.SchemaSampleOptions, s *mongodb.SessionProvider) (*Translator, error) {

	lgr := log.GlobalLogger()

	session, err := s.AdminSession(context.Background())
	if err != nil {
		return nil, err
	}

	sch, err := sample.ReadSchema(o, session, lgr)
	if err != nil {
		return nil, err
	}

	if sch == nil {
		return nil, fmt.Errorf("no schema found in sampleSource")
	}

	cfg := config.Default()

	info, err := mongodb.LoadInfo(lgr, s, session, sch, cfg)
	if err != nil {
		return nil, err
	}

	return &Translator{
		info:   info,
		schema: sch,
	}, nil
}

// TranslateQuery takes a MySQL query in string form, and translates it into
// an aggregation pipeline.
func (t *Translator) TranslateQuery(dbName, sql string) ([]bson.D, string, error) {
	stmt, err := parser.Parse(sql)
	if err != nil {
		return nil, "", err
	}

	ctx := createConnectionCtx(t.info)

	catalog, err := createCatalog(t.schema, ctx.Variables())
	if err != nil {
		return nil, "", err
	}

	naivePlan, err := evaluator.AlgebrizeQuery(stmt, dbName, ctx.Variables(), catalog)
	if err != nil {
		return nil, "", err
	}

	optimizedPlan := evaluator.OptimizePlan(ctx, naivePlan)

	ms, ok := optimizedPlan.(*evaluator.MongoSourceStage)
	if !ok {
		err = fmt.Errorf("query was not fully pushed down: root plan stage was a %T", optimizedPlan)
		return nil, "", err
	}

	return ms.Pipeline(), ms.Collection(), nil
}

func createCatalog(schema *schema.Schema, vars *variable.Container) (*catalog.Catalog, error) {
	c, err := catalog.Build(schema, vars)
	if err != nil {
		return nil, fmt.Errorf("unable to build catalog: %v", err)
	}
	return c, nil
}
