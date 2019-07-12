package translator

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/catalog"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/astutil"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/schema/sample"

	"github.com/pkg/errors"

	"go.mongodb.org/mongo-driver/bson"
)

// Translator is a type for translating MySQL queries to MongoDB aggregation
// pipelines based on a schema.Schema and mongodb.Info.
type Translator struct {
	info   *mongodb.Info
	schema *schema.Schema
}

// NewTranslator creates a new Translator by fetching and translating the latest
// schema stored in the sampleSource database.
func NewTranslator(ctx context.Context, cfg *config.Schema, s *mongodb.SessionProvider) (*Translator, error) {
	lgr := log.GlobalLogger()
	log.SetOutputWriter(ioutil.Discard)

	session, err := s.AuthenticatedAdminSession(context.Background())
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = session.Close()
	}()

	sampleCfg := sample.NewMongosqldConfig(cfg, nil)
	sampler := sample.NewSampler(sampleCfg, lgr, s)
	sch, err := sampler.Sample(ctx)
	if err != nil {
		return nil, err
	}

	if sch == nil {
		return nil, fmt.Errorf("no schema found in sampleSource")
	}

	defaultCfg := config.Default()
	info, err := mongodb.LoadInfo(ctx, lgr, s, session, sch, defaultCfg)
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
func (t *Translator) TranslateQuery(ctx context.Context, dbName, sql string) ([]bson.D, string, error) {
	stmt, err := parser.Parse(sql)
	if err != nil {
		return nil, "", err
	}

	lg := log.GlobalLogger()
	vars := createVariables(t.info)
	catalog, err := createCatalog(t.schema, vars, t.info)
	if err != nil {
		return nil, "", err
	}

	algebrizerCfg := evaluator.NewAlgebrizerConfig(lg, dbName, catalog)

	naivePlan, err := evaluator.AlgebrizeQuery(algebrizerCfg, stmt)
	if err != nil {
		return nil, "", err
	}

	pushdownCfg := evaluator.NewPushdownConfig(lg, vars)
	optimizerCfg := evaluator.NewOptimizerConfig(lg, vars)
	optimizedPlan, err := evaluator.OptimizePlan(ctx, optimizerCfg, naivePlan)
	if err != nil && !evaluator.IsNonFatalPushdownError(err) {
		return nil, "", err
	}

	translated, err := evaluator.PushdownPlan(ctx, pushdownCfg, optimizedPlan)
	if err != nil && !evaluator.IsNonFatalPushdownError(err) {
		return nil, "", err
	}

	ms, ok := translated.(*evaluator.MongoSourceStage)
	if !ok {
		err = fmt.Errorf("query was not fully pushed down: root plan stage was a %T", optimizedPlan)
		return nil, "", err
	}

	bsonPipeline, err := astutil.DeparsePipeline(ms.Pipeline())
	if err != nil {
		return nil, "", errors.Wrap(err, "failed to deparse ast.Pipeline into []bson.D")
	}

	return bsonPipeline, ms.Collection(), nil
}

func createVariables(info *mongodb.Info) catalog.VariableContainer {
	gbl := variable.NewGlobalContainer(nil)
	version := values.NewSQLVarchar(values.VariableSQLValueKind, info.Version)
	gbl.SetSystemVariable(variable.MongoDBVersion, version)

	ctn := variable.NewSessionContainer(gbl)
	ctn.SetSystemVariable(variable.MongoDBVersion, version)
	return ctn
}

func createCatalog(schema *schema.Schema, vars catalog.VariableContainer, info *mongodb.Info) (*catalog.SQLCatalog, error) {
	c, err := catalog.Build(schema, vars, info)
	if err != nil {
		return nil, fmt.Errorf("unable to build catalog: %v", err)
	}
	return c, nil
}
