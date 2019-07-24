//+build integration

package sample

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/bsonutil"
	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/internal/option"
	"github.com/10gen/sqlproxy/internal/strutil"
	"github.com/10gen/sqlproxy/internal/testutil/dbutils"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/stretchr/testify/require"
)

const (
	db1, db2, db3 = "sampleTest1", "sampleTest2", "sampleTest3"
	c1, c2        = "c1", "c2"
)

func init() {
	cfg.Schema.Stored.Source = "sampleStore"
	cfg.Schema.Sample.Namespaces = []string{
		"sampleTest*.*", "sampleStore.*",
	}
}

func TestFetchNamespaces(t *testing.T) {
	provider, err := mongodb.NewSqldSessionProvider(cfg)
	if err != nil {
		t.Fatalf("failed to set up session provider to test server: %v", err)
	}

	session, err := provider.Session(context.Background())
	if err != nil {
		t.Fatalf("failed to set up session to test server: %v", err)
	}
	defer session.Close()

	matcher, err := strutil.NewMatcher([]string{"*.*"})
	if err != nil {
		t.Fatal(err)
	}

	req := require.New(t)

	cleanupData(session)
	defer cleanupData(session)

	doc := bsonutil.NewDArray(bsonutil.NewD())

	dbutils.InsertDocuments(session, db1, c1, doc)
	dbutils.InsertDocuments(session, db2, c2, doc)
	dbutils.InsertDocuments(session, db2, c1, doc)

	mappings, err := fetchNamespaces(context.Background(), session, lgr, matcher)
	req.Nil(err, "error fetching namespaces")

	req.Equal(len(mappings[db1]), 1)
	req.Equal(mappings[db1][0], c1)
	req.Equal(len(mappings[db2]), 2)

	dbutils.DropDatabase(session, db2)
	mappings, err = fetchNamespaces(context.Background(), session, lgr, matcher)
	req.Nil(err, "error fetching namespaces")
	_, found := mappings[db1]

	errFound := "found unexpected database"
	errMissing := "could not find expected database"
	req.True(found, errMissing)

	req.Equal(len(mappings[db1]), 1)
	req.Equal(mappings[db1][0], c1)
	_, found = mappings[db2]
	req.False(found, errFound)
	_, found = mappings["admin"]
	req.False(found, errFound)
	_, found = mappings["config"]
	req.False(found, errFound)
	_, found = mappings["local"]
	req.False(found, errFound)
	_, found = mappings["system"]
	req.False(found, errFound)
}

func TestGetViewPipelinesInDatabase(t *testing.T) {
	provider, err := mongodb.NewSqldSessionProvider(cfg)
	if err != nil {
		t.Fatalf("failed to set up session provider to test server: %v", err)
	}

	session, err := provider.Session(context.Background())
	if err != nil {
		t.Fatalf("failed to set up session to test server: %v", err)
	}

	defer session.Close()
	cleanupData(session)
	defer cleanupData(session)

	baseCollection, view1, view2 := "baseCollection", "view-1-on-collection", "view-2-on-collection"

	view1Ns := formatNamespace(db1, view1, false)
	view2Ns := formatNamespace(db1, view2, false)
	baseCollectionNs := formatNamespace(db1, baseCollection, false)

	pipeline1 := bsonutil.NewD(
		bsonutil.NewDocElem("$group", bsonutil.NewD(
			bsonutil.NewDocElem("_id", bsonutil.NewD()),
			bsonutil.NewDocElem("b", bsonutil.NewD(
				bsonutil.NewDocElem("$sum", int32(1)),
			)),
		)),
	)

	pipeline2 := bsonutil.NewD(
		bsonutil.NewDocElem("$addFields", bsonutil.NewD(
			bsonutil.NewDocElem("c", int32(1)),
		)),
	)

	req := require.New(t)

	err = createView(session, db1, baseCollection, view1, bsonutil.NewDArray(pipeline1))
	req.NoError(err, "failed to create view 1")

	err = createView(session, db1, view1, view2, bsonutil.NewDArray(pipeline2))
	req.NoError(err, "failed to create view 2")

	pipelines, err := GetViewPipelinesInDatabase(context.Background(), session, db1)
	req.NoError(err, "failed to get views in pipeline")

	req.Equal("", pipelines[baseCollectionNs].Collection,
		"found base collection in view pipelines map")
	req.Nil(pipelines[baseCollectionNs].Pipeline,
		"found base collection pipeline in view pipelines map")

	req.Equal(bsonutil.NewDArray(pipeline1), pipelines[view1Ns].Pipeline, "view1 pipeline does not match")
	req.Equal(pipelines[view1Ns].Collection, baseCollection,
		"base collection for view 1 does not match")

	req.Equal(bsonutil.NewDArray(pipeline1, pipeline2), pipelines[view2Ns].Pipeline,
		"view2 pipeline does not match")
	req.Equal(pipelines[view2Ns].Collection, baseCollection,
		"base collection for view 2 does not match")

}

func TestSample(t *testing.T) {
	provider, err := mongodb.NewSqldSessionProvider(cfg)
	if err != nil {
		t.Fatalf("failed to set up session provider to test server: %v", err)
	}

	session, err := provider.Session(context.Background())
	if err != nil {
		t.Fatalf("failed to set up session to test server: %v", err)
	}
	defer session.Close()

	req := require.New(t)

	cleanupData(session)
	defer cleanupData(session)

	doc := bsonutil.NewDArray(bsonutil.NewD(bsonutil.NewDocElem("a", bsonutil.NewM())))

	dbutils.InsertDocuments(session, db1, c1, doc)
	dbutils.InsertDocuments(session, db2, c2, doc)
	dbutils.InsertDocuments(session, db2, c1, doc)
	dbutils.InsertDocuments(session, cfg.Schema.Stored.Source, c1, doc)

	// enabling profiling should introduce an additional system.profile
	// collection which should not be sampled
	dbutils.RunCmd(session, db2, bsonutil.NewD(bsonutil.NewDocElem("profile", 1)), &struct{}{})

	opts := NewMongosqldConfig(&cfg.Schema, nil)
	sampler := NewSampler(opts, lgr, getSessionProvider(req))
	sampleSchema, err := sampler.Sample(context.Background())

	req.Nilf(err, "did not expect error in sampling")
	req.NotNilf(sampleSchema, "did not expect sample schema to be nil")
	dbutils.RunCmd(session, db2, bsonutil.NewD(bsonutil.NewDocElem("profile", 0)), &struct{}{})

	req.NotZero(countTables(sampleSchema), "found no sampled namespaces")

	errMsg := "whitelisted namespaces should be present"
	req.True(containsNS(sampleSchema, formatNamespace(db1, c1, false)), errMsg)
	req.True(containsNS(sampleSchema, formatNamespace(db2, c2, false)), errMsg)
	req.True(containsNS(sampleSchema, formatNamespace(db2, c1, false)), errMsg)
	req.True(containsNS(sampleSchema, formatNamespace(cfg.Schema.Stored.Source, c1, false)), errMsg)

	errMsg = "non-existent namespaces should not be present"
	req.Nil(sampleSchema.Database("admin"), errMsg)
	req.Nil(sampleSchema.Database("config"), errMsg)
	req.Nil(sampleSchema.Database("local"), errMsg)
	req.Nil(sampleSchema.Database("system"), errMsg)

	errMsg = "non-existent namespaces should not be present"
	req.False(containsNS(sampleSchema, formatNamespace(db1, c2, false)), errMsg)
	req.False(containsNS(sampleSchema, formatNamespace(db3, c2, false)), errMsg)
	req.False(containsNS(sampleSchema, formatNamespace(db3, c1, false)), errMsg)
	req.False(containsNS(sampleSchema, formatNamespace(db2, "system.profile", false)), errMsg)

	errMsg = "special sampling namespaces should not be present"
	req.False(containsNS(sampleSchema, formatNamespace(cfg.Schema.Stored.Source, "schemas", false)), errMsg)
	req.False(containsNS(sampleSchema, formatNamespace(cfg.Schema.Stored.Source, "names", false)), errMsg)
}

func TestNamespaceSelectors(t *testing.T) {
	provider, err := mongodb.NewSqldSessionProvider(cfg)
	if err != nil {
		t.Fatalf("failed to set up session provider to test server: %v", err)
	}

	session, err := provider.Session(context.Background())
	if err != nil {
		t.Fatalf("failed to set up session to test server: %v", err)
	}
	defer session.Close()

	sampleTestDatabases := []string{"bic_test", "bic_blackbox", "bic_functions_test", "persist_test"}
	databaseNamespaces := map[string][]string{
		sampleTestDatabases[0]: {"eleanor", "bar", "hello"},
		sampleTestDatabases[1]: {"bob", "alice", "joe"},
		sampleTestDatabases[2]: {"eleanor", "joe", "bobby"},
	}

	// ideally, we'd delete all databases in the target mongod cluster
	// but this is undesirable when running the bic locally so we
	// enumerate what collections we know exist, instead.
	cleanupData(session, sampleTestDatabases...)
	defer cleanupData(session, sampleTestDatabases...)

	doc := bsonutil.NewDArray(bsonutil.NewD(bsonutil.NewDocElem("a", bsonutil.NewM())))
	for db, collections := range databaseNamespaces {
		for _, collection := range collections {
			dbutils.InsertDocuments(session, db, collection, doc)
		}
	}

	namespaceSelectorTests := []struct {
		description        string
		samplePattern      []string
		expectedNamespaces []string
	}{
		{"inclusion",
			[]string{"*.*"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar",
				"bic_blackbox.joe", "bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.eleanor", "bic_functions_test.bobby", "bic_functions_test.joe",
			},
		},
		{"db_inclusion",
			[]string{"bic_test.*"},
			[]string{"bic_test.eleanor", "bic_test.hello", "bic_test.bar"},
		},
		{"combined_db_inclusion",
			[]string{"bic_test.*", "bic_blackbox.*"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar",
				"bic_blackbox.joe", "bic_blackbox.alice", "bic_blackbox.bob",
			},
		},
		{"db_exclusion",
			[]string{"~bic_test.*"},
			[]string{
				"bic_blackbox.joe", "bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.eleanor", "bic_functions_test.bobby", "bic_functions_test.joe",
			},
		},
		{"combined_db_exclusion",
			[]string{"~bic_test.*", "~bic_blackbox.*"},
			[]string{
				"bic_functions_test.eleanor", "bic_functions_test.bobby", "bic_functions_test.joe",
			},
		},
		{"db_inclusion_and_db_exclusion",
			[]string{"bic_blackbox.*", "~bic_test.*"},
			[]string{
				"bic_blackbox.joe", "bic_blackbox.alice", "bic_blackbox.bob",
			},
		},
		{"combined_db_inclusion_and_db_exclusion",
			[]string{"bic_blackbox.*", "bic_functions_test.*", "~bic_test.*"},
			[]string{
				"bic_blackbox.joe", "bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.eleanor", "bic_functions_test.bobby", "bic_functions_test.joe",
			},
		},
		{"db_inclusion_and_combined_db_exclusion",
			[]string{"bic_blackbox.*", "~bic_test.*", "~bic_functions_test.*"},
			[]string{
				"bic_blackbox.joe", "bic_blackbox.alice", "bic_blackbox.bob",
			},
		},
		{"collection_inclusion",
			[]string{"*.joe"},
			[]string{"bic_functions_test.joe", "bic_blackbox.joe"},
		},
		{"combined_collection_inclusion",
			[]string{"*.joe", "*.eleanor"},
			[]string{"bic_functions_test.joe", "bic_blackbox.joe", "bic_test.eleanor",
				"bic_functions_test.eleanor",
			},
		},
		{"collection_exclusion",
			[]string{"~*.joe"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar",
				"bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.eleanor", "bic_functions_test.bobby",
			},
		},
		{"combined_collection_exclusion",
			[]string{"~*.joe", "~*.hello", "~*.eleanor"},
			[]string{
				"bic_test.bar", "bic_functions_test.bobby",
				"bic_blackbox.alice", "bic_blackbox.bob",
			},
		},
		{"collection_inclusion_and_collection_exclusion",
			[]string{"*.hello", "~*.joe"},
			[]string{"bic_test.hello"},
		},
		{"combined_collection_inclusion_and_collection_exclusion",
			[]string{"~*.joe", "*.hello", "*.bob*"},
			[]string{"bic_test.hello", "bic_blackbox.bob", "bic_functions_test.bobby"},
		},
		{"collection_inclusion_and_combined_collection_exclusion",
			[]string{"*.hello", "~*.joe", "~*.eleanor"},
			[]string{"bic_test.hello"},
		},
		{"collection_inclusion_and_db_inclusion",
			[]string{"bic_blackbox.joe", "bic_test.*"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar", "bic_blackbox.joe",
			},
		},
		{"combined_collection_inclusion_and_db_inclusion",
			[]string{"bic_blackbox.joe", "bic_blackbox.alice", "bic_test.*"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar",
				"bic_blackbox.joe", "bic_blackbox.alice",
			},
		},
		{"collection_inclusion_and_combined_db_inclusion",
			[]string{"bic_blackbox.joe", "bic_test.*", "bic_functions_test.*"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar", "bic_blackbox.joe",
				"bic_functions_test.eleanor", "bic_functions_test.bobby", "bic_functions_test.joe",
			},
		},
		{"collection_inclusion_and_db_exclusion",
			[]string{"bic_blackbox.joe", "~bic_test.*"},
			[]string{
				"bic_blackbox.joe",
			},
		},
		{"combined_collection_inclusion_and_db_exclusion",
			[]string{"bic_blackbox.joe", "bic_blackbox.alice", "~bic_test.*"},
			[]string{
				"bic_blackbox.joe", "bic_blackbox.alice",
			},
		},
		{"collection_inclusion_and_combined_db_exclusion",
			[]string{"bic_blackbox.joe", "~bic_test.*", "~bic_functions_test.*"},
			[]string{
				"bic_blackbox.joe",
			},
		},
		{"ns_inclusion",
			[]string{"bic_blackbox.joe"},
			[]string{"bic_blackbox.joe"},
		},
		{"combined_ns_inclusion",
			[]string{"bic_blackbox.joe", "bic_functions_test.joe"},
			[]string{"bic_blackbox.joe", "bic_functions_test.joe"},
		},
		{"ns_exclusion",
			[]string{"~bic_blackbox.joe"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar",
				"bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.eleanor", "bic_functions_test.bobby", "bic_functions_test.joe",
			},
		},
		{"combined_ns_exclusion",
			[]string{"~bic_blackbox.joe", "~bic_functions_test.bobby"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar",
				"bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.eleanor", "bic_functions_test.joe",
			},
		},
		{"ns_inclusion_and_ns_exclusion",
			[]string{"bic_test.bar", "~bic_blackbox.joe"},
			[]string{"bic_test.bar"},
		},
		{"combined_ns_inclusion_and_ns_exclusion",
			[]string{"bic_test.bar", "bic_blackbox.alice", "~bic_blackbox.joe"},
			[]string{"bic_test.bar", "bic_blackbox.alice"},
		},
		{"ns_inclusion_and_combined_ns_exclusion",
			[]string{"bic_test.bar", "~bic_blackbox.joe", "~bic_functions_test.joe"},
			[]string{"bic_test.bar"},
		},
		{"ns_inclusion_and_db_inclusion",
			[]string{"bic_blackbox.joe", "bic_test.*"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar",
				"bic_blackbox.joe",
			},
		},
		{"combined_ns_inclusion_and_db_inclusion",
			[]string{"bic_blackbox.joe", "bic_functions_test.joe", "bic_test.*"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar",
				"bic_blackbox.joe", "bic_functions_test.joe",
			},
		},
		{"ns_inclusion_and_combined_db_inclusion",
			[]string{"bic_blackbox.joe", "bic_test.*", "bic_blackbox.*"},
			[]string{
				"bic_test.eleanor", "bic_test.hello", "bic_test.bar",
				"bic_blackbox.joe", "bic_blackbox.alice", "bic_blackbox.bob",
			},
		},
		{"ns_inclusion_and_db_exclusion",
			[]string{"bic_blackbox.joe", "~bic_test.*"},
			[]string{"bic_blackbox.joe"},
		},
		{"combined_ns_inclusion_and_db_exclusion",
			[]string{"bic_blackbox.joe", "bic_test.bar", "~bic_test.*"},
			[]string{"bic_blackbox.joe"},
		},
		{"ns_inclusion_and_combined_db_exclusion",
			[]string{"bic_blackbox.joe", "~bic_blackbox.*", "~bic_test.*"},
			nil,
		},
		{"ns_inclusion_and_collection_inclusion",
			[]string{"bic_blackbox.joe", "*.eleanor"},
			[]string{"bic_blackbox.joe", "bic_functions_test.eleanor", "bic_test.eleanor"},
		},
		{"combined_ns_inclusion_and_collection_inclusion",
			[]string{"bic_blackbox.joe", "bic_functions_test.joe", "*.eleanor"},
			[]string{
				"bic_blackbox.joe", "bic_test.eleanor",
				"bic_functions_test.eleanor", "bic_functions_test.joe",
			},
		},
		{"ns_inclusion_and_combined_collection_inclusion",
			[]string{"bic_blackbox.joe", "bic_functions_test.joe", "*.eleanor"},
			[]string{
				"bic_blackbox.joe", "bic_test.eleanor",
				"bic_functions_test.joe", "bic_functions_test.eleanor",
			},
		},
		{"ns_inclusion_and_collection_exclusion",
			[]string{"bic_blackbox.joe", "~*.eleanor"},
			[]string{"bic_blackbox.joe"},
		},
		{"combined_ns_inclusion_and_collection_exclusion",
			[]string{"bic_blackbox.joe", "bic_functions_test.joe", "~*.eleanor"},
			[]string{"bic_blackbox.joe", "bic_functions_test.joe"},
		},
		{"ns_inclusion_and_combined_collection_exclusion",
			[]string{"bic_blackbox.joe", "~*.eleanor", "~*.joe"},
			nil,
		},
		{"ns_exclusion_and_db_exclusion",
			[]string{"~bic_blackbox.joe", "~bic_test.*"},
			[]string{
				"bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.eleanor", "bic_functions_test.bobby", "bic_functions_test.joe",
			},
		},
		{"combined_ns_exclusion_and_db_exclusion",
			[]string{"~bic_blackbox.joe", "~bic_functions_test.bobby", "~bic_test.*"},
			[]string{
				"bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.eleanor", "bic_functions_test.joe",
			},
		},
		{"ns_exclusion_and_combined_db_exclusion",
			[]string{"~bic_blackbox.joe", "~bic_test.*", "~bic_functions_test.*"},
			[]string{"bic_blackbox.alice", "bic_blackbox.bob"},
		},
		{"ns_exclusion_and_db_inclusion",
			[]string{"~bic_blackbox.joe", "bic_test.*"},
			[]string{"bic_test.eleanor", "bic_test.hello", "bic_test.bar"},
		},
		{"combined_ns_exclusion_and_db_inclusion",
			[]string{"~bic_blackbox.joe", "~bic_functions_test.bobby", "bic_functions_test.*"},
			[]string{"bic_functions_test.eleanor", "bic_functions_test.joe"},
		},
		{"ns_exclusion_and_combined_db_inclusion",
			[]string{"~bic_blackbox.joe", "bic_functions_test.*", "bic_blackbox.*"},
			[]string{
				"bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.eleanor", "bic_functions_test.bobby", "bic_functions_test.joe",
			},
		},
		{"ns_exclusion_and_collection_inclusion",
			[]string{"~bic_blackbox.joe", "*.bobby"},
			[]string{"bic_functions_test.bobby"},
		},
		{"combined_ns_exclusion_and_collection_inclusion",
			[]string{"~bic_blackbox.joe", "~bic_test.bobby", "*.hello"},
			[]string{"bic_test.hello"},
		},
		{"ns_exclusion_and_combined_collection_inclusion",
			[]string{"~bic_blackbox.joe", "*.bobby", "*.eleanor"},
			[]string{"bic_functions_test.bobby", "bic_functions_test.eleanor", "bic_test.eleanor"},
		},
		{"ns_exclusion_and_collection_exclusion",
			[]string{"~bic_blackbox.joe", "~*.eleanor"},
			[]string{
				"bic_test.hello", "bic_test.bar",
				"bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.bobby", "bic_functions_test.joe",
			},
		},
		{"combined_ns_exclusion_and_collection_exclusion",
			[]string{"~bic_blackbox.joe", "~bic_test.hello", "~*.eleanor"},
			[]string{
				"bic_test.bar",
				"bic_blackbox.alice", "bic_blackbox.bob",
				"bic_functions_test.bobby", "bic_functions_test.joe",
			},
		},
		{"ns_exclusion_and_combined_collection_exclusion",
			[]string{"~bic_blackbox.joe", "~*.eleanor", "~*.bob*"},
			[]string{
				"bic_test.hello", "bic_test.bar",
				"bic_blackbox.alice", "bic_functions_test.joe",
			},
		},
	}

	nsOpts := config.Default().Schema
	for _, test := range namespaceSelectorTests {
		t.Run(test.description, func(t *testing.T) {
			req := require.New(t)

			nsOpts.Sample.Namespaces = test.samplePattern

			opts := NewMongosqldConfig(&nsOpts, nil)
			sampler := NewSampler(opts, lgr, getSessionProvider(req))
			sampleSchema, err := sampler.Sample(context.Background())

			req.Nilf(err, "did not expect error in sampling")
			req.NotNilf(sampleSchema, "did not expect sample schema to be nil")

			yb, err := sampleSchema.ToDRDL().ToYAML()
			if err != nil {
				panic(err)
			}
			t.Logf("sampled schema:\n%s", string(yb))

			req.Equalf(len(test.expectedNamespaces), countTables(sampleSchema), "namespace count not equal")

			for _, expectedNamespace := range test.expectedNamespaces {
				req.Truef(containsNS(sampleSchema, expectedNamespace), "namespace %q should be in schema", expectedNamespace)
			}
		})
	}
}

func TestSampleTableAndColumnCollisions(t *testing.T) {
	provider, err := mongodb.NewSqldSessionProvider(cfg)
	if err != nil {
		t.Fatalf("failed to set up session provider to test server: %v", err)
	}

	session, err := provider.Session(context.Background())
	if err != nil {
		t.Fatalf("failed to set up session to test server: %v", err)
	}
	defer session.Close()

	cleanupData(session)
	defer cleanupData(session)

	req := require.New(t)

	doc1 := bsonutil.NewDArray(
		bsonutil.NewD(bsonutil.NewDocElem("XX", 2)),
		bsonutil.NewD(bsonutil.NewDocElem("xX_0", 4)),
		bsonutil.NewD(bsonutil.NewDocElem("xX", bsonutil.NewDArray(bsonutil.NewD(bsonutil.NewDocElem("c", 1))))),
		bsonutil.NewD(bsonutil.NewDocElem("Xx", bsonutil.NewDArray(bsonutil.NewD(bsonutil.NewDocElem("b", 3))))),
	)

	doc2 := bsonutil.NewDArray(bsonutil.NewD(bsonutil.NewDocElem("hello", 2)))

	t1 := "foo"
	t2 := fmt.Sprintf("%v_Xx_0", t1)
	t3 := "X"
	t4 := "x"

	doc := bsonutil.NewDArray(bsonutil.NewD())

	dbutils.InsertDocuments(session, db1, t1, doc1)
	dbutils.InsertDocuments(session, db1, t2, doc2)
	dbutils.InsertDocuments(session, db1, t3, doc)
	dbutils.InsertDocuments(session, db1, t4, doc)

	opts := NewMongosqldConfig(&cfg.Schema, nil)
	sampler := NewSampler(opts, lgr, getSessionProvider(req))
	sampleSchema, err := sampler.Sample(context.Background())

	req.Nil(err)
	req.NotNilf(sampleSchema, "sample schema is nil")
	dbutils.RunCmd(session, db2, bsonutil.NewD(bsonutil.NewDocElem("profile", 0)), &struct{}{})

	req.NotZero(countTables(sampleSchema), "no namespaces sampled")

	db1c1 := formatNamespace(db1, t1, false)
	db1c2 := formatNamespace(db1, t2, false)
	db1c3 := formatNamespace(db1, t3, false)
	db1c4 := formatNamespace(db1, t4, false)

	req.True(containsNS(sampleSchema, db1c1))
	req.True(containsNS(sampleSchema, db1c2))
	req.True(containsNS(sampleSchema, db1c3))
	req.True(containsNS(sampleSchema, db1c4))
	req.Equal(1, len(sampleSchema.Databases()))

	dbs := sampleSchema.DatabasesSorted()
	req.Equal(dbs[0].Name(), db1)
	req.Equal(len(dbs[0].Tables()), 6)

	type sqlTableMapping struct {
		Table, Collection string
	}

	type sqlColumnMapping struct {
		Column, Field string
	}

	expectedTableMappings := []sqlTableMapping{
		{"foo", "foo"},
		{"foo_xX", "foo"},
		{"foo_Xx_0", "foo_Xx_0"},
		{"foo_Xx_1", "foo"},
		{"X", "X"},
		{"x_0", "x"},
	}

	mappings := []sqlTableMapping{}
	for _, table := range dbs[0].TablesSorted() {
		mapping := sqlTableMapping{table.SQLName(), table.MongoName()}
		mappings = append(mappings, mapping)
	}

	req.Equal(mappings, expectedTableMappings)

	getColumnMappings := func(t *schema.Table) (mappings []sqlColumnMapping) {
		for _, c := range t.ColumnsSorted() {
			mapping := sqlColumnMapping{c.SQLName(), c.MongoName()}
			mappings = append(mappings, mapping)
		}
		return mappings
	}

	table := dbs[0].Table("foo_Xx_1")
	req.NotNilf(table, "did not find table foo_Xx_1")
	expectedColumnMappings := []sqlColumnMapping{
		{"_id", "_id"}, {"Xx.b", "Xx.b"}, {"Xx_idx", "Xx_idx"},
	}
	req.Equal(getColumnMappings(table), expectedColumnMappings)

	table = dbs[0].Table("foo_Xx_0")
	req.NotNilf(table, "did not find table foo_Xx_0")
	expectedColumnMappings = []sqlColumnMapping{
		{"_id", "_id"}, {"hello", "hello"},
	}
	req.Equal(getColumnMappings(table), expectedColumnMappings)

	table = dbs[0].Table("foo_xX")
	req.NotNilf(table, "did not find table foo_xX")
	expectedColumnMappings = []sqlColumnMapping{
		{"_id", "_id"}, {"xX.c", "xX.c"}, {"xX_idx", "xX_idx"},
	}
	req.Equal(getColumnMappings(table), expectedColumnMappings)

	table = dbs[0].Table("x_0")
	req.NotNilf(table, "did not find table x_0")
	expectedColumnMappings = []sqlColumnMapping{{"_id", "_id"}}
	req.Equal(getColumnMappings(table), expectedColumnMappings)

	table = dbs[0].Table("X")
	req.NotNilf(table, "did not find table X")
	req.Equal(getColumnMappings(table), expectedColumnMappings)

	table = dbs[0].Table("foo")
	req.NotNilf(table, "did not find table foo")
	expectedColumnMappings = []sqlColumnMapping{
		{"_id", "_id"}, {"XX", "XX"}, {"xX_0", "xX_0"},
	}
	req.Equal(getColumnMappings(table), expectedColumnMappings)
}

func TestWriteModeRoundTrip(t *testing.T) {
	logger := log.NewComponentLogger(log.SchemaComponent, log.GlobalLogger())
	req := require.New(t)

	provider, err := mongodb.NewSqldSessionProvider(cfg)
	req.Nil(err)

	session, err := provider.Session(context.Background())
	req.Nil(err)
	defer session.Close()

	dbName := "foo"
	cleanupWriteData(session, dbName)
	defer cleanupData(session, dbName)

	tableName := "foo"
	table1 := newTableTestHelper(
		logger,
		tableName,
		tableName,
		[]bson.D{},
		[]*schema.Column{
			schema.NewColumn("a", schema.SQLInt, "a", schema.MongoInt64, false, option.NoneString()),
			schema.NewColumn("b", schema.SQLVarchar, "B", schema.MongoString, false, option.SomeString("fooo")),
			schema.NewColumn("c", schema.SQLVarchar, "C", schema.MongoString, true, option.SomeString("HELLO!")),
		},
		[]schema.Index{
			schema.NewIndex("bAr", true, false,
				[]schema.IndexPart{schema.NewIndexPart("a", 1), schema.NewIndexPart("b", -1)},
			),
			schema.NewIndex("", false, true,
				[]schema.IndexPart{schema.NewIndexPart("b", 1), schema.NewIndexPart("c", 1)},
			),
		},
		option.SomeString("WORLD"),
	)

	cc, err := table1.GenerateCreateCollection()
	req.Nil(err)
	ci := table1.GenerateCreateIndexes()
	ctx := context.Background()
	result := bsonutil.NewD()
	for _, c := range []*bson.D{&cc, &ci} {
		err = session.Run(ctx, dbName, *c, &result)
		req.Nil(err)
	}
	writeCfg := config.Default()
	writeCfg.Schema.WriteMode = true
	sampler := NewSampler(NewMongosqldConfig(&writeCfg.Schema, variable.NewGlobalContainer(writeCfg)), logger, provider)
	schema, err := sampler.Sample(ctx)
	req.Nil(err)
	schemaDB := schema.Database(dbName)
	schemaTable := schemaDB.Table(tableName)
	// call ColumnsSorted to populated the cachedSortedColumns so the test succeeds.
	_ = schemaTable.ColumnsSorted()
	req.Equal(table1, schemaTable, "tables should be equal")
}

func cleanupData(session *mongodb.Session, databases ...string) {
	dbutils.DropDatabase(session, cfg.Schema.Stored.Source)
	dbutils.DropDatabase(session, db1)
	dbutils.DropDatabase(session, db2)
	dbutils.DropDatabase(session, db3)
	for _, db := range databases {
		dbutils.DropDatabase(session, db)
	}
}

func cleanupWriteData(session *mongodb.Session, databases ...string) {
	for _, db := range databases {
		dbutils.DropDatabase(session, db)
	}
}

func containsNS(sch *schema.Schema, ns string) bool {
	parts := strings.Split(ns, ".")
	dbName := parts[0]
	tblName := parts[1]

	db := sch.Database(dbName)
	if db == nil {
		return false
	}

	tbl := db.Table(tblName)
	return tbl != nil
}

func countTables(sch *schema.Schema) int {
	count := 0
	for _, db := range sch.Databases() {
		count += len(db.Tables())
	}
	return count
}

func newTableTestHelper(lg log.Logger, tbl, col string,
	pipeline []bson.D, cols []*schema.Column,
	indexes []schema.Index, comment option.String) *schema.Table {
	out, err := schema.NewTable(lg, tbl, col, pipeline, cols, indexes, comment)
	if err != nil {
		panic("this table should not error")
	}
	return out
}
