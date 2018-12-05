package sample_test

import (
	"testing"

	"github.com/10gen/sqlproxy/internal/config"
	. "github.com/10gen/sqlproxy/internal/sample"
	"github.com/10gen/sqlproxy/log"

	"github.com/10gen/mongo-go-driver/bson"
	"github.com/stretchr/testify/require"
)

var (
	lgr = log.GlobalLogger()
	cfg = config.Default()
)

func TestNewNamespace(t *testing.T) {
	nsDb, nsCol, versionID := "db1", "c2", bson.NewObjectId()

	ns := NewNamespace(nsDb, nsCol, versionID)

	req := require.New(t)
	req.Equal(ns.Database, nsDb, "database names do not match: %v and %v", ns.Database, nsDb)
	req.Equal(ns.Collection, nsCol, "collection name does not match: %v and %v", ns.Collection, nsCol)
	req.Equal(ns.VersionID, versionID, "version id does not match: %v and %v", ns.VersionID, versionID)
}
