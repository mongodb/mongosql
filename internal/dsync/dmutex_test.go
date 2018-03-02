package dsync_test

import (
	"testing"

	"time"

	"context"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/10gen/sqlproxy/internal/config"
	. "github.com/10gen/sqlproxy/internal/dsync"
	"github.com/10gen/sqlproxy/internal/testutils/dbutils"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	mtxConfig = DMutexConfig{
		Name:              "test",
		DatabaseName:      "dsync_tests",
		CollectionName:    "dmutex",
		HeartbeatInterval: 0,
		ProcessName:       "dmutex_test",
		Timeout:           30 * time.Second,
	}
	cfg = config.Default()
)

func TestDMutex_Lock(t *testing.T) {
	sp, err := mongodb.NewSqldSessionProvider(cfg)
	if err != nil {
		t.Fatalf(err.Error())
	}

	session, err := sp.Session(context.Background())
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer session.Close()

	mtxConfig.SessionProvider = sp
	mtxConfig.Logger = log.GlobalLogger()

	Convey("Subject: DMutex_Lock", t, func() {
		cleanupData(session)
		mtx := NewDMutex(mtxConfig)

		Convey("when no lock exists, the lock should be acquired", func() {
			err = mtx.Lock(context.Background())
			So(err, ShouldBeNil)

			result := dbutils.Exists(
				session,
				mtxConfig.DatabaseName,
				mtxConfig.CollectionName,
				bson.D{
					{Name: "_id", Value: mtxConfig.Name},
					{Name: "processName", Value: mtxConfig.ProcessName},
				},
			)

			So(result, ShouldBeTrue)
		})

		Convey("when the lock exists and belongs to our process, it should be acquired", func() {
			dt := time.Now().UTC().Add(10 * time.Second)
			lockDoc := bson.D{
				{Name: "_id", Value: mtxConfig.Name},
				{Name: "processName", Value: mtxConfig.ProcessName},
				{Name: "expirationTime", Value: dt},
			}

			dbutils.InsertDocuments(
				session,
				mtxConfig.DatabaseName,
				mtxConfig.CollectionName,
				[]bson.D{lockDoc},
			)

			err = mtx.Lock(context.Background())
			So(err, ShouldBeNil)

			result := dbutils.Exists(
				session,
				mtxConfig.DatabaseName,
				mtxConfig.CollectionName,
				bson.D{
					{Name: "_id", Value: mtxConfig.Name},
					{Name: "processName", Value: mtxConfig.ProcessName},
					{Name: "expirationTime", Value: bson.D{{Name: "$gt", Value: dt}}},
				},
			)

			So(result, ShouldBeTrue)
		})

		Convey("when an expired lock belongs to another process, it should be acquired", func() {
			dt := time.Now().UTC().Add(-1 * time.Minute)
			lockDoc := bson.D{
				{Name: "_id", Value: mtxConfig.Name},
				{Name: "processName", Value: "some other process"},
				{Name: "expirationTime", Value: dt},
			}

			dbutils.InsertDocuments(
				session,
				mtxConfig.DatabaseName,
				mtxConfig.CollectionName,
				[]bson.D{lockDoc},
			)

			err = mtx.Lock(context.Background())
			So(err, ShouldBeNil)

			result := dbutils.Exists(
				session,
				mtxConfig.DatabaseName,
				mtxConfig.CollectionName,
				bson.D{
					{Name: "_id", Value: mtxConfig.Name},
					{Name: "processName", Value: mtxConfig.ProcessName},
					{Name: "expirationTime", Value: bson.D{{Name: "$gt", Value: dt}}},
				},
			)

			So(result, ShouldBeTrue)
		})

		Convey("when lock exists and belongs to another process, it shouldn't be acquired", func() {
			dt := time.Now().UTC().Add(1 * time.Minute)
			lockDoc := bson.D{
				{Name: "_id", Value: mtxConfig.Name},
				{Name: "processName", Value: "some other process"},
				{Name: "expirationTime", Value: dt},
			}

			dbutils.InsertDocuments(
				session,
				mtxConfig.DatabaseName,
				mtxConfig.CollectionName,
				[]bson.D{lockDoc},
			)

			err = mtx.Lock(context.Background())
			So(err, ShouldNotBeNil)

			result := dbutils.Exists(
				session,
				mtxConfig.DatabaseName,
				mtxConfig.CollectionName,
				bson.D{
					{Name: "_id", Value: mtxConfig.Name},
					{Name: "processName", Value: mtxConfig.ProcessName},
				},
			)

			So(result, ShouldBeFalse)
		})
	})
}

func TestDMutex_Unlock(t *testing.T) {
	sp, err := mongodb.NewSqldSessionProvider(cfg)
	if err != nil {
		t.Fatalf(err.Error())
	}

	session, err := sp.Session(context.Background())
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer session.Close()

	mtxConfig.SessionProvider = sp
	mtxConfig.Logger = log.GlobalLogger()

	Convey("Subject: DMutex_Unlock", t, func() {
		cleanupData(session)
		mtx := NewDMutex(mtxConfig)

		Convey("when no lock exists, the lock should be deleted", func() {
			err = mtx.Unlock(context.Background())
			So(err, ShouldBeNil)

			result := dbutils.Exists(
				session,
				mtxConfig.DatabaseName,
				mtxConfig.CollectionName,
				bson.D{
					{Name: "_id", Value: mtxConfig.Name},
				},
			)

			So(result, ShouldBeFalse)
		})

		Convey("when the lock exists and belongs to our process, it should be deleted", func() {
			lockDoc := bson.D{
				{Name: "_id", Value: mtxConfig.Name},
				{Name: "processName", Value: mtxConfig.ProcessName},
				{Name: "expirationTime", Value: time.Now().UTC().Add(10 * time.Minute)},
			}

			dbutils.InsertDocuments(
				session,
				mtxConfig.DatabaseName,
				mtxConfig.CollectionName,
				[]bson.D{lockDoc},
			)

			err = mtx.Unlock(context.Background())
			So(err, ShouldBeNil)

			result := dbutils.Exists(
				session,
				mtxConfig.DatabaseName,
				mtxConfig.CollectionName,
				bson.D{
					{Name: "_id", Value: mtxConfig.Name},
				},
			)

			So(result, ShouldBeFalse)
		})

		Convey("when expired lock exists and belongs to our process, it should be deleted", func() {
			lockDoc := bson.D{
				{Name: "_id", Value: mtxConfig.Name},
				{Name: "processName", Value: mtxConfig.ProcessName},
				{Name: "expirationTime", Value: time.Now().UTC().Add(-1 * time.Minute)},
			}

			dbutils.InsertDocuments(
				session,
				mtxConfig.DatabaseName,
				mtxConfig.CollectionName,
				[]bson.D{lockDoc},
			)

			err = mtx.Unlock(context.Background())
			So(err, ShouldBeNil)

			result := dbutils.Exists(
				session,
				mtxConfig.DatabaseName,
				mtxConfig.CollectionName,
				bson.D{
					{Name: "_id", Value: mtxConfig.Name},
				},
			)

			So(result, ShouldBeFalse)
		})

		Convey("when a lock exists and belongs to another process, should not be deleted", func() {
			lockDoc := bson.D{
				{Name: "_id", Value: mtxConfig.Name},
				{Name: "processName", Value: "some other process"},
			}

			dbutils.InsertDocuments(
				session,
				mtxConfig.DatabaseName,
				mtxConfig.CollectionName,
				[]bson.D{lockDoc},
			)

			err = mtx.Lock(context.Background())
			So(err, ShouldNotBeNil)

			result := dbutils.Exists(
				session,
				mtxConfig.DatabaseName,
				mtxConfig.CollectionName,
				bson.D{
					{Name: "_id", Value: mtxConfig.Name},
				},
			)

			So(result, ShouldBeTrue)
		})
	})
}

func cleanupData(session *mongodb.Session) {
	dbutils.DropDatabase(session, mtxConfig.DatabaseName)
}
