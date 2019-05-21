package optimizer_test

import (
	"context"
	"testing"

	"github.com/10gen/mongoast/internal/testutil"
	. "github.com/10gen/mongoast/optimizer"
	"github.com/10gen/mongoast/parser"
)

func TestCompressProjectionStagesDown(t *testing.T) {
	testCases := testutil.LoadTestCases("projection_stage_compression_down.json")

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			in, err := parser.ParsePipeline(test.Input)
			if err != nil {
				t.Fatalf("Failed to parse input pipeline: %v", err)
			}

			compressed := RunPasses(context.Background(),
				in,
				NormalizePipeline,
				ProjectionStageCompressionDown,
				NormalizePipeline,
			)
			actualStr := parser.DeparsePipeline(compressed).String()

			expectedPipeline, err := parser.ParsePipeline(test.Expected)
			if err != nil {
				t.Fatalf("Failed to parse expected pipeline: %v", err)
			}

			expectedStr := parser.DeparsePipeline(expectedPipeline).String()
			if actualStr != expectedStr {
				t.Errorf("Compressed pipeline does not match expected. Got %v, want %v.", actualStr, expectedStr)
			}

		})
	}
}

func TestCompressProjectionStagesUp(t *testing.T) {
	testCases := testutil.LoadTestCases("projection_stage_compression_up.json")

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			in, err := parser.ParsePipeline(test.Input)
			if err != nil {
				t.Fatalf("Failed to parse input pipeline: %v", err)
			}

			compressed := RunPasses(context.Background(),
				in,
				NormalizePipeline,
				ProjectionStageCompressionUp,
				NormalizePipeline,
			)
			actualStr := parser.DeparsePipeline(compressed).String()

			expectedPipeline, err := parser.ParsePipeline(test.Expected)
			if err != nil {
				t.Fatalf("Failed to parse expected pipeline: %v", err)
			}

			expectedStr := parser.DeparsePipeline(expectedPipeline).String()
			if actualStr != expectedStr {
				t.Errorf("Compressed pipeline does not match expected. Got %v, want %v.", actualStr, expectedStr)
			}

		})
	}
}
