package mongodb_test

import (
	"testing"

	"github.com/10gen/sqlproxy/mongodb"

	"github.com/stretchr/testify/require"
)

func TestVersionAtLeast(t *testing.T) {
	t.Run("Subject: VersionAtLeast", func(t *testing.T) {
		req := require.New(t)

		info := &mongodb.Info{
			VersionArray: []uint8{3, 2, 1},
		}

		req.True(info.VersionAtLeast(3, 2, 1), "should be true")
		req.False(info.VersionAtLeast(3, 2, 2), "should be false")
		req.False(info.VersionAtLeast(3, 3, 0), "should be false")
		req.False(info.VersionAtLeast(4, 0, 0), "should be false")
		req.False(info.VersionAtLeast(4, 4, 4), "should be false")
		req.True(info.VersionAtLeast(3, 2, 0), "should be true")
		req.True(info.VersionAtLeast(3, 0, 2), "should be true")
		req.True(info.VersionAtLeast(2, 3, 3), "should be true")
	})

	t.Run("Subject: VersionsAtLeast With Compatible Version", func(t *testing.T) {
		req := require.New(t)

		info := &mongodb.Info{
			VersionArray: []uint8{3, 0, 0},
		}
		req.Nil(info.SetCompatibleVersion("3.2.1"), "error setting compatibility version")

		req.True(info.VersionAtLeast(3, 2, 1), "should be true")
		req.False(info.VersionAtLeast(3, 2, 2), "should be false")
		req.False(info.VersionAtLeast(3, 3, 0), "should be false")
		req.False(info.VersionAtLeast(4, 0, 0), "should be false")
		req.False(info.VersionAtLeast(4, 4, 4), "should be false")
		req.True(info.VersionAtLeast(3, 2, 0), "should be true")
		req.True(info.VersionAtLeast(3, 0, 2), "should be true")
		req.True(info.VersionAtLeast(2, 3, 3), "should be true")
	})
}
