package sample_test

import (
	"context"
	"testing"

	"time"

	"github.com/10gen/sqlproxy/internal/config"
	. "github.com/10gen/sqlproxy/internal/sample"
	"github.com/10gen/sqlproxy/internal/testutils/dbutils"
	"github.com/10gen/sqlproxy/mongodb"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSampler_Refresh(t *testing.T) {
	provider, err := mongodb.NewSqldSessionProvider(cfg)
	if err != nil {
		t.Fatalf("failed to set up session provider to test server: %v", err)
	}

	session, err := provider.Session(context.Background())
	if err != nil {
		t.Fatalf("failed to set up session to test server: %v", err)
	}
	defer session.Close()

	Convey("Given a database with data", t, func() {
		cleanupData(session)
		dbutils.InsertDocuments(session, db1, c1, doc)
		dbutils.InsertDocuments(session, db1, c2, doc)

		Convey("and a standalone sampler", func() {
			schemaOptions := &config.SchemaSampleOptions{
				Mode:       config.ReadSampleMode,
				Namespaces: cfg.Schema.Sample.Namespaces,
			}
			sampler := NewSampler(schemaOptions, "pname", provider)
			ctx, cancel := context.WithCancel(context.Background())
			go sampler.Run(ctx)
			time.Sleep(5 * time.Second)

			Convey("the refreshed schema should be different after creating a new db", func() {
				dbutils.InsertDocuments(session, db2, c1, doc)
				err = sampler.Refresh(ctx)
				So(err, ShouldBeNil)

				newSchema := sampler.Schema(ctx)
				cancel()

				So(len(newSchema.Databases()), ShouldEqual, 2)
			})
		})

		Convey("and a clustered read sampler", func() {
			schemaOptions := &config.SchemaSampleOptions{
				Mode:       config.ReadSampleMode,
				Namespaces: cfg.Schema.Sample.Namespaces,
				Source:     cfg.Schema.Sample.Source,
			}
			sampler := NewSampler(schemaOptions, "pname", provider)
			ctx, cancel := context.WithCancel(context.Background())
			go sampler.Run(ctx)
			time.Sleep(5 * time.Second)

			Convey("the refresh call should error", func() {
				err = sampler.Refresh(ctx)
				cancel()
				So(err, ShouldNotBeNil)
			})
		})

		Convey("and a clustered write sampler", func() {
			schemaOptions := &config.SchemaSampleOptions{
				Mode:       config.WriteSampleMode,
				Namespaces: cfg.Schema.Sample.Namespaces,
				Source:     cfg.Schema.Sample.Source,
			}

			sampler := NewSampler(schemaOptions, "pname", provider)
			ctx, cancel := context.WithCancel(context.Background())
			go sampler.Run(ctx)
			time.Sleep(5 * time.Second)

			Convey("the refreshed schema should be different after creating a new db", func() {
				dbutils.InsertDocuments(session, db2, c1, doc)
				err = sampler.Refresh(ctx)
				So(err, ShouldBeNil)

				newSchema := sampler.Schema(ctx)
				cancel()

				So(len(newSchema.Databases()), ShouldEqual, 2)

				Convey("and it should be persisted to the database", func() {
					cursor := dbutils.Find(session, schemaOptions.Source, SchemasCollection, 1000)
					initialBatch := cursor.InitialBatch()
					So(len(initialBatch), ShouldEqual, 5) // the original 2 plus the new 3
				})
			})
		})
	})
}
