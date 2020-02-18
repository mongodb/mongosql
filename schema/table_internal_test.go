package schema

import (
	"testing"

	"github.com/10gen/sqlproxy/log"
	"github.com/stretchr/testify/require"

	"go.mongodb.org/mongo-driver/bson"
)

func TestTable_PostProcess(t *testing.T) {
	// This is a regression test for BI-2464. That fixed a bug where
	// we did not deep copy a parent's pipeline during post-processing.
	// Because we did not deep copy, we were vulnerable to overwriting
	// sibling tables' pipelines. Specifically, consider the following
	// scenario that existed pre-BI-2464:
	//
	//   A collection, c, with two array fields, a1 and a2.  This maps
	//   to three Tables: c, c_a1, and c_a2. During post-processing, a
	//   table's parent's pipeline is prepended to its own table. If
	//
	//     len(c.pipeline) + len(c_a1.pipeline) <= cap(c.pipeline)
	//
	//   then no new memory is allocated for c_a1.pipeline when c_a1
	//   is post-processed.  The array that supports c.pipeline will
	//   have c_a1.pipeline's data written to it,  starting at index
	//   len(c.pipeline).
	//
	//   If the same holds true for c_a2, as in if
	//
	//     len(c.pipeline) + len(c_a2.pipeline) <= cap(c.pipeline)
	//
	//   then no new memory is allocated for c_a2.pipeline when c_a2
	//   is post-processed.  The array that supports c.pipeline will
	//   have c_a2.pipeline's data written to it,  starting at index
	//   len(c.pipeline).
	//
	//   This is problematic! Notice that we just overwrote the array
	//   that underlies c.pipeline's slice with c_a2's pipeline data.
	//   We also did that in the previous step when we post-processed
	//   c_a1. This means that c_a1's pipeline is overwritten with
	//   c_a2's pipeline!
	//
	// That type of scenario is tested here.

	req := require.New(t)

	parentStage := bson.D{{Key: "parent", Value: 1}}

	parentPipeline := make([]bson.D, 1, 2)
	parentPipeline[0] = parentStage

	parent := &Table{
		pipeline: parentPipeline,
	}

	child1Stage := bson.D{{Key: "c1", Value: "foo"}}
	child2Stage := bson.D{{Key: "c2", Value: "bar"}}

	child1Pipeline := []bson.D{child1Stage}
	child2Pipeline := []bson.D{child2Stage}

	child1 := &Table{
		pipeline: child1Pipeline,
		parent:   parent,
	}
	child2 := &Table{
		pipeline: child2Pipeline,
		parent:   parent,
	}

	// PostProcess child1
	child1.PostProcess(log.NoOpLogger(), false)

	// Ensure parent's pipeline is not changed.
	req.Len(parent.pipeline, 1, "len(parent.pipeline)")
	req.Equal(parentStage, parent.pipeline[0], "parent.pipeline[0]")

	// Ensure child1's pipeline is prepended with parent's pipeline.
	req.Len(child1.pipeline, 2, "len(child1.pipeline)")
	req.Equal(parentStage, child1.pipeline[0], "child1.pipeline[0]")
	req.Equal(child1Stage, child1.pipeline[1], "child1.pipeline[1]")

	// PostProcess child2
	child2.PostProcess(log.NoOpLogger(), false)

	// Ensure parent's pipeline is not changed.
	req.Len(parent.pipeline, 1, "len(parent.pipeline)")
	req.Equal(parentStage, parent.pipeline[0], "parent.pipeline[0]")

	// Ensure child2's pipeline is prepended with parent's pipeline.
	req.Len(child2.pipeline, 2, "len(child2.pipeline)")
	req.Equal(parentStage, child2.pipeline[0], "child2.pipeline[0]")
	req.Equal(child2Stage, child2.pipeline[1], "child2.pipeline[1]")

	// Ensure child1's pipeline is prepended with parent's pipeline.
	req.Len(child1.pipeline, 2, "len(child1.pipeline)")
	req.Equal(parentStage, child1.pipeline[0], "child1.pipeline[0]")
	req.Equal(child1Stage, child1.pipeline[1], "child1.pipeline[1]")
}
