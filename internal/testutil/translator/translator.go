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
func NewTranslator(ctx context.Context, cfg *config.Schema, sp *mongodb.SessionProvider) (*Translator, error) {
	lgr := log.GlobalLogger()
	log.SetOutputWriter(ioutil.Discard)

	session, err := sp.AuthenticatedAdminSession()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = session.Close()
	}()

	sampleCfg := sample.NewMongosqldConfig(cfg, nil)
	sampler := sample.NewSampler(sampleCfg, lgr, sp)
	sch, err := sampler.Sample(ctx)
	if err != nil {
		return nil, err
	}

	if sch == nil {
		return nil, fmt.Errorf("no schema found in sampleSource")
	}

	defaultCfg := config.Default()
	info, err := mongodb.LoadInfo(ctx, lgr, sp, session, sch, defaultCfg)
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
	ctlg, err := createCatalog(t.schema, vars, t.info)
	if err != nil {
		return nil, "", err
	}

	// algebrizer settings
	mongoDBToplogy := vars.GetString(variable.MongoDBTopology)
	sqlValueKind := evaluator.GetSQLValueKind(vars)
	sqlSelectLimit := vars.GetUint64(variable.SQLSelectLimit)
	mongoDBMaxVarcharLength := vars.GetUint64(variable.MongoDBMaxVarcharLength)
	groupConcatMaxLen := vars.GetInt64(variable.GroupConcatMaxLen)
	polymorphicTypeConversionMode := vars.GetString(variable.PolymorphicTypeConversionMode)
	mdbVersion := evaluator.GetMongoDBVersion(vars)

	// pushdown settings
	shouldPushDown := vars.GetBool(variable.Pushdown)
	pushDownSelfJoins := vars.GetBool(variable.OptimizeSelfJoins)
	format := evaluator.NoOutputFormat
	formatVersion := evaluator.NoOutputVersion

	// optimizer settings
	collation := vars.GetCollation(variable.CollationConnection)
	optimizeCrossJoins := vars.GetBool(variable.OptimizeCrossJoins)
	optimizeEvaluations := vars.GetBool(variable.OptimizeEvaluations)
	optimizeFiltering := vars.GetBool(variable.OptimizeFiltering)
	optimizeInnerJoins := vars.GetBool(variable.OptimizeInnerJoins)
	reconcileArithmeticAggFunctions := vars.GetBool(variable.ReconcileArithmeticAggFunctions)

	// For now, assume no --writeMode.
	algebrizerCfg := evaluator.NewAlgebrizerConfig(lg, dbName, ctlg, vars, mongoDBToplogy, false,
		sqlValueKind, sqlSelectLimit, mongoDBMaxVarcharLength, groupConcatMaxLen,
		polymorphicTypeConversionMode, mdbVersion, true, false)

	naivePlan, err := evaluator.AlgebrizeQuery(ctx, algebrizerCfg, stmt)
	if err != nil {
		return nil, "", err
	}

	pushdownCfg := evaluator.NewPushdownConfig(lg, mdbVersion, false, false, true, shouldPushDown,
		pushDownSelfJoins, sqlValueKind, format, formatVersion)
	optimizerCfg := evaluator.NewOptimizerConfig(lg, collation, sqlValueKind, optimizeCrossJoins,
		optimizeEvaluations, optimizeFiltering, optimizeInnerJoins, reconcileArithmeticAggFunctions)

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

func createVariables(info *mongodb.Info) *variable.Container {
	gbl := variable.NewGlobalContainer(nil)
	version := values.NewSQLVarchar(values.VariableSQLValueKind, info.Version)
	gbl.SetSystemVariable(variable.MongoDBVersion, version)

	ctn := variable.NewSessionContainer(gbl)
	ctn.SetSystemVariable(variable.MongoDBVersion, version)
	return ctn
}

func createCatalog(schema *schema.Schema, vars *variable.Container, info *mongodb.Info) (*catalog.SQLCatalog, error) {
	c, err := catalog.Build(schema, vars, info, false)
	if err != nil {
		return nil, fmt.Errorf("unable to build catalog: %v", err)
	}
	return c, nil
}
