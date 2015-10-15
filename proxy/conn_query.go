package proxy

import (
	"fmt"
	"github.com/erh/mixer/sqlparser"
	"github.com/mongodb/mongo-tools/common/log"
	"github.com/siddontang/mixer/client"
	"github.com/siddontang/mixer/hack"
	. "github.com/siddontang/mixer/mysql"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
)

func (c *Conn) handleQuery(sql string) (err error) {
	log.Logf(log.Always, "---- handleQuery: %s\n", sql)

	defer func() {
		if e := recover(); e != nil {
			log.Logf(log.Always, "%s\n", debug.Stack())
			err = fmt.Errorf("execute %s error %v", sql, e)
			return
		}
	}()

	sql = strings.TrimRight(sql, ";")

	var stmt sqlparser.Statement
	stmt, err = sqlparser.Parse(sql)
	if err != nil {

		// This is an ugly hack such that if someone tries to set some parameter to the default
		// ignore.  This is because the sql parser barfs.  We should probably fix there for reals.
		sqlUpper := strings.ToUpper(sql)
		if sqlUpper[0:4] == "SET " && sqlUpper[len(sqlUpper)-8:] == "=DEFAULT" {
			// wow, this is ugly
			return c.writeOK(nil)
		}

		return fmt.Errorf(`parse sql "%s" error: %s`, sql, err)
	}

	switch v := stmt.(type) {
	case *sqlparser.Select:
		return c.handleSelect(v, sql, nil)
	case *sqlparser.Insert:
		return c.handleExec(stmt, sql, nil)
	case *sqlparser.Update:
		return c.handleExec(stmt, sql, nil)
	case *sqlparser.Delete:
		return c.handleExec(stmt, sql, nil)
	case *sqlparser.Replace:
		return c.handleExec(stmt, sql, nil)
	case *sqlparser.Set:
		return c.handleSet(v)
	case *sqlparser.Begin:
		return c.handleBegin()
	case *sqlparser.Commit:
		return c.handleCommit()
	case *sqlparser.Rollback:
		return c.handleRollback()
	case *sqlparser.SimpleSelect:
		return c.handleSimpleSelect(sql, v)
	case *sqlparser.Show:
		return c.handleShow(sql, v)
	case *sqlparser.Admin:
		return c.handleAdmin(v)
	case *sqlparser.DDL:
		return c.handleDDL(v)
	default:
		return fmt.Errorf("statement %T not supported now", stmt)
	}

	return nil
}

func (c *Conn) getConn(n, isSelect bool) (co *client.SqlConn, err error) {
	return nil, fmt.Errorf("getConn is broken")
}

func (c *Conn) getShardConns(isSelect bool, stmt sqlparser.Statement, bindVars map[string]interface{}) ([]*client.SqlConn, error) {
	return nil, fmt.Errorf("getShardConns is broken")
}

func (c *Conn) executeInShard(conns []*client.SqlConn, sql string, args []interface{}) ([]*Result, error) {
	var wg sync.WaitGroup
	wg.Add(len(conns))

	rs := make([]interface{}, len(conns))

	f := func(rs []interface{}, i int, co *client.SqlConn) {
		r, err := co.Execute(sql, args...)
		if err != nil {
			rs[i] = err
		} else {
			rs[i] = r
		}

		wg.Done()
	}

	for i, co := range conns {
		go f(rs, i, co)
	}

	wg.Wait()

	var err error
	r := make([]*Result, len(conns))
	for i, v := range rs {
		if e, ok := v.(error); ok {
			err = e
			break
		}
		r[i] = rs[i].(*Result)
	}

	return r, err
}

func (c *Conn) closeShardConns(conns []*client.SqlConn, rollback bool) {
	if c.isInTransaction() {
		return
	}

	for _, co := range conns {
		if rollback {
			co.Rollback()
		}

		co.Close()
	}
}

func (c *Conn) newEmptyResultset(stmt *sqlparser.Select) *Resultset {
	r := new(Resultset)
	r.Fields = make([]*Field, len(stmt.SelectExprs))

	for i, expr := range stmt.SelectExprs {
		r.Fields[i] = &Field{}
		switch e := expr.(type) {
		case *sqlparser.StarExpr:
			r.Fields[i].Name = []byte("*")
		case *sqlparser.NonStarExpr:
			if e.As != nil {
				r.Fields[i].Name = e.As
				r.Fields[i].OrgName = hack.Slice(nstring(e.Expr))
			} else {
				r.Fields[i].Name = hack.Slice(nstring(e.Expr))
			}
		default:
			r.Fields[i].Name = hack.Slice(nstring(e))
		}
	}

	r.Values = make([][]interface{}, 0)
	r.RowDatas = make([]RowData, 0)

	return r
}

func makeBindVars(args []interface{}) map[string]interface{} {
	bindVars := make(map[string]interface{}, len(args))

	for i, v := range args {
		bindVars[fmt.Sprintf("v%d", i+1)] = v
	}

	return bindVars
}

func (c *Conn) handleSelect(stmt *sqlparser.Select, sql string, args []interface{}) error {
	log.Logf(log.DebugLow, "handleSelect sql: %s", sql)
	log.Logf(log.DebugLow, "handleSelect stmt: %#v", stmt)
	log.Logf(log.DebugLow, "handleSelect args: %#v", args)

	// TODO: deal with this
	// bindVars := makeBindVars(args)

	var currentDB string = ""
	if c.currentSchema != nil {
		currentDB = c.currentSchema.DB
	}

	names, values, err := c.server.eval.EvalSelect(currentDB, sql, stmt, c)
	if err != nil {
		return err
	}

	rs, err := c.buildResultset(names, values)
	if err != nil {
		return err
	}

	return c.writeResultset(c.status, rs)
}

func (c *Conn) handleExec(stmt sqlparser.Statement, sql string, args []interface{}) error {
	return fmt.Errorf("no exec statements allowed")
}

func (c *Conn) mergeSelectResult(rs []*Result, stmt *sqlparser.Select) error {
	r := rs[0].Resultset

	status := c.status | rs[0].Status

	for i := 1; i < len(rs); i++ {
		status |= rs[i].Status

		//check fields equal

		for j := range rs[i].Values {
			r.Values = append(r.Values, rs[i].Values[j])
			r.RowDatas = append(r.RowDatas, rs[i].RowDatas[j])
		}
	}

	//to do order by, group by, limit offset
	c.sortSelectResult(r, stmt)
	//to do, add log here, sort may error because order by key not exist in resultset fields

	if err := c.limitSelectResult(r, stmt); err != nil {
		return err
	}

	return c.writeResultset(status, r)
}

func (c *Conn) sortSelectResult(r *Resultset, stmt *sqlparser.Select) error {
	if stmt.OrderBy == nil {
		return nil
	}

	sk := make([]SortKey, len(stmt.OrderBy))

	for i, o := range stmt.OrderBy {
		sk[i].Name = nstring(o.Expr)
		sk[i].Direction = o.Direction
	}

	return r.Sort(sk)
}

func (c *Conn) limitSelectResult(r *Resultset, stmt *sqlparser.Select) error {
	if stmt.Limit == nil {
		return nil
	}

	var offset, count int64
	var err error
	if stmt.Limit.Offset == nil {
		offset = 0
	} else {
		if o, ok := stmt.Limit.Offset.(sqlparser.NumVal); !ok {
			return fmt.Errorf("invalid select limit %s", nstring(stmt.Limit))
		} else {
			if offset, err = strconv.ParseInt(hack.String([]byte(o)), 10, 64); err != nil {
				return err
			}
		}
	}

	if o, ok := stmt.Limit.Rowcount.(sqlparser.NumVal); !ok {
		return fmt.Errorf("invalid limit %s", nstring(stmt.Limit))
	} else {
		if count, err = strconv.ParseInt(hack.String([]byte(o)), 10, 64); err != nil {
			return err
		} else if count < 0 {
			return fmt.Errorf("invalid limit %s", nstring(stmt.Limit))
		}
	}

	if offset+count > int64(len(r.Values)) {
		count = int64(len(r.Values)) - offset
	}

	r.Values = r.Values[offset : offset+count]
	r.RowDatas = r.RowDatas[offset : offset+count]

	return nil
}
