package collation_test

import (
	"testing"

	"github.com/10gen/sqlproxy/collation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	req := require.New(t)

	subject, err := collation.Get(collation.Name("utf8_bin"))
	req.Nilf(err, "failed to get collation")

	req.Equalf(subject.Name, collation.Name("utf8_bin"), "unexpected name")
	req.Equalf(subject.ID, collation.ID(83), "unexpected id")
	req.Equalf(subject.CharsetName, collation.CharsetName("utf8"), "unexpected charset name")

	_, err = collation.Get(collation.Name("asdfasgwqegqweg"))
	req.NotNilf(err, "expected error but got none")

}

func TestGetByID(t *testing.T) {
	req := require.New(t)

	subject, err := collation.GetByID(collation.ID(83))
	req.Nilf(err, "failed to get collation")

	req.Equalf(subject.Name, collation.Name("utf8_bin"), "unexpected name")
	req.Equalf(subject.ID, collation.ID(83), "unexpected id")
	req.Equalf(subject.CharsetName, collation.CharsetName("utf8"), "unexpected charset name")

	_, err = collation.GetByID(collation.ID(0))
	req.NotNilf(err, "expected error but got none")
}

func TestMust(t *testing.T) {
	req := require.New(t)

	subject := collation.Must(collation.Get(collation.Name("utf8_bin")))

	req.Equalf(subject.Name, collation.Name("utf8_bin"), "unexpected name")
	req.Equalf(subject.ID, collation.ID(83), "unexpected id")
	req.Equalf(subject.CharsetName, collation.CharsetName("utf8"), "unexpected charset name")

	f := func() { collation.Must(collation.Get(collation.Name("asdfasgewgqwegqweg"))) }
	assert.Panics(t, f)
}
