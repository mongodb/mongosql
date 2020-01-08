package catalog_test

import (
	"context"
	"testing"

	"github.com/10gen/sqlproxy/evaluator/catalog"

	"github.com/stretchr/testify/require"
)

func TestCatalog(t *testing.T) {
	req := require.New(t)
	testCtx := context.Background()

	c := catalog.New("def")

	req.Equal(string(c.Name), "def")
	dbs, _ := c.Databases(testCtx)
	req.Equal(len(dbs), 0)

	_, err := c.Database(testCtx, "test")
	req.NotNil(err)

	_, err = c.AddDatabase("test")
	req.Nil(err)
	dbs, _ = c.Databases(testCtx)
	req.Equal(len(dbs), 1)

	_, err = c.AddDatabase("test")
	req.NotNil(err)

	_, err = c.AddDatabase("TEST")
	req.NotNil(err)

	_, err = c.Database(testCtx, "blah")
	req.NotNil(err)

	d, err := c.Database(testCtx, "test")
	req.Nil(err)
	req.NotNil(d)
	req.Equal(string(d.Name()), "test")

	d, err = c.Database(testCtx, "TEST")
	req.Nil(err)
	req.NotNil(d)
	req.Equal(string(d.Name()), "test")

	_, err = c.AddDatabase("test1")
	req.Nil(err)

	_, err = c.AddDatabase("test2")
	req.Nil(err)

	_, err = c.AddDatabase("test3")
	req.Nil(err)

	dbs, _ = c.Databases(testCtx)
	req.Equal(len(dbs), 4)
}

func TestDatabase(t *testing.T) {
	req := require.New(t)
	testCtx := context.Background()

	d, err := catalog.New("def").AddDatabase("test")
	req.Nil(err)
	req.NotNil(d)
	req.Equal(string(d.Name()), "test")
	tbls, _ := d.Tables(testCtx)
	req.Equal(len(tbls), 0)

	_, err = d.Table(testCtx, "foo")
	req.NotNil(err)

	t0 := catalog.NewInMemoryTable("foo")
	err = d.AddTable(t0)
	req.Nil(err)
	tbls, _ = d.Tables(testCtx)
	req.Equal(len(tbls), 1)

	t1 := catalog.NewInMemoryTable("foo")
	req.NotNil(d.AddTable(t1))

	_, err = d.Table(testCtx, "blah")
	req.NotNil(err)

	t2, err := d.Table(testCtx, "foo")
	req.Nil(err)
	req.Equal(t2.Name(), "foo")

	t3, err := d.Table(testCtx, "FOO")
	req.Nil(err)
	req.NotNil(d)
	req.Equal(t3.Name(), "foo")

	req.Nil(d.AddTable(catalog.NewInMemoryTable("foo1")))
	req.Nil(d.AddTable(catalog.NewInMemoryTable("foo2")))
	req.Nil(d.AddTable(catalog.NewInMemoryTable("foo3")))

	tbls, _ = d.Tables(testCtx)
	req.Equal(len(tbls), 4)
}
