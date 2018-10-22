package parser

import (
	"fmt"
	"strconv"
)

func anonymizeStatement(stmt Statement) Statement {
	switch stmt.(type) {
	case *Select, *SimpleSelect, *Union:
		// these are the only kinds of queries we should be attempting
		// to anonymize (i.e. only SELECT queries)
	default:
		// attempting to anonymize anything else is a programmer error
		panic(fmt.Errorf(
			"can only anonymize select statements, but got a %T",
			stmt,
		))
	}

	newStmt, err := walk(newAnonymizer(), stmt.Copy())
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

func (a *anonymizer) renameDB(db []byte) []byte {
	if db == nil {
		return nil
	}
	str := string(db)
	sub, ok := a.dbs[str]
	if ok {
		return []byte(sub)
	}

	sub = fmt.Sprintf("db_%d", a.dbIdx)
	a.dbIdx++
	a.dbs[str] = sub
	return []byte(sub)
}

func (a *anonymizer) renameTable(tbl []byte) []byte {
	if tbl == nil {
		return nil
	}
	str := string(tbl)
	sub, ok := a.tables[str]
	if ok {
		return []byte(sub)
	}

	sub = fmt.Sprintf("tbl_%d", a.tblIdx)
	a.tblIdx++
	a.tables[str] = sub
	return []byte(sub)
}

func (a *anonymizer) renameColumn(col []byte) []byte {
	if col == nil {
		return nil
	}
	str := string(col)
	sub, ok := a.cols[str]
	if ok {
		return []byte(sub)
	}

	sub = fmt.Sprintf("col_%d", a.colIdx)
	a.colIdx++
	a.cols[str] = sub
	return []byte(sub)
}

func (a *anonymizer) PreVisit(current CST) (CST, error) {
	switch typed := current.(type) {
	case *TableName:
		typed.Name = a.renameTable(typed.Name)
		typed.Qualifier = a.renameDB(typed.Qualifier)
	case *AliasedTableExpr:
		typed.As = a.renameTable(typed.As)
	case *ColName:
		typed.Name = a.renameColumn(typed.Name)
		typed.Qualifier = a.renameTable(typed.Qualifier)
		typed.Database = a.renameDB(typed.Database)
	case *StarExpr:
		typed.TableName = a.renameTable(typed.TableName)
		typed.DatabaseName = a.renameDB(typed.DatabaseName)
	case *NonStarExpr:
		typed.As = a.renameColumn(typed.As)
	case Comments:
		return Comments([][]byte{}), nil
	case *DateVal:
		typed.Val = []byte("2012-03-04")
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
