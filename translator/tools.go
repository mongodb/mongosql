package translator

import (
	"fmt"
	
	"github.com/mongodb/mongo-tools/common/log"
	"gopkg.in/mgo.v2/bson"

	"github.com/erh/mongo-sql-temp/config"
)

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

func IterToNamesAndValues(iter FindResults, columns []config.Column) ([]string, [][]interface{}, error) {
	names := []string{}
	var seen map[string]bool = make(map[string]bool)
	
	gotId := false
	for _, c := range(columns) {
		if c.Name == "_id" {
			gotId = true
		}
		names = append(names, c.Name)
		seen[c.Name] = true
	}

	if !gotId {
		names = append(names, "_id")
	}
	
	values := make([][]interface{}, 0)

	seen["_id"] = true

	var first bool = true
	var doc bson.M
	for iter.Next(&doc) {
		if first {
			first = false
			for name, _ := range doc {
				if name != "_id" && !seen[name] {
					names = append(names, name)
					seen[name] = true
				}
			}
		}

		columns := make([]interface{}, 0)

		for _, name := range names {
			columns = append(columns, doc[name])
		}

		for k, v := range doc {
			if !seen[k] {
				names = append(names, k)
				seen[k] = true
				columns = append(columns, v)
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
