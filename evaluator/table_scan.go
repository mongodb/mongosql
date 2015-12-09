package evaluator

import (
	"fmt"
	"github.com/10gen/sqlproxy/config"
	"github.com/mongodb/mongo-tools/common/bsonutil"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"strconv"
	"strings"
)

// TableScan is the primary interface for SQLProxy to a MongoDB
// installation and executes simple queries against collections.
type TableScan struct {
	dbName      string
	tableName   string
	matcher     SQLExpr
	iter        FindResults
	dbConfig    *config.Schema
	session     *mgo.Session
	tableConfig *config.TableConfig
	ctx         *ExecutionCtx
	err         error
}

// Open establishes a connection to database collection for this table.
func (ts *TableScan) Open(ctx *ExecutionCtx) error {
	return ts.init(ctx)
}

func (ts *TableScan) init(ctx *ExecutionCtx) error {
	ts.ctx = ctx

	if len(ts.dbName) == 0 {
		ts.dbName = ctx.Db
	}

	return ts.setIterator(ctx)
}

func (ts *TableScan) setIterator(ctx *ExecutionCtx) error {

	if len(ts.dbName) == 0 {
		ts.dbName = ctx.Db
	}

	ts.dbConfig = ctx.Config.Schemas[ts.dbName]
	if ts.dbConfig == nil {
		return fmt.Errorf("db (%s) doesn't exist - table (%s)", ts.dbName, ts.tableName)
	}

	ts.tableConfig = ts.dbConfig.Tables[ts.tableName]
	if ts.tableConfig == nil {
		return fmt.Errorf("table (%s) doesn't exist in db (%s)", ts.tableName, ts.dbName)
	}

	pcs := strings.SplitN(ts.tableConfig.Collection, ".", 2)

	ts.session = ctx.Session.Copy()
	db := ctx.Session.DB(pcs[0])
	collection := db.C(pcs[1])
	query := collection.Find(nil)
	ts.iter = MgoFindResults{query.Iter()}

	return nil
}

var bsonDType = reflect.TypeOf(bson.D{})

// extractFieldByName takes a field name and document, and returns a value representing
// the value of that field in the document in a format that can be printed as a string.
// It will also handle dot-delimited field names for nested arrays or documents.
func extractFieldByName(fieldName string, document interface{}) interface{} {
	dotParts := strings.Split(fieldName, ".")
	var subdoc interface{} = document

	for _, path := range dotParts {
		docValue := reflect.ValueOf(subdoc)
		if !docValue.IsValid() {
			return ""
		}
		docType := docValue.Type()
		docKind := docType.Kind()
		if docKind == reflect.Map {
			subdocVal := docValue.MapIndex(reflect.ValueOf(path))
			if subdocVal.Kind() == reflect.Invalid {
				return ""
			}
			subdoc = subdocVal.Interface()
		} else if docKind == reflect.Slice {
			if docType == bsonDType {
				// dive into a D as a document
				asD := subdoc.(bson.D)
				var err error
				subdoc, err = bsonutil.FindValueByKey(path, &asD)
				if err != nil {
					return ""
				}
			} else {
				//  check that the path can be converted to int
				arrayIndex, err := strconv.Atoi(path)
				if err != nil {
					return ""
				}
				// bounds check for slice
				if arrayIndex < 0 || arrayIndex >= docValue.Len() {
					return ""
				}
				subdocVal := docValue.Index(arrayIndex)
				if subdocVal.Kind() == reflect.Invalid {
					return ""
				}
				subdoc = subdocVal.Interface()
			}
		} else {
			// trying to index into a non-compound type - just return blank.
			return ""
		}
	}
	return subdoc
}

func (ts *TableScan) Next(row *Row) bool {
	if ts.iter == nil {
		return false
	}

	var hasNext bool

	for {
		d := &bson.D{}
		hasNext = ts.iter.Next(d)

		if !hasNext {
			break
		}

		values := Values{}
		data := d.Map()
		var err error

		for _, column := range ts.tableConfig.Columns {
			value := Value{
				Name: column.Name,
				View: column.Name,
			}

			if len(column.Source) != 0 {
				value.Data = extractFieldByName(column.Source, data)
			} else {
				value.Data = data[column.Name]
			}

			value.Data, err = NewSQLValue(value.Data, column.Type)
			if err != nil {
				ts.err = err
				return false
			}

			values = append(values, value)
			delete(data, column.Name)
		}

		// now add all other columns
		for key, value := range data {
			value := Value{
				Name: key,
				View: key,
				Data: value,
			}
			values = append(values, value)
		}
		row.Data = []TableRow{{ts.tableName, values, ts.tableConfig}}

		evalCtx := &EvalCtx{[]Row{*row}, ts.ctx}

		if ts.matcher != nil {
			m, err := Matches(ts.matcher, evalCtx)
			if err != nil {
				ts.err = err
				return false
			}
			if m {
				break
			}
		} else {
			break
		}
	}

	if !hasNext {
		ts.err = ts.iter.Err()
	}

	return hasNext
}

func (ts *TableScan) OpFields() []*Column {
	columns := []*Column{}

	// TODO: we currently only return headers from the schema
	// though the actual data is everything that comes from
	// the database
	for _, c := range ts.tableConfig.Columns {
		column := &Column{
			Table: ts.tableName,
			Name:  c.Name,
			View:  c.Name,
		}
		columns = append(columns, column)
	}

	return columns
}

func (ts *TableScan) Close() error {
	defer ts.session.Close()

	if ts.iter == nil {
		return nil
	}

	return ts.iter.Close()
}

func (ts *TableScan) Err() error {
	return ts.err
}
