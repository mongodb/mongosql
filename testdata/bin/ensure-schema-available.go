package main

import (
	"database/sql"

	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	timeout  = flag.Int64("timeout", 60, "How many seconds to wait for mongosqld schema to become available")
	bindAddr = flag.String("addr", "127.0.0.1:3307", "")
)

const (
	schemaError = "MongoDB schema not yet available"
)

func init() {
	flag.Parse()
}

func main() {
	addr := fmt.Sprintf("root@tcp(%v)/information_schema?allowNativePasswords=1", *bindAddr)

	db, err := sql.Open("mysql", addr)
	if err != nil {
		fmt.Printf("can not connect: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	for i := int64(0); i < *timeout; i++ {
		_, err := db.Exec("use information_schema")
		if err != nil {
			if strings.Contains(err.Error(), schemaError) {
				time.Sleep(1 * time.Second)
				fmt.Printf("waiting for schema...\n")
				continue
			}
			fmt.Printf("error waiting for schema: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("schema is now available!")
		os.Exit(0)
	}

	fmt.Printf("schema not available after %v seconds\n", *timeout)
	os.Exit(1)
}
