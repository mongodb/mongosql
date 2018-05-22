package evaluator

import (
	"testing"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/catalog"
	"github.com/stretchr/testify/require"
)

// Tests for any private functions in the evaluator package.

// TestGetFastPlanStageTest tests the functionality of getFastPlanStage.
func TestGetFastPlanStageTest(t *testing.T) {
	req := require.New(t)
	mongoSourceStage := &MongoSourceStage{
		selectIDs:  []int{0},
		dbName:     "foo",
		tableNames: []string{"bar"},
		aliasNames: []string{"biz"},
		tableType:  catalog.TableType("view"),
	}

	fastPlan, ok := getFastPlanStage(mongoSourceStage)
	req.NotNil(fastPlan, "fastPlan should not be nil")
	req.True(ok, "ok should be true")

	optimizableUnion := &ProjectStage{
		source: &UnionStage{
			left:  mongoSourceStage,
			right: mongoSourceStage,
			kind:  UnionAll,
		},
	}

	fastPlan, ok = getFastPlanStage(optimizableUnion)
	req.NotNil(fastPlan, "fastPlan should not be nil")
	req.True(ok, "ok should be true")

	notOptimizableUnion := &ProjectStage{
		source: &UnionStage{
			left:  mongoSourceStage,
			right: mongoSourceStage,
			kind:  UnionDistinct,
		},
	}

	fastPlan, ok = getFastPlanStage(notOptimizableUnion)
	req.Nil(fastPlan, "fastPlan should be nil")
	req.False(ok, "ok should be false")

	optimizableMetaUnion1 := &ProjectStage{
		source: &UnionStage{
			left:  optimizableUnion,
			right: mongoSourceStage,
			kind:  UnionAll,
		},
	}

	fastPlan, ok = getFastPlanStage(optimizableMetaUnion1)
	req.NotNil(fastPlan, "fastPlan should not be nil")
	req.True(ok, "ok should be true")

	optimizableMetaUnion2 := &ProjectStage{
		source: &UnionStage{
			left:  optimizableMetaUnion1,
			right: mongoSourceStage,
			kind:  UnionAll,
		},
	}

	fastPlan, ok = getFastPlanStage(optimizableMetaUnion2)
	req.NotNil(fastPlan, "fastPlan should not be nil")
	req.True(ok, "ok should be true")

	notOptimizableMetaUnion1 := &ProjectStage{
		source: &UnionStage{
			left:  notOptimizableUnion,
			right: mongoSourceStage,
			kind:  UnionAll,
		},
	}

	fastPlan, ok = getFastPlanStage(notOptimizableMetaUnion1)
	req.Nil(fastPlan, "fastPlan should be nil")
	req.False(ok, "ok should be false")

	notOptimizableMetaUnion2 := &ProjectStage{
		source: &UnionStage{
			left:  notOptimizableUnion,
			right: optimizableUnion,
			kind:  UnionAll,
		},
	}

	fastPlan, ok = getFastPlanStage(notOptimizableMetaUnion2)
	req.Nil(fastPlan, "fastPlan should be nil")
	req.False(ok, "ok should be false")

	notOptimizableMetaUnion3 := &ProjectStage{
		source: &UnionStage{
			left:  optimizableMetaUnion1,
			right: optimizableUnion,
			kind:  UnionDistinct,
		},
	}

	fastPlan, ok = getFastPlanStage(notOptimizableMetaUnion3)
	req.Nil(fastPlan, "fastPlan should be nil")
	req.False(ok, "ok should be false")
}

// TestEnsureFastPlanProjectInvariant tests the
// functionality of ensureFastPlanProjectInvariant.
func TestEnsureFastPlanProjectInvariant(t *testing.T) {
	req := require.New(t)
	mongoSourceStage := &MongoSourceStage{
		selectIDs:  []int{0},
		dbName:     "foo",
		tableNames: []string{"bar"},
		aliasNames: []string{"biz"},
		tableType:  catalog.TableType("view"),
		pipeline: []bson.D{
			{{Name: "$project", Value: bson.D{{Name: "foo", Value: 1}}}},
		},
	}

	fastPlan, ok := getFastPlanStage(mongoSourceStage)
	req.NotNil(fastPlan, "fastPlan should not be nil")
	req.True(ok, "ok should be true")
	ensureFastPlanProjectInvariant(fastPlan)
	ms, ok := fastPlan.(*MongoSourceStage)
	req.True(ok, "ok should be true")
	req.Equal(0, ms.pipeline[0][0].Value.(bson.D).Map()["_id"],
		"_id:0 must be added to project")

	mongoSourceStage1 := &MongoSourceStage{
		selectIDs:  []int{0},
		dbName:     "foo",
		tableNames: []string{"bar"},
		aliasNames: []string{"biz"},
		tableType:  catalog.TableType("view"),
		pipeline: []bson.D{
			{{Name: "$project", Value: bson.D{{Name: "foo", Value: 1}}}},
		},
	}
	mongoSourceStage2 := &MongoSourceStage{
		selectIDs:  []int{0},
		dbName:     "foo",
		tableNames: []string{"bar"},
		aliasNames: []string{"biz"},
		tableType:  catalog.TableType("view"),
		pipeline: []bson.D{
			{{Name: "$project", Value: bson.D{{Name: "bar", Value: 2}}}},
		},
	}
	optimizableUnion := &ProjectStage{
		source: &UnionStage{
			left:  mongoSourceStage1,
			right: mongoSourceStage2,
			kind:  UnionAll,
		},
	}

	fastPlan, ok = getFastPlanStage(optimizableUnion)
	req.NotNil(fastPlan, "fastPlan should not be nil")
	req.True(ok, "ok should be true")
	ensureFastPlanProjectInvariant(fastPlan)
	us, ok := fastPlan.(*UnionStage)
	req.True(ok, "ok should be true")
	ms1, ok := us.left.(*MongoSourceStage)
	req.True(ok, "ok should be true")
	req.Equal(0, ms1.pipeline[0][0].Value.(bson.D).Map()["_id"],
		"_id:0 must be added to left stage project")
	ms2, ok := us.right.(*MongoSourceStage)
	req.True(ok, "ok should be true")
	req.Equal(0, ms2.pipeline[0][0].Value.(bson.D).Map()["_id"],
		"_id:0 must be added to right stage project")
}
