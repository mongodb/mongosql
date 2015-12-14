package evaluator

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func PrintExpressionTree(e SQLExpr, w io.Writer) error {
	bufW := bufio.NewWriter(w)
	v := &printingVisitor{w: bufW}
	_, err := v.Visit(e)
	bufW.Flush()
	return err
}

type printingVisitor struct {
	w           *bufio.Writer
	indentLevel int
}

func (v *printingVisitor) writeLine(s string) {
	v.w.WriteString(strings.Repeat(" ", v.indentLevel*4) + s + "\n")
}

func (v *printingVisitor) writeLinef(format string, a ...interface{}) {
	v.w.WriteString(strings.Repeat(" ", v.indentLevel*4) + fmt.Sprintf(format, a...) + "\n")
}

func (v *printingVisitor) Visit(e SQLExpr) (SQLExpr, error) {

	switch typedE := e.(type) {
	case *SQLAggFunctionExpr:
		v.writeLine("SQLAggFunctionExpr({func})")
	case *SQLAndExpr:
		v.writeLine("SQLAndExpr:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case *SQLAddExpr:
		v.writeLine("SQLAddExpr:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case *SQLCaseExpr:
		v.writeLine("SQLCaseExpr")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case *SQLDivideExpr:
		v.writeLine("SQLDivideExpr:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case *SQLEqualsExpr:
		v.writeLine("SQLEqualsExpr:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case *SQLExistsExpr:
		v.writeLine("SQLExistsExpr({subquery})")
	case *SQLFieldExpr:
		v.writeLinef("SQLFieldExpr(%s.%s)", typedE.tableName, typedE.fieldName)
	case *SQLGreaterThanExpr:
		v.writeLine("SQLGreaterThanExpr:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case *SQLGreaterThanOrEqualExpr:
		v.writeLine("SQLGreaterThanOrEqualExpr:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case *SQLInExpr:
		v.writeLine("SQLInExpr:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case *SQLLessThanExpr:
		v.writeLine("SQLLessThanExpr:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case *SQLLessThanOrEqualExpr:
		v.writeLine("SQLLessThanOrEqualExpr:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case *SQLLikeExpr:
		v.writeLine("SQLLikeExpr:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case *SQLMultiplyExpr:
		v.writeLine("SQLMultiplyExpr:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case *SQLNotExpr:
		v.writeLine("SQLNotExpr:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case *SQLNotEqualsExpr:
		v.writeLine("SQLNotEqualsExpr:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case *SQLNullCmpExpr:
		v.writeLine("SQLNullCmpExpr:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case *SQLOrExpr:
		v.writeLine("SQLOrExpr:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case *SQLScalarFunctionExpr:
		v.writeLine("SQLScalarFunctionExpr({func})")
	case *SQLSubqueryCmpExpr:
		v.writeLine("SQLSubqueryCmpExpr:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case *SQLSubqueryExpr:
		v.writeLine("SQLSubqueryExpr({subquery})")
	case *SQLSubtractExpr:
		v.writeLine("SQLSubtractExpr:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case *SQLUnaryMinusExpr:
		v.writeLine("SQLUnaryMinusExpr:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case *SQLUnaryPlusExpr:
		v.writeLine("SQLUnaryPlusExpr:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case *SQLUnaryTildeExpr:
		v.writeLine("SQLUnaryTildeExpr:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case *SQLTupleExpr:
		v.writeLine("SQLTupleExpr:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--

	// values
	case SQLBool:
		v.writeLinef("SQLBool(%v)", bool(typedE))
	case SQLDate:
		v.writeLinef("SQLDate")
	case SQLDateTime:
		v.writeLinef("SQLDateTime")
	case SQLFloat:
		v.writeLinef("SQLFloat(%v)", float64(typedE))
	case SQLInt:
		v.writeLinef("SQLInt(%v)", int(typedE))
	case SQLNullValue:
		v.writeLinef("SQLNull")
	case SQLString:
		v.writeLinef("SQLString(%v)", string(typedE))
	case SQLTime:
		v.writeLinef("SQLTime")
	case SQLTimestamp:
		v.writeLinef("SQLTimestamp")
	case SQLValues:
		v.writeLine("SQLValues:")
		v.indentLevel++
		walk(v, e)
		v.indentLevel--
	case SQLUint32:
		v.writeLinef("SQLUint32(%v)", uint32(typedE))
	default:
		return nil, fmt.Errorf("unhandled expression %T", typedE)
	}
	return e, nil
}
