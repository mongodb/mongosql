package evaluator

import (
	"fmt"
	"strings"

	"gopkg.in/mgo.v2/bson"
)

// mappingRegistry provides a way to get a field name from a table/column.
type mappingRegistry struct {
	columns []*Column
	fields  map[string]map[string]string
}

func (mr *mappingRegistry) addColumn(column *Column) {
	mr.columns = append(mr.columns, column)
}

func (mr *mappingRegistry) registerMapping(tbl, column, field string) {

	if mr.fields == nil {
		mr.fields = make(map[string]map[string]string)
	}

	if _, ok := mr.fields[tbl]; !ok {
		mr.fields[tbl] = make(map[string]string)
	}

	mr.fields[tbl][column] = field
}

func (mr *mappingRegistry) lookupFieldName(tbl, column string) (string, bool) {
	if mr.fields == nil {
		return "", false
	}

	columnToField, ok := mr.fields[tbl]
	if !ok {
		return "", false
	}

	field, ok := columnToField[column]
	return field, ok
}

// TableScan is the primary interface for SQLProxy to a MongoDB
// installation and executes simple queries against collections.
type TableScan struct {
	dbName          string
	tableName       string
	aliasName       string
	fqns            string // the fully qualified namespace in MongoDB
	mappingRegistry *mappingRegistry
	pipeline        []bson.D
	matcher         SQLExpr
	ctx             *ExecutionCtx
	iter            FindResults
	err             error
}

func NewTableScan(ctx *ExecutionCtx, dbName, tableName string, aliasName string) (*TableScan, error) {

	if dbName == "" {
		return nil, fmt.Errorf("dbName is empty")
	}
	if tableName == "" {
		return nil, fmt.Errorf("tableName is empty")
	}
	ts := &TableScan{
		dbName:    dbName,
		tableName: tableName,
		aliasName: aliasName,
	}

	if ts.aliasName == "" {
		ts.aliasName = ts.tableName
	}

	database, ok := ctx.Schema.Databases[ts.dbName]
	if !ok {
		return nil, fmt.Errorf("db (%s) doesn't exist - table (%s)", dbName, tableName)
	}
	tableSchema, ok := database.Tables[ts.tableName]
	if !ok {
		return nil, fmt.Errorf("table (%s) doesn't exist in db (%s)", tableName, dbName)
	}

	ts.fqns = tableSchema.FQNS

	ts.mappingRegistry = &mappingRegistry{}
	for _, c := range tableSchema.RawColumns {
		column := &Column{
			Table: ts.aliasName,
			Name:  c.SqlName,
			View:  c.SqlName,
			Type:  c.SqlType,
		}
		ts.mappingRegistry.addColumn(column)
		ts.mappingRegistry.registerMapping(ts.aliasName, c.SqlName, c.Name)
	}

	ts.pipeline = []bson.D{}

	return ts, nil
}

// WithPipeline creates a new TableScan operator by copying everything
// and changing only the pipeline.
func (ts *TableScan) WithPipeline(pipeline []bson.D) *TableScan {
	return &TableScan{
		dbName:          ts.dbName,
		tableName:       ts.tableName,
		aliasName:       ts.aliasName,
		fqns:            ts.fqns,
		matcher:         ts.matcher,
		mappingRegistry: ts.mappingRegistry,
		pipeline:        pipeline,
	}
}

// WithMappingRegistry creates a new TableScan operator by copying everything
// and changing only the mappingRegistry.
func (ts *TableScan) WithMappingRegistry(mappingRegistry *mappingRegistry) *TableScan {

	return &TableScan{
		dbName:          ts.dbName,
		tableName:       ts.tableName,
		aliasName:       ts.aliasName,
		fqns:            ts.fqns,
		matcher:         ts.matcher,
		pipeline:        ts.pipeline,
		mappingRegistry: mappingRegistry,
	}
}

// Open establishes a connection to database collection for this table.
func (ts *TableScan) Open(ctx *ExecutionCtx) error {
	ts.ctx = ctx

	pcs := strings.SplitN(ts.fqns, ".", 2)

	ts.iter = MgoFindResults{ctx.Session.DB(pcs[0]).C(pcs[1]).Pipe(ts.pipeline).Iter()}

	return nil
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

		values := make(map[string]Values)
		data := d.Map()

		var err error

		for _, column := range ts.mappingRegistry.columns {

			mappedFieldName, ok := ts.mappingRegistry.lookupFieldName(column.Table, column.Name)
			if !ok {
				ts.err = fmt.Errorf("Unable to find mapping from %v.%v to a field name.", column.Table, column.Name)
				return false
			}

			value := Value{
				Name: column.Name,
				View: column.View,
				Data: extractFieldByName(mappedFieldName, data),
			}

			value.Data, err = NewSQLValue(value.Data, column.Type)
			if err != nil {
				ts.err = err
				return false
			}

			tableName := column.Table
			if tableName == ts.tableName {
				tableName = ts.aliasName
			}

			if _, ok := values[tableName]; !ok {
				values[tableName] = Values{}
			}

			values[tableName] = append(values[tableName], value)
			delete(data, mappedFieldName)
		}

		tableRows := TableRows{}
		for k, v := range values {
			tableRows = append(tableRows, TableRow{k, v})
		}

		row.Data = tableRows

		evalCtx := &EvalCtx{Rows{*row}, ts.ctx}

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

	return hasNext
}

func (ts *TableScan) OpFields() (columns []*Column) {
	return ts.mappingRegistry.columns
}

func (ts *TableScan) Close() error {
	return ts.iter.Close()
}

func (ts *TableScan) Err() error {
	if err := ts.iter.Err(); err != nil {
		return err
	}
	return ts.err
}
