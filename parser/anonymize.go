package parser

import (
	"fmt"
	"strconv"
)

// AnonymizeStatement returns a query with column and literal values replaced
// with generic values.
func AnonymizeStatement(stmt Statement) Statement {
	switch stmt.(type) {
	case *Select, *Union:
		// these are the only kinds of queries we should be attempting
		// to anonymize (i.e. only SELECT queries)
	default:
		// attempting to anonymize anything else is a programmer error
		panic(fmt.Errorf(
			"can only anonymize select statements, but got a %T",
			stmt,
		))
	}

	newStmt, err := Walk(newAnonymizer(), stmt.Copy())
	if err != nil {
		panic(err)
	}

	return newStmt.(Statement)
}

type anonymizer struct {
	dbs, tables, cols     map[string]string
	dbIdx, tblIdx, colIdx int
}

func newAnonymizer() *anonymizer {
	return &anonymizer{
		dbs:    make(map[string]string),
		tables: make(map[string]string),
		cols:   make(map[string]string),
		dbIdx:  1,
		tblIdx: 1,
		colIdx: 1,
	}
}

func (a *anonymizer) renameDB(db string) string {
	sub, ok := a.dbs[db]
	if ok {
		return sub
	}

	sub = fmt.Sprintf("db_%d", a.dbIdx)
	a.dbIdx++
	a.dbs[db] = sub
	return sub
}

func (a *anonymizer) renameTable(tbl string) string {
	sub, ok := a.tables[tbl]
	if ok {
		return sub
	}

	sub = fmt.Sprintf("tbl_%d", a.tblIdx)
	a.tblIdx++
	a.tables[tbl] = sub
	return sub
}

func (a *anonymizer) renameColumn(col string) string {
	sub, ok := a.cols[col]
	if ok {
		return sub
	}

	sub = fmt.Sprintf("col_%d", a.colIdx)
	a.colIdx++
	a.cols[col] = sub
	return sub
}

func (a *anonymizer) PreVisit(current CST) (CST, error) {
	switch typed := current.(type) {
	case *TableName:
		typed.Name = a.renameTable(typed.Name)
		typed.Qualifier = typed.Qualifier.Map(a.renameDB)
	case *AliasedTableExpr:
		typed.As = typed.As.Map(a.renameTable)
	case *ColName:
		typed.Name = a.renameColumn(typed.Name)
		typed.Qualifier = typed.Qualifier.Map(a.renameTable)
		typed.Database = typed.Database.Map(a.renameDB)
	case *StarExpr:
		typed.TableName = typed.TableName.Map(a.renameTable)
		typed.DatabaseName = typed.DatabaseName.Map(a.renameDB)
	case *NonStarExpr:
		typed.As = typed.As.Map(a.renameColumn)
	case Comments:
		return Comments([]string{}), nil
	case *DateVal:
		typed.Val = "2012-03-04"
	case StrVal:
		return StrVal("abc"), nil
	case NumVal:
		_, err := strconv.ParseInt(string(typed), 10, 64)
		if err == nil {
			return NumVal("123"), nil
		}
		return NumVal("4.56"), nil
	case *NullVal, *UnknownVal:
		return &NullVal{}, nil
	case *TrueVal, *FalseVal:
		return &FalseVal{}, nil
	case *ValArg:
		panic("cannot anonymize ValArg")
	}
	return current, nil
}

func (a *anonymizer) PostVisit(current CST) (CST, error) {
	return current, nil
}
