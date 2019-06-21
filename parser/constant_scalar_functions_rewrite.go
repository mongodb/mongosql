package parser

import (
	"fmt"
	"strconv"
	"time"

	"github.com/10gen/sqlproxy/schema"
)

// ConstantScalarFunctionRewriter contains information necessary to rewrite
// the user, version, database, and connID functions
type ConstantScalarFunctionRewriter struct {
	connID       uint64
	dbName       string
	mySQLVersion string
	remoteHost   string
	user         string
}

// PreVisit is called for every node before its children are walked.
func (*ConstantScalarFunctionRewriter) PreVisit(current CST) (CST, error) {
	return current, nil
}

var _ Walker = (*ConstantScalarFunctionRewriter)(nil)

// PostVisit is called for every node after its children are walked.
func (c *ConstantScalarFunctionRewriter) PostVisit(current CST) (CST, error) {

	if node, ok := current.(*FuncExpr); ok {
		switch funcName := node.Name; funcName {
		case "pi":
			return NumVal("3.141592653589793E0"), nil
		case "curtime", "current_time", "utc_time":
			return &DateVal{
				Name: "datetime",
				Val:  time.Now().In(schema.DefaultLocale).Format("2006-01-02 15:04:05.000000"),
			}, nil
		case "current_timestamp", "now":
			return &DateVal{
				Name: "datetime",
				Val:  time.Now().In(schema.DefaultLocale).Format("2006-01-02 15:04:05.000000"),
			}, nil
		case "curdate", "current_date":
			return &DateVal{
				Name: "date",
				Val:  time.Now().In(schema.DefaultLocale).Format("2006-01-02"),
			}, nil
		case "utc_timestamp":
			return &DateVal{
				Name: "datetime",
				Val:  time.Now().In(time.UTC).Format("2006-01-02 15:04:05.000000"),
			}, nil
		case "utc_date":
			return &DateVal{
				Name: "date",
				Val:  time.Now().In(time.UTC).Format("2006-01-02"),
			}, nil
		case "connection_id":
			selExprs := make(SelectExprs, 2)
			selExprs[0] = &NonStarExpr{Expr: NumVal(strconv.FormatUint(c.connID, 10))}
			selExprs[1] = &NonStarExpr{Expr: KeywordVal("unsigned")}
			return &FuncExpr{
				Name:  "cast",
				Exprs: selExprs,
			}, nil
		case "database", "schema":
			return StrVal(c.dbName), nil
		case "version":
			return StrVal(c.mySQLVersion), nil
		case "user", "current_user", "session_user", "system_user":
			str := fmt.Sprintf("%s@%s", c.user, c.remoteHost)
			return StrVal(str), nil
		}

	}
	return current, nil
}

// RewriteConstantScalarFunctions rewrites the constant scalar functions for
// pi, constant time/date functions, connection_id, database, version, and user
func RewriteConstantScalarFunctions(statement Statement, connID uint64, dbName, mySQLVersion, remoteHost, user string) (Statement, error) {
	result, err := Walk(&ConstantScalarFunctionRewriter{connID, dbName, mySQLVersion, remoteHost, user}, statement)
	if err != nil {
		return nil, err
	}
	return result.(Statement), nil
}
