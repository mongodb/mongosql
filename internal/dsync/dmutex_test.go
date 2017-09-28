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

	lgr := log.GlobalLogger()
	mtxConfig.SessionProvider = sp
	mtxConfig.Logger = &lgr

	Convey("Subject: DMutex_Lock", t, func() {
		cleanupData(session)
		mtx := NewDMutex(mtxConfig)

		Convey("when no lock exists, the lock should be acquired", func() {
			err = mtx.Lock(context.Background())
			So(err, ShouldBeNil)

			result := dbutils.Exists(session, mtxConfig.DatabaseName, mtxConfig.CollectionName, bson.D{
				{"_id", mtxConfig.Name},
				{"processName", mtxConfig.ProcessName},
			})

			So(result, ShouldBeTrue)
		})

		Convey("when the lock exists and belongs to our process, the lock should be acquired", func() {
			dt := time.Now().UTC().Add(10 * time.Second)
			lockDoc := bson.D{
				{"_id", mtxConfig.Name},
				{"processName", mtxConfig.ProcessName},
				{"expirationTime", dt},
			}

			dbutils.InsertDocuments(session, mtxConfig.DatabaseName, mtxConfig.CollectionName, []bson.D{lockDoc})

			err = mtx.Lock(context.Background())
			So(err, ShouldBeNil)

			result := dbutils.Exists(session, mtxConfig.DatabaseName, mtxConfig.CollectionName, bson.D{
				{"_id", mtxConfig.Name},
				{"processName", mtxConfig.ProcessName},
				{"expirationTime", bson.D{{"$gt", dt}}},
			})

			So(result, ShouldBeTrue)
		})

		Convey("when an expired lock exists and belongs to another process, the lock should be acquired", func() {
			dt := time.Now().UTC().Add(-1 * time.Minute)
			lockDoc := bson.D{
				{"_id", mtxConfig.Name},
				{"processName", "some other process"},
				{"expirationTime", dt},
			}

			dbutils.InsertDocuments(session, mtxConfig.DatabaseName, mtxConfig.CollectionName, []bson.D{lockDoc})

			err = mtx.Lock(context.Background())
			So(err, ShouldBeNil)

			result := dbutils.Exists(session, mtxConfig.DatabaseName, mtxConfig.CollectionName, bson.D{
				{"_id", mtxConfig.Name},
				{"processName", mtxConfig.ProcessName},
				{"expirationTime", bson.D{{"$gt", dt}}},
			})

			So(result, ShouldBeTrue)
		})

		Convey("when a lock exists and belongs to another process, the lock should not be acquired", func() {
			dt := time.Now().UTC().Add(1 * time.Minute)
			lockDoc := bson.D{
				{"_id", mtxConfig.Name},
				{"processName", "some other process"},
				{"expirationTime", dt},
			}

			dbutils.InsertDocuments(session, mtxConfig.DatabaseName, mtxConfig.CollectionName, []bson.D{lockDoc})

			err = mtx.Lock(context.Background())
			So(err, ShouldNotBeNil)

			result := dbutils.Exists(session, mtxConfig.DatabaseName, mtxConfig.CollectionName, bson.D{
				{"_id", mtxConfig.Name},
				{"processName", mtxConfig.ProcessName},
			})

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

	lgr := log.GlobalLogger()
	mtxConfig.SessionProvider = sp
	mtxConfig.Logger = &lgr

	Convey("Subject: DMutex_Unlock", t, func() {
		cleanupData(session)
		mtx := NewDMutex(mtxConfig)

		Convey("when no lock exists, the lock should be deleted", func() {
			err = mtx.Unlock(context.Background())
			So(err, ShouldBeNil)

			result := dbutils.Exists(session, mtxConfig.DatabaseName, mtxConfig.CollectionName, bson.D{
				{"_id", mtxConfig.Name},
			})

			So(result, ShouldBeFalse)
		})

		Convey("when the lock exists and belongs to our process, the lock should be deleted", func() {
			lockDoc := bson.D{
				{"_id", mtxConfig.Name},
				{"processName", mtxConfig.ProcessName},
				{"expirationTime", time.Now().UTC().Add(10 * time.Minute)},
			}

			dbutils.InsertDocuments(session, mtxConfig.DatabaseName, mtxConfig.CollectionName, []bson.D{lockDoc})

			err = mtx.Unlock(context.Background())
			So(err, ShouldBeNil)

			result := dbutils.Exists(session, mtxConfig.DatabaseName, mtxConfig.CollectionName, bson.D{
				{"_id", mtxConfig.Name},
			})

			So(result, ShouldBeFalse)
		})

		Convey("when an expired lock exists and belongs to our process, the lock should be deleted", func() {
			lockDoc := bson.D{
				{"_id", mtxConfig.Name},
				{"processName", mtxConfig.ProcessName},
				{"expirationTime", time.Now().UTC().Add(-1 * time.Minute)},
			}

			dbutils.InsertDocuments(session, mtxConfig.DatabaseName, mtxConfig.CollectionName, []bson.D{lockDoc})

			err = mtx.Unlock(context.Background())
			So(err, ShouldBeNil)

			result := dbutils.Exists(session, mtxConfig.DatabaseName, mtxConfig.CollectionName, bson.D{
				{"_id", mtxConfig.Name},
			})

			So(result, ShouldBeFalse)
		})

		Convey("when a lock exists and belongs to another process, the lock should not be deleted", func() {
			lockDoc := bson.D{
				{"_id", mtxConfig.Name},
				{"processName", "some other process"},
			}

			dbutils.InsertDocuments(session, mtxConfig.DatabaseName, mtxConfig.CollectionName, []bson.D{lockDoc})

			err = mtx.Lock(context.Background())
			So(err, ShouldNotBeNil)

			result := dbutils.Exists(session, mtxConfig.DatabaseName, mtxConfig.CollectionName, bson.D{
				{"_id", mtxConfig.Name},
			})

			So(result, ShouldBeTrue)
		})
	})
}

func cleanupData(session *mongodb.Session) {
	dbutils.DropDatabase(session, mtxConfig.DatabaseName)
}
