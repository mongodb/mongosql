package main

import (
	"fmt"

	"github.com/10gen/mongosql-rs/go/mongosql"
)

func main() {
	v := mongosql.Version()
	fmt.Printf("version: %s\n", v)
	tr := mongosql.Translate("select")
	fmt.Printf("got translation: %s\n", tr)
}
