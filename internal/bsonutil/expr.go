package bsonutil

import "github.com/10gen/mongo-go-driver/bson"

//
// Expression Translation Wrappers
//
const (
	OpAbs              = "$abs"
	OpAcos             = "$acos"
	OpAcosh            = "$acosh"
	OpAdd              = "$add"
	OpAddToSet         = "$addToSet"
	OpAllElementsTrue  = "$allElementsTrue"
	OpAnd              = "$and"
	OpAnyElementTrue   = "$anyElementTrue"
	OpArrElemAt        = "$arrayElemAt"
	OpAsin             = "$asin"
	OpAsinh            = "$asinh"
	OpAtan             = "$atan"
	OpAtanh            = "$atanh"
	OpAtan2            = "$atan2"
	OpAvg              = "$avg"
	OpCeil             = "$ceil"
	OpConcat           = "$concat"
	OpCond             = "$cond"
	OpConvert          = "$convert"
	OpCos              = "$cos"
	OpCosh             = "$cosh"
	OpDateFromParts    = "$dateFromParts"
	OpDateFromString   = "$dateFromString"
	OpDateToString     = "$dateToString"
	OpDayOfMonth       = "$dayOfMonth"
	OpDayOfWeek        = "$dayOfWeek"
	OpDayOfYear        = "$dayOfYear"
	OpDegreesToRadians = "$degreesToRadians"
	OpDivide           = "$divide"
	OpEq               = "$eq"
	OpExp              = "$exp"
	OpExists           = "$exists"
	OpFilter           = "$filter"
	OpFirst            = "$first"
	OpFloor            = "$floor"
	OpGt               = "$gt"
	OpGte              = "$gte"
	OpHour             = "$hour"
	OpIfNull           = "$ifNull"
	OpIn               = "$in"
	OpIndexOfCP        = "$indexOfCP"
	OpIsArray          = "$isArray"
	OpLt               = "$lt"
	OpLte              = "$lte"
	OpNaturalLog       = "$ln"
	OpLog              = "$log"
	OpLTrim            = "$ltrim"
	OpMap              = "$map"
	OpMax              = "$max"
	OpMin              = "$min"
	OpMinute           = "$minute"
	OpMillisecond      = "$millisecond"
	OpMod              = "$mod"
	OpMonth            = "$month"
	OpMultiply         = "$multiply"
	OpNeq              = "$ne"
	OpNotIn            = "$nin"
	OpNor              = "$nor"
	OpNot              = "$not"
	OpOr               = "$or"
	OpPow              = "$pow"
	OpPush             = "$push"
	OpRadiansToDegrees = "$radiansToDegrees"
	OpRange            = "$range"
	OpReduce           = "$reduce"
	OpRegex            = "$regex"
	OpRegexOptions     = "$options"
	OpReverseArray     = "$reverseArray"
	OpRTrim            = "$rtrim"
	OpSecond           = "$second"
	OpSin              = "$sin"
	OpSinh             = "$sinh"
	OpSize             = "$size"
	OpSlice            = "$slice"
	OpSplit            = "$split"
	OpSqrt             = "$sqrt"
	OpStdDevPop        = "$stdDevPop"
	OpStdDevSamp       = "$stdDevSamp"
	OpStrLenBytes      = "$strLenBytes"
	OpStrlenCP         = "$strLenCP"
	OpSubstr           = "$substrCP"
	OpSubtract         = "$subtract"
	OpSum              = "$sum"
	OpSwitch           = "$switch"
	OpTan              = "$tan"
	OpTanh             = "$tanh"
	OpToLower          = "$toLower"
	OpToUpper          = "$toUpper"
	OpTrim             = "$trim"
	OpTrunc            = "$trunc"
	OpType             = "$type"
	OpWeek             = "$week"
	OpYear             = "$year"
	OpZip              = "$zip"
)

// WrapInBinOp builds an expression that evaluates a two argument operator
// on the two passed argument expressions.
func WrapInBinOp(op string, left, right interface{}) bson.M {
	return NewM(NewDocElem(op, NewArray(left, right)))
}

// WrapInCond returns a document that evalutes to truePart
// if any of conds is true, and falsePart otherwise.
func WrapInCond(truePart, falsePart interface{}, conds ...interface{}) interface{} {
	var condition interface{}

	switch len(conds) {
	case 0:
		return falsePart
	case 1:
		condition = conds[0]
	default:
		condition = WrapInOp(OpOr, conds...)
	}

	if condition == true {
		return truePart
	}

	if condition == false {
		return falsePart
	}

	return WrapInOp(OpCond, condition, truePart, falsePart)
}

// WrapInOp returns a document which passes all arguments to the op.
func WrapInOp(op string, args ...interface{}) interface{} {
	return NewM(NewDocElem(op, args))
}

// WrapInType wraps the passed expression in an expression
// that returns the type of the expression.
func WrapInType(v interface{}) bson.M {
	return NewM(NewDocElem(OpType, v))
}
