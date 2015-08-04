package sqleval

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"sort"
)

func (e *Evalulator)PrintTableData(names []string, values [][]interface{}) {
	for _, name := range(names) {
		fmt.Printf("| %-9s ", name)
	}
	fmt.Printf(" |\n")

	for _, row := range(values) {
		for _, value := range(row) {
			fmt.Printf("| %-9v ", value)
		}
		fmt.Printf(" |\n")
	}
}

func IterToNamesAndValues(iter *mgo.Iter) ([]string, [][]interface{}, error) {

	names := []string{"_id"} // we want this to be first
	values := make([][]interface{}, 0)

	var seen map[string]bool = make(map[string]bool)
	seen["_id"] = true

	var first bool = true
	var doc bson.M
	for iter.Next(&doc) {

		if first {
			first = false
			for name,_ := range(doc) {
				if name != "_id" {
					names = append(names, name)
					seen[name] = true
				}
			}
			sort.Strings(names)
		}

		columns := make([]interface{}, 0)
		
		for _,name := range(names) {
			columns = append(columns, doc[name])
		}
		
		for k,v := range(doc) {
			if !seen[k] {
				names = append(names, k)
				seen[k] = true
				columns = append(columns, v)
			}
		}

		values = append(values, columns)
	}

	err := iter.Err()
	if err != nil {
		return nil, nil, err
	}
	
	for idx,row := range(values) {
		for len(row) < len(names) {
			row = append(row, nil)
		}
		values[idx] = row
	}
	
	return names, values, nil
}
