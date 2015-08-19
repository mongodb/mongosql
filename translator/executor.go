package translator

import (
	"fmt"
)

func executor(aType interface{}) (interface{}, error) {

	switch expr := aType.(type) {

	case FieldComp:
		return expr, nil
	case LOJ:
		return expr, nil
	case ROJ:
		return expr, nil
	case FOJ:
		return expr, nil
	case COJ:
		return expr, nil
	case AsExpr:
		return expr, nil
	case UnionAll:
		return expr, nil
	case Union:
		return expr, nil
	case Relation:
		return expr, nil
	case CrltSubquery:
		return expr, nil
	case NumVal:
		return expr, nil
	case ValTuple:
		return expr, nil
	case NullVal:
		return expr, nil
	case ColName:
		return expr, nil
	case StrVal:
		return expr, nil
	case BinaryExpr:
		return expr, nil
	case AndExpr:
		return expr, nil
	case OrExpr:
		return expr, nil
	case ComparisonExpr:
		return expr, nil
	case RangeCond:
		return expr, nil
	case NullCheck:
		return expr, nil
	case UnaryExpr:
		return expr, nil
	case NotExpr:
		return expr, nil
	case ParenBoolExpr:
		return expr, nil
	case Subquery:
		return expr, nil
	case ValArg:
		return expr, nil
	case FuncExpr:
		return expr, nil
	case CaseExpr:
		return expr, nil
	case ExistsExpr:
		return expr, nil
	default:
		return nil, fmt.Errorf("can't handle expression type %T", expr)
	}
}
