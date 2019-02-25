package catalog_test

import (
	"testing"

	"github.com/10gen/sqlproxy/evaluator/catalog"

	"github.com/stretchr/testify/require"
)

func TestCatalog(t *testing.T) {
	req := require.New(t)

	c := catalog.New("def", nil)

	req.Equal(string(c.Name), "def")
	req.Equal(len(c.Databases()), 0)

	_, err := c.Database("test")
	req.NotNil(err)

	_, err = c.AddDatabase("test")
	req.Nil(err)
	req.Equal(len(c.Databases()), 1)

	_, err = c.AddDatabase("test")
	req.NotNil(err)

	_, err = c.AddDatabase("TEST")
	req.NotNil(err)

	_, err = c.Database("blah")
	req.NotNil(err)

	d, err := c.Database("test")
	req.Nil(err)
	req.NotNil(d)
	req.Equal(string(d.Name()), "test")

	d, err = c.Database("TEST")
	req.Nil(err)
	req.NotNil(d)
	req.Equal(string(d.Name()), "test")

	_, err = c.AddDatabase("test1")
	req.Nil(err)

	_, err = c.AddDatabase("test2")
	req.Nil(err)

	_, err = c.AddDatabase("test3")
	req.Nil(err)

	req.Equal(len(c.Databases()), 4)
}

func TestDatabase(t *testing.T) {
	req := require.New(t)

	d, err := catalog.New("def", nil).AddDatabase("test")
	req.Nil(err)
	req.NotNil(d)
	req.Equal(string(d.Name()), "test")
	req.Equal(len(d.Tables()), 0)

	_, err = d.Table("foo")
	req.NotNil(err)

	t0 := catalog.NewInMemoryTable("foo")
	err = d.AddTable(t0)
	req.Nil(err)
	req.Equal(len(d.Tables()), 1)

	t1 := catalog.NewInMemoryTable("foo")
	req.NotNil(d.AddTable(t1))

	_, err = d.Table("blah")
	req.NotNil(err)

	t2, err := d.Table("foo")
	req.Nil(err)
	req.Equal(t2.Name(), "foo")

	t3, err := d.Table("FOO")
	req.Nil(err)
	req.NotNil(d)
	req.Equal(t3.Name(), "foo")

	req.Nil(d.AddTable(catalog.NewInMemoryTable("foo1")))
	req.Nil(d.AddTable(catalog.NewInMemoryTable("foo2")))
	req.Nil(d.AddTable(catalog.NewInMemoryTable("foo3")))

	req.Equal(len(d.Tables()), 4)
}
