package main

import (
	"fmt"

	"github.com/10gen/mongosql-rs/go/mongosql"
)

func main() {
	fmt.Println("HELLO!")
	v := mongosql.Version()
	fmt.Printf("version: %s\n", v)
	tr := mongosql.Translate("select * from foo join bar")
	fmt.Printf("got translation: %s\n", tr)
}
