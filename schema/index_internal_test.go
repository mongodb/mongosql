package schema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateUniqueNames(t *testing.T) {
	type test struct {
		name               string
		indexes            Indexes
		expectedSQLNames   []string
		expectedMongoNames []string
	}

	runTest := func(test test) {
		t.Run(test.name, func(t *testing.T) {
			req := require.New(t)

			test.indexes.createUniqueNames()
			for i, index := range test.indexes {
				req.Equal(test.expectedSQLNames[i], index.SQLName(), "name SQLName does not match at index %d", i)
				req.Equal(test.expectedMongoNames[i], index.MongoName(), "name MongoName does not match at index %d", i)
			}
		})
	}

	tests := []test{
		{
			name:               "empty",
			indexes:            Indexes{},
			expectedSQLNames:   []string{},
			expectedMongoNames: []string{},
		},
		{
			name:               "one named index",
			indexes:            Indexes{NewIndex("fOo", false, false, []IndexPart{})},
			expectedSQLNames:   []string{"foo"},
			expectedMongoNames: []string{"fOo"},
		},
		{
			name: "one unnamed index",
			indexes: Indexes{NewIndex("", false, false, []IndexPart{
				NewIndexPart("A", 1),
				NewIndexPart("b", -1),
			})},
			expectedSQLNames:   []string{"a_1_b_-1"},
			expectedMongoNames: []string{"a_1_b_-1"},
		},
		{
			name: "two named indexes",
			// Named indexes can have uppercased mongo names, the names may
			// come directly from MongoDB. It is impossible for created index names
			// to have uppercase.
			indexes: Indexes{
				NewIndex("fOo", false, false, []IndexPart{}),
				NewIndex("BaR", false, false, []IndexPart{}),
			},
			expectedSQLNames:   []string{"foo", "bar"},
			expectedMongoNames: []string{"fOo", "BaR"},
		},
		{
			name: "two unnamed indexes",
			indexes: Indexes{
				NewIndex("", false, false, []IndexPart{
					NewIndexPart("a", 1),
					NewIndexPart("B", 1),
				}),
				NewIndex("", false, true, []IndexPart{
					NewIndexPart("A", 1),
					NewIndexPart("b", 1),
				}),
			},
			expectedSQLNames:   []string{"a_1_b_1", "a_text_b_text"},
			expectedMongoNames: []string{"a_1_b_1", "a_text_b_text"},
		},
		{
			name: "collisions",
			indexes: Indexes{
				NewIndex("a_1_b_1", false, false, []IndexPart{}),
				NewIndex("a_text_b_text", false, false, []IndexPart{}),
				NewIndex("", false, false, []IndexPart{
					NewIndexPart("a", 1),
					NewIndexPart("b", 1),
				}),
				NewIndex("", false, true, []IndexPart{
					NewIndexPart("a", 1),
					NewIndexPart("b", 1),
				}),
				// There is actually no way for this index to also exist, it would be an error in mongodb, but this
				// is to stress the naming algorithm.
				NewIndex("", false, true, []IndexPart{
					NewIndexPart("a", 1),
					NewIndexPart("B", 1),
				}),
			},
			expectedSQLNames:   []string{"a_1_b_1", "a_text_b_text", "a_1_b_1_0", "a_text_b_text_0", "a_text_b_text_1"},
			expectedMongoNames: []string{"a_1_b_1", "a_text_b_text", "a_1_b_1_0", "a_text_b_text_0", "a_text_b_text_1"},
		},
	}

	for _, test := range tests {
		runTest(test)
	}
}
