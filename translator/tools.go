package translator

import (
	"fmt"
	"strings"
	
	"github.com/mongodb/mongo-tools/common/log"
	"gopkg.in/mgo.v2/bson"

	"github.com/erh/mongo-sql-temp/config"
)

func caseInsensitiveEquals(a string, b string) bool {
	// TODO don't copy
	return strings.ToUpper(a) == strings.ToUpper(b)
}

func (e *Evalulator) PrintTableData(names []string, values [][]interface{}) {
	for _, name := range names {
		fmt.Printf("| %-9s ", name)
	}
	fmt.Printf(" |\n")

	for _, row := range values {
		for _, value := range row {
			fmt.Printf("| %-9v ", value)
		}
		fmt.Printf(" |\n")
	}
}

func getColumn(doc *bson.M, name string) interface{} {
	val := (*doc)[name]
	if val != nil {
		return val
	}

	for k, v := range *doc {
		if caseInsensitiveEquals(name, k) {
			return v
		}
	}

	return nil
}

func IterToNamesAndValues(iter FindResults, columns []config.Column, includeExtra bool) ([]string, [][]interface{}, error) {
	names := []string{}
	var seen map[string]bool = make(map[string]bool)
	
	for _, c := range(columns) {
		names = append(names, c.Name)
		seen[c.Name] = true
	}

	values := make([][]interface{}, 0)

	var doc bson.M
	for iter.Next(&doc) {
		columns := make([]interface{}, 0)

		for _, name := range names {
			columns = append(columns, getColumn(&doc, name))
		}

		if includeExtra {
			for k, v := range doc {
				if !seen[k] {
					names = append(names, k)
					seen[k] = true
					columns = append(columns, v)
				}
			}
		}

		values = append(values, columns)
	}

	// check for errors
	err := iter.Err()
	if err != nil {
		return nil, nil, err
	}

	// pad
	for idx, row := range values {
		for len(row) < len(names) {
			row = append(row, nil)
		}
		values[idx] = row
	}

	if len(values) == 0 {
		names = []string{}
	}

	log.Logf(log.DebugHigh, "%#v", values)

	return names, values, nil
}
