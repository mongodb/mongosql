package astutil

import (
	"math"
	"strings"

	"github.com/10gen/mongoast/ast"

	"github.com/10gen/sqlproxy/internal/bsonutil"

	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// Expression Translation Literals
var (
	NullLiteral            = NullValue()
	TrueLiteral            = BooleanValue(true)
	FalseLiteral           = BooleanValue(false)
	ZeroInt32Literal       = Int32Value(0)
	OneInt32Literal        = Int32Value(1)
	PiLiteral              = FloatValue(math.Pi)
	PiOverTwoLiteral       = FloatValue(math.Pi / 2)
	EmptyStringLiteral     = StringValue("")
	ThisVarRef             = ast.NewVariableRef("this")
	ValueVarRef            = ast.NewVariableRef("value")
	DateComponentSeparator = ast.NewArray(
		StringValue("!"), StringValue("\""), StringValue("#"),
		StringConstant("$"), StringValue("%"), StringValue("&"),
		StringValue("'"), StringValue("("), StringValue(")"),
		StringValue("*"), StringValue("+"), StringValue(","),
		StringValue("-"), StringValue("."), StringValue("/"),
		StringValue(":"), StringValue(";"), StringValue("<"),
		StringValue("="), StringValue(">"), StringValue("?"),
		StringValue("@"), StringValue("["), StringValue("\\"),
		StringValue("]"), StringValue("^"), StringValue("_"),
		StringValue("`"), StringValue("{"), StringValue("|"),
		StringValue("}"), StringValue("~"),
	)
)

const (
	// millisecondsPerDay is the number of milliseconds in a day.
	millisecondsPerDay = 8.64e+7
)

// factorial is an array giving the factorial of 0 <= n <= 15.
var factorial = []float64{
	1.0,
	1.0,
	2.0,
	6.0,
	24.0,
	120.0,
	720.0,
	5040.0,
	40320.0,
	362880.0,
	3628800.0,
	39916800.0,
	479001600.0,
	6227020800.0,
	87178291200.0,
	1307674368000.0,
}

// FieldRefFromFieldName creates an ast.FieldRef from a fieldName that
// possibly includes "."s. See FieldRefFromFieldNameWithParent for more
// details and examples.
func FieldRefFromFieldName(fieldName string) *ast.FieldRef {
	return FieldRefFromFieldNameWithParent(fieldName, nil)
}

// FieldRefFromFieldNameWithParent creates an ast.FieldRef from a fieldName
// that possibly includes "."s. Each name preceding a "." is the parent of
// the following ref; the top-level parent (which is most deeply nested) is
// provided as an argument to this function. It can be nil.
// For example:
//   FieldRefFromFieldNameWithParent("c.b.a", nil) returns
//   &ast.FieldRef{
//     Name: "a",
//     Parent: &ast.FieldRef{
//               Name: "b",
//               Parent: &ast.FieldRef{Name: "c", Parent: nil}
//             }
//   }
//
// and
//
//   FieldRefFromFieldNameWithParent("b.a", ast.NewVariableRef("this") returns
//   &ast.FieldRef{
//     Name: "a",
//     Parent: &ast.FieldRef{
//               Name: "b",
//               Parent: &ast.VariableRef{Name: "this"}
//             }
//   }
func FieldRefFromFieldNameWithParent(fieldName string, parent ast.Expr) *ast.FieldRef {
	parts := strings.Split(fieldName, ".")

	ref := parent
	for _, part := range parts {
		ref = ast.NewFieldRef(part, ref)
	}

	return ref.(*ast.FieldRef)
}

// WrapInAcosComputation wraps the argument in an expression
// that computes the arccos of the argument.
func WrapInAcosComputation(expr ast.Expr) ast.Expr {
	var input ast.Expr
	inputLetAssignment := make([]*ast.LetVariable, 0, 1)

	// If given a constant or a variable or field reference, there is no
	// need to bind it in a $let.
	switch expr.(type) {
	case *ast.VariableRef, *ast.FieldRef, *ast.Constant:
		input = expr
	default:
		input = ast.NewVariableRef("input")
		inputLetAssignment = append(inputLetAssignment,
			ast.NewLetVariable("input", expr),
		)
	}

	absInput := ast.NewVariableRef("absInput")
	absInputLetAssignment := []*ast.LetVariable{
		ast.NewLetVariable("absInput", ast.NewFunction(bsonutil.OpAbs, input)),
	}

	// The power series for arccos does not converge well, so instead use
	// this function: from the Handbook of Mathematical Functions, by
	// Milton Abramowitz and Irene Stegun: arccos(x)=sqrt(1-x) *
	// (a0+a1∗x+a2∗x2+a3∗x3). This function is only good far away from -1,
	// so we just mirror the function for negative values by subtracting
	// from Pi (the value of acos(-1)). The constants a0-a3 are defined as
	// follows:
	a0 := FloatValue(1.5707288)
	a1 := FloatValue(-0.2121144)
	a2 := FloatValue(0.0742610)
	a3 := FloatValue(-0.0187293)

	firstTerm := ast.NewFunction(bsonutil.OpSqrt, ast.NewBinary(bsonutil.OpSubtract, FloatValue(1.0), absInput))
	secondTerm := WrapInOp(bsonutil.OpAdd,
		a0,
		ast.NewBinary(bsonutil.OpMultiply, a1, absInput),
		ast.NewBinary(bsonutil.OpMultiply, a2, ast.NewBinary(bsonutil.OpPow, absInput, Int32Value(2))),
		ast.NewBinary(bsonutil.OpMultiply, a3, ast.NewBinary(bsonutil.OpPow, absInput, Int32Value(3))),
	)

	evaluation := ast.NewLet(absInputLetAssignment,
		WrapInCond(
			ast.NewBinary(bsonutil.OpMultiply, firstTerm, secondTerm),
			ast.NewBinary(bsonutil.OpSubtract,
				PiLiteral,
				ast.NewBinary(bsonutil.OpMultiply,
					firstTerm,
					secondTerm)),
			ast.NewBinary(bsonutil.OpGte, input, ZeroInt32Literal),
		),
	)

	// If there was no input $let binding, just return the evaluation.
	if inputLetAssignment == nil {
		return evaluation
	}

	return ast.NewLet(inputLetAssignment, evaluation)
}

// WrapInCase returns an expression to use as one of the branches arguments to WrapInSwitch.
// caseExpr must evaluate to a boolean.
func WrapInCase(caseExpr, thenExpr ast.Expr) *ast.Document {
	return ast.NewDocument(
		ast.NewDocumentElement("case", caseExpr),
		ast.NewDocumentElement("then", thenExpr),
	)
}

// WrapInConcat returns the aggregation expression
// {$concat: [expr1, expr2, ...]}
// https://docs.mongodb.com/manual/reference/operator/aggregation/concat/
func WrapInConcat(exprs []ast.Expr) *ast.Function {
	return WrapInOp(bsonutil.OpConcat, exprs...)
}

// WrapInCond returns a document that evalutes to truePart
// if any of conds is true, and falsePart otherwise.
func WrapInCond(truePart, falsePart ast.Expr, conds ...ast.Expr) ast.Expr {
	var condition ast.Expr

	switch len(conds) {
	case 0:
		return falsePart
	case 1:
		condition = conds[0]
	default:
		condition = WrapInOp(bsonutil.OpOr, conds...)
	}

	if condition == TrueLiteral {
		return truePart
	}

	if condition == FalseLiteral {
		return falsePart
	}

	return WrapInOp(bsonutil.OpCond, condition, truePart, falsePart)
}

// WrapInConvert takes input and wraps it in a $convert operation naively, without
// accounting for all special conversions needed to reflect mySQL behavior. DO NOT USE this
// function to convert directly; instead, call evaluator/translateConvert for a correct answer in all cases.
func WrapInConvert(input ast.Expr, to string, onError, onNull ast.Expr) ast.Expr {
	return ast.NewFunction(bsonutil.OpConvert, ast.NewDocument(
		ast.NewDocumentElement("input", input),
		ast.NewDocumentElement("to", StringValue(to)),
		ast.NewDocumentElement("onError", onError),
		ast.NewDocumentElement("onNull", onNull),
	))
}

// WrapInCosPowerSeries wraps the argument in an expression that computes the
// cos Maclaurin power series of the argument, expr.
// http://mathworld.wolfram.com/MaclaurinSeries.html
func WrapInCosPowerSeries(expr ast.Expr) *ast.Let {
	input := ast.NewVariableRef("input")
	inputLetAssignment := []*ast.LetVariable{
		ast.NewLetVariable("input", expr),
	}

	return ast.NewLet(inputLetAssignment,
		WrapInOp(bsonutil.OpAdd,
			OneInt32Literal,
			WrapInPowerSeriesTerm(input, 2),
			WrapInPowerSeriesTerm(input, 4),
			WrapInPowerSeriesTerm(input, 6),
			WrapInPowerSeriesTerm(input, 8),
			WrapInPowerSeriesTerm(input, 10),
			WrapInPowerSeriesTerm(input, 12),
			WrapInPowerSeriesTerm(input, 14),
		),
	)
}

// WrapInDateFormat wraps an Aggregation Expression that evaluates to a date
// in a date_format expression that will use '$dateFromString' to format
// a date to a string.
func WrapInDateFormat(date ast.Expr, mysqlFormat string) (ast.Expr, bool) {
	var format string
	for i := 0; i < len(mysqlFormat); i++ {
		if mysqlFormat[i] == '%' {
			if i != len(mysqlFormat)-1 {
				switch mysqlFormat[i+1] {
				case '%':
					format += "%%"
				case 'd':
					format += "%d"
				case 'f':
					format += "%L000"
				case 'H', 'k':
					format += "%H"
				case 'i':
					format += "%M"
				case 'j':
					format += "%j"
				case 'm':
					format += "%m"
				case 's', 'S':
					format += "%S"
				case 'T':
					format += "%H:%M:%S"
				case 'U':
					format += "%U"
				case 'Y':
					format += "%Y"
				default:
					return nil, false
				}
				i++
			} else {
				// MongoDB fails when the last character is a % sign in the format string.
				return nil, false
			}
		} else {
			format += string(mysqlFormat[i])
		}
	}

	if val, ok := date.(*ast.Constant); ok && val.Value.Type == bsontype.Null {
		return nil, true
	}

	return WrapInNullCheckedCond(
		NullLiteral,
		WrapInDateToString(date, format),
		date,
	), true
}

// WrapInDateFromParts returns a date given the year, month and day passed in.
func WrapInDateFromParts(year, month, dayOfMonth ast.Expr) ast.Expr {
	return ast.NewFunction(bsonutil.OpDateFromParts, ast.NewDocument(
		ast.NewDocumentElement("year", ast.NewFunction(bsonutil.OpYear, year)),
		ast.NewDocumentElement("month", ast.NewFunction(bsonutil.OpMonth, month)),
		ast.NewDocumentElement("day", ast.NewFunction(bsonutil.OpDayOfMonth, dayOfMonth)),
	))
}

// WrapInDateToString converts date to a string according to the specified format.
func WrapInDateToString(date ast.Expr, format string) *ast.Function {
	return ast.NewFunction(bsonutil.OpDateToString, ast.NewDocument(
		ast.NewDocumentElement("date", date),
		ast.NewDocumentElement("format", StringValue(format)),
	))

}

// WrapInEqCase returns a document that is a case arm that checks equality between expr1 and expr2.
func WrapInEqCase(expr1, expr2, thenExpr ast.Expr) ast.Expr {
	caseExpr := ast.NewBinary(bsonutil.OpEq, expr1, expr2)
	return ast.NewDocument(
		ast.NewDocumentElement("case", caseExpr),
		ast.NewDocumentElement("then", thenExpr),
	)
}

// WrapInFilter returns the aggregation expression {$filter: {input: input, as: as, cond: cond }}.
// https://docs.mongodb.com/manual/reference/operator/aggregation/filter/
func WrapInFilter(input ast.Expr, as string, cond ast.Expr) *ast.Function {
	asExpr := ast.NewUnknown(bsoncore.Value{
		Type: bsontype.String,
		Data: bsoncore.AppendString(nil, as),
	})
	return ast.NewFunction(bsonutil.OpFilter, ast.NewDocument(
		ast.NewDocumentElement("input", input),
		ast.NewDocumentElement("as", asExpr),
		ast.NewDocumentElement("cond", cond),
	))
}

// WrapInIfNull returns v if it isn't nil, otherwise, it returns ifNull.
func WrapInIfNull(v, ifNull ast.Expr) ast.Expr {
	if value, ok := v.(*ast.Constant); ok {
		if value.Value.Type == bsontype.Null {
			return ifNull
		}
		return v
	}

	return WrapInOp(bsonutil.OpIfNull, v, ifNull)
}

// WrapInInRange returns an expression that evaluates to true if val is in range [min, max).
// val must evaluate to a number.
func WrapInInRange(val ast.Expr, min, max float64) ast.Expr {
	return ast.NewBinary(bsonutil.OpAnd,
		ast.NewBinary(bsonutil.OpGte, val, FloatValue(min)),
		ast.NewBinary(bsonutil.OpLt, val, FloatValue(max)),
	)
}

// WrapInIntDiv performs an integer division (truncated division).
func WrapInIntDiv(numerator, denominator ast.Expr) ast.Expr {
	return ast.NewFunction(bsonutil.OpTrunc,
		ast.NewBinary(bsonutil.OpDivide, numerator, denominator),
	)
}

// WrapInIsLeapYear creates an expression that returns true if the argument is
// a leap year, and false otherwise. This function assume val is an integer
// year.
func WrapInIsLeapYear(val ast.Expr) *ast.Let {
	v := ast.NewVariableRef("val")
	letAssignment := []*ast.LetVariable{
		ast.NewLetVariable("val", val),
	}

	// This computes the expression:
	// (v % 4 == 0) && (v % 100 != 0) || (v % 400 == 0).
	return ast.NewLet(letAssignment,
		ast.NewBinary(bsonutil.OpOr,
			ast.NewBinary(bsonutil.OpAnd,
				ast.NewBinary(bsonutil.OpEq,
					ast.NewBinary(bsonutil.OpMod, v, Int32Value(4)),
					ZeroInt32Literal),
				ast.NewBinary(bsonutil.OpNeq,
					ast.NewBinary(bsonutil.OpMod, v, Int32Value(100)),
					ZeroInt32Literal),
			),
			ast.NewBinary(bsonutil.OpEq,
				ast.NewBinary(bsonutil.OpMod, v, Int32Value(400)),
				ZeroInt32Literal),
		),
	)
}

// WrapInLRTrim returns a trimmed version of args.
func WrapInLRTrim(isLTrimType bool, args ast.Expr) *ast.Function {
	var (
		splitArray   = WrapInOp(bsonutil.OpSplit, args, StringValue(" "))
		substrIndex  ast.Expr
		substrLength ast.Expr
	)

	if !isLTrimType {
		splitArray = ast.NewFunction(bsonutil.OpReverseArray, splitArray)
	}

	splitArrayVarRef := ast.NewVariableRef("splitArray")
	letAssignment := []*ast.LetVariable{
		ast.NewLetVariable("splitArray", splitArray),
	}

	letEvaluation := ast.NewFunction(bsonutil.OpZip, ast.NewDocument(
		ast.NewDocumentElement("inputs", ast.NewArray(
			splitArrayVarRef,
			WrapInOp(bsonutil.OpRange, ZeroInt32Literal, ast.NewFunction(bsonutil.OpSize, splitArrayVarRef)),
		)),
	))

	mapInput := ast.NewLet(letAssignment, letEvaluation)

	zipArrayVarRef := ast.NewVariableRef("zipArray")
	mapIn := WrapInCond(
		ast.NewFunction(bsonutil.OpStrlenCP, args),
		WrapInOp(bsonutil.OpArrElemAt, zipArrayVarRef, OneInt32Literal),
		ast.NewBinary(
			bsonutil.OpEq,
			WrapInOp(bsonutil.OpArrElemAt, zipArrayVarRef, ZeroInt32Literal),
			StringValue(""),
		),
	)

	min := ast.NewFunction(bsonutil.OpMin, WrapInMap(mapInput, "zipArray", mapIn))

	if isLTrimType {
		substrIndex = min
		substrLength = ast.NewFunction(bsonutil.OpStrlenCP, args)
	} else {
		substrIndex = ZeroInt32Literal
		substrLength = ast.NewBinary(bsonutil.OpSubtract, ast.NewFunction(bsonutil.OpStrlenCP, args), min)
	}

	return WrapInOp(bsonutil.OpSubstr, args, substrIndex, substrLength)
}

// WrapInMap returns the aggregation expression {$map: {input: input, as: as, in: in }}.
// https://docs.mongodb.com/manual/reference/operator/aggregation/map/
func WrapInMap(input ast.Expr, as string, in ast.Expr) *ast.Function {
	asExpr := ast.NewUnknown(bsoncore.Value{
		Type: bsontype.String,
		Data: bsoncore.AppendString(nil, as),
	})
	return ast.NewFunction(bsonutil.OpMap, ast.NewDocument(
		ast.NewDocumentElement("input", input),
		ast.NewDocumentElement("as", asExpr),
		ast.NewDocumentElement("in", in),
	))
}

func bsonValueIsNull(value bsoncore.Value) *ast.Unknown {
	if value.Type == bsontype.Null {
		return TrueLiteral
	}
	return FalseLiteral
}

// WrapInNullCheck returns true if v is null, false otherwise.
func WrapInNullCheck(v ast.Expr) ast.Expr {
	// Constants and Unknown simply wrap bsoncore.Values; if those values
	// are known to be null, there is no need to wrap in a Binary operation.
	switch t := v.(type) {
	case *ast.Constant:
		return bsonValueIsNull(t.Value)
	case *ast.Unknown:
		return bsonValueIsNull(t.Value)
	}

	return ast.NewBinary(bsonutil.OpLte, v, NullLiteral)
}

// WrapInNullCheckedCond returns a document that evaluates to truePart
// if any of the null checked conds is true, and falsePart otherwise.
func WrapInNullCheckedCond(truePart, falsePart ast.Expr, conds ...ast.Expr) ast.Expr {
	var condition ast.Expr
	newConds := make([]ast.Expr, 0, len(conds))
	for _, cond := range conds {
		unknown, isUnknown := cond.(*ast.Unknown)
		constant, isConstant := cond.(*ast.Constant)
		if (isUnknown && unknown.Value.Type == bsontype.Null) ||
			(isConstant && constant.Value.Type == bsontype.Null) {
			return truePart
		}
		newConds = append(newConds, WrapInNullCheck(cond))
	}
	switch len(newConds) {
	case 0:
		return falsePart
	case 1:
		condition = newConds[0]
	default:
		condition = WrapInOp(bsonutil.OpOr, newConds...)
	}

	return WrapInOp(bsonutil.OpCond, condition, truePart, falsePart)
}

// WrapInOp returns a document which passes all arguments to the op.
func WrapInOp(op string, args ...ast.Expr) *ast.Function {
	return ast.NewFunction(op, ast.NewArray(args...))
}

// WrapInPowerSeriesTerm takes an input and a power and produces the power
// series term for that integer as a MongoDB aggregration expression that is
// defined as input^power/ factorial(power).
func WrapInPowerSeriesTerm(input ast.Expr, power uint32) *ast.Binary {
	ret := ast.NewBinary(bsonutil.OpDivide,
		ast.NewBinary(bsonutil.OpPow, input, Int64Value(int64(power))),
		FloatValue(factorial[power]),
	)
	pmod4 := power % 4
	// powers that are equal to 3 or 2 modulo 4 are negative in the Cos and
	// Sine series.
	if pmod4 == 3 || pmod4 == 2 {
		return ast.NewBinary(bsonutil.OpMultiply, FloatValue(-1.0), ret)
	}
	return ret
}

// WrapInRange returns the aggregation expression {$range: [start, stop, step]}.
// https://docs.mongodb.com/manual/reference/operator/aggregation/range/
func WrapInRange(start, stop, step ast.Expr) *ast.Function {
	if step == nil {
		return WrapInOp(bsonutil.OpRange, start, stop)
	}
	return WrapInOp(bsonutil.OpRange, start, stop, step)
}

// WrapInReduce returns the aggregation expression
// {$reduce: {input: input, initialValue: initialValue, in: in }}.
// https://docs.mongodb.com/manual/reference/operator/aggregation/range/
func WrapInReduce(input, initialValue, in ast.Expr) *ast.Function {
	return ast.NewFunction(bsonutil.OpReduce, ast.NewDocument(
		ast.NewDocumentElement("input", input),
		ast.NewDocumentElement("initialValue", initialValue),
		ast.NewDocumentElement("in", in),
	))
}

// WrapInRegex returns a regex document for use in match expressions.
// https://docs.mongodb.com/manual/reference/operator/query/regex/
func WrapInRegex(pattern, options string) *ast.Document {
	return ast.NewDocument(
		ast.NewDocumentElement(bsonutil.OpRegex, StringValue(pattern)),
		ast.NewDocumentElement(bsonutil.OpRegexOptions, StringValue(options)),
	)
}

// WrapInRound generates an expression to round a floating point number
// the way MySQL does. This is the simplest implementation of round I have found:
// https://github.com/golang/go/issues/4594#issuecomment-66073312.
func WrapInRound(val ast.Expr) ast.Expr {
	// The MongoDB aggregation language generated by this function for
	// non-literal values implements the following algorithm presented in go
	// code:
	// if x < 0 {
	//      return math.Ceil(x-.5)
	// }
	// return math.Floor(x+.5)
	// For literal values, the rounding is done in-memory, in Go.
	switch t := val.(type) {
	case *ast.Constant:
		switch t.Value.Type {
		case bsontype.Int64, bsontype.Int32:
			return t
		case bsontype.Double:
			return Int64Value(int64(math.Round(t.Value.Double())))
		}
	}

	condExpr := ast.NewBinary(bsonutil.OpLt, val, FloatValue(0.0))
	lt0 := ast.NewFunction(bsonutil.OpCeil, ast.NewBinary(bsonutil.OpSubtract, val, FloatValue(0.5)))
	gte0 := ast.NewFunction(bsonutil.OpFloor, ast.NewBinary(bsonutil.OpAdd, val, FloatValue(0.5)))
	return WrapInCond(lt0, gte0, condExpr)
}

// WrapInRoundWithPrecision returns arg rounded to placeVal places.
func WrapInRoundWithPrecision(arg ast.Expr, placeVal float64) ast.Expr {
	decimal := math.Pow(float64(10), placeVal)
	if decimal < 1 {
		return ZeroInt32Literal
	}

	decimalVarRef := ast.NewVariableRef("decimal")
	letAssignment := []*ast.LetVariable{
		ast.NewLetVariable("decimal", FloatValue(decimal)),
	}

	condExpr := ast.NewBinary(bsonutil.OpGte, arg, ZeroInt32Literal)
	lt0 := ast.NewFunction(bsonutil.OpCeil,
		ast.NewBinary(bsonutil.OpSubtract,
			ast.NewBinary(bsonutil.OpMultiply, arg, decimalVarRef),
			FloatValue(0.5),
		),
	)
	gte0 := ast.NewFunction(bsonutil.OpFloor,
		ast.NewBinary(bsonutil.OpAdd,
			ast.NewBinary(bsonutil.OpMultiply, arg, decimalVarRef),
			FloatValue(0.5),
		),
	)

	letEvaluation := ast.NewBinary(bsonutil.OpDivide, WrapInCond(gte0, lt0, condExpr), decimalVarRef)

	return ast.NewLet(letAssignment, letEvaluation)
}

// WrapInSinPowerSeries wraps the argument in an expression that computes the
// sin Maclaurin power series of the argument, expr.
// http://mathworld.wolfram.com/MaclaurinSeries.html
func WrapInSinPowerSeries(expr ast.Expr) *ast.Let {
	input := ast.NewVariableRef("input")
	inputLetAssignment := []*ast.LetVariable{
		ast.NewLetVariable("input", expr),
	}

	return ast.NewLet(inputLetAssignment,
		WrapInOp(bsonutil.OpAdd,
			input,
			WrapInPowerSeriesTerm(input, 3),
			WrapInPowerSeriesTerm(input, 5),
			WrapInPowerSeriesTerm(input, 7),
			WrapInPowerSeriesTerm(input, 9),
			WrapInPowerSeriesTerm(input, 11),
			WrapInPowerSeriesTerm(input, 13),
			WrapInPowerSeriesTerm(input, 15),
		),
	)
}

// WrapInStringToArray converts an expression v (which must evaluate to a string)
// to an array e.g. "hello" -> ["h", "e", "l", "l", "o"] and returns the array.
func WrapInStringToArray(v ast.Expr) *ast.Function {
	input := WrapInOp(bsonutil.OpRange, ZeroInt32Literal, ast.NewFunction(bsonutil.OpStrlenCP, v))
	in := WrapInOp(bsonutil.OpSubstr, v, ThisVarRef, OneInt32Literal)

	return ast.NewFunction(bsonutil.OpMap, ast.NewDocument(
		ast.NewDocumentElement("input", input),
		ast.NewDocumentElement("in", in),
	))
}

// WrapInSwitch returns the aggregation expression
// {$switch: branches: branches, default: defaultExpr }
// https://docs.mongodb.com/manual/reference/operator/aggregation/switch/
func WrapInSwitch(defaultExpr ast.Expr, branches ...ast.Expr) *ast.Function {
	return ast.NewFunction(bsonutil.OpSwitch, ast.NewDocument(
		ast.NewDocumentElement("branches", ast.NewArray(branches...)),
		ast.NewDocumentElement("default", defaultExpr),
	))
}

// WrapInType wraps the passed expression in an expression
// that returns the type of the expression.
func WrapInType(v ast.Expr) *ast.Function {
	return ast.NewFunction(bsonutil.OpType, v)
}

// WrapInWeekCalculation calculates the week of a given date based on the
// passed argument, expr, which is some MongoDB Aggregation Pipeline
// expression, and the mode, which is an integer.
func WrapInWeekCalculation(expr ast.Expr, mode int64) *ast.Let {
	date, year := ast.NewVariableRef("date"), ast.NewVariableRef("year")
	getJan1 := func() *ast.Function {
		return ast.NewFunction(bsonutil.OpDateFromParts, ast.NewDocument(
			ast.NewDocumentElement("year", year),
			ast.NewDocumentElement("month", OneInt32Literal),
			ast.NewDocumentElement("day", OneInt32Literal),
		))
	}

	getNextJan1 := func() *ast.Function {
		return ast.NewFunction(bsonutil.OpDateFromParts, ast.NewDocument(
			ast.NewDocumentElement("year", ast.NewBinary(bsonutil.OpAdd, year, OneInt32Literal)),
			ast.NewDocumentElement("month", OneInt32Literal),
			ast.NewDocumentElement("day", OneInt32Literal),
		))
	}

	// generateDaySubtract generates the main week calculation shared
	// by all modes except 0, 2 (since those can use MongoDB's week function).
	// The calculation is:
	// trunc((date - dayOne) / (7 * millisecondsPerDay) + 1).
	generateDaySubtract := func(dayOne ast.Expr) *ast.Function {
		return ast.NewFunction(bsonutil.OpTrunc,
			ast.NewBinary(bsonutil.OpAdd,
				OneInt32Literal,
				ast.NewBinary(bsonutil.OpDivide,
					ast.NewBinary(bsonutil.OpSubtract, date, dayOne),
					FloatValue(7*millisecondsPerDay),
				),
			),
		)
	}

	// generate4DaysBody generates the body for modes where the first
	// week is defined by having 4 days in the year, these are modes
	// 1, 3, 4, and 6.
	generate4DaysBody := func(diffConstant int64) *ast.Let {
		// This description is used for Monday as first day of the
		// week. See below for an explanation of the Sunday first day
		// case. Calculate the first day of the first week of this
		// year based on the dayOfWeek of YYYY-01-01 of this year, note
		// that it may be from the previous year. The Day Diff column
		// is the
		// amount of days to Add or Subtract from YYYY-01-01:
		// Day Of Week for Jan 1   |   Day Diff
		// ---------------------------------------------
		//                     1   |   + 1
		//                     2   |   + 0
		//                     3   |   - 1
		//                     4   |   - 2
		//                     5   |   - 3
		//                     6   |   + 3
		//                     7   |   + 2
		// This can be simplified to:
		// diff = -x + 2
		// if diff > -4 {
		//      return diff
		// }
		// return diff + 7
		// for the Sunday version of this, we need to use -x + 1
		// instead of -x + 2, and that is the only difference, that is
		// the point of "diffConstant", it will be either 1 or 2.
		jan1 := getJan1()
		jan1DayOfWeek := ast.NewVariableRef("jan1DayOfWeek")
		dayOfWeekLetAssignment := []*ast.LetVariable{
			ast.NewLetVariable("jan1DayOfWeek", ast.NewFunction(bsonutil.OpDayOfWeek, jan1)),
		}

		diffVarRef := ast.NewVariableRef("diff")
		dayOneLetAssignment := []*ast.LetVariable{
			ast.NewLetVariable(
				"diff",
				ast.NewBinary(bsonutil.OpAdd,
					Int64Value(diffConstant),
					ast.NewBinary(bsonutil.OpMultiply, jan1DayOfWeek, Int32Value(-1)),
				),
			),
		}

		dayOneLetEvaluation := WrapInCond(
			diffVarRef,
			ast.NewBinary(bsonutil.OpAdd, diffVarRef, Int32Value(7)),
			ast.NewBinary(bsonutil.OpGt, diffVarRef, Int32Value(-4)),
		)

		dayOne := ast.NewBinary(bsonutil.OpAdd,
			jan1,
			ast.NewBinary(bsonutil.OpMultiply,
				FloatValue(millisecondsPerDay),
				ast.NewLet(dayOneLetAssignment, dayOneLetEvaluation),
			),
		)
		return ast.NewLet(dayOfWeekLetAssignment, generateDaySubtract(dayOne))
	}

	// generateMondayBody generates the body for modes where the first
	// week is defined by having a Monday, these are modes
	// 5 and 7.
	generateMondayBody := func() *ast.Let {
		// These are more simple than the 4 days mode. The diff from Jan1
		// can be defined using (7 - x + 2) % 7.
		jan1 := getJan1()
		jan1DayOfWeek := ast.NewVariableRef("jan1DayOfWeek")
		dayOfWeekLetAssignment := []*ast.LetVariable{
			ast.NewLetVariable("jan1DayOfWeek", ast.NewFunction(bsonutil.OpDayOfWeek, jan1)),
		}

		dayOne := ast.NewBinary(bsonutil.OpAdd,
			jan1,
			ast.NewBinary(bsonutil.OpMultiply,
				FloatValue(millisecondsPerDay),
				ast.NewBinary(bsonutil.OpMod,
					ast.NewBinary(bsonutil.OpAdd,
						Int32Value(2),
						ast.NewBinary(bsonutil.OpSubtract, Int32Value(7), jan1DayOfWeek),
					),
					Int32Value(7),
				),
			),
		)
		return ast.NewLet(dayOfWeekLetAssignment, generateDaySubtract(dayOne))
	}

	// WrapInZeroCheck - half of all modes allow weeks numbers
	// between 0-53, and the other half allow 1-53. To compute the week
	// for modes allowing 1-53, we compute the week for the associated 0-53
	// mode, and if it results in week 0, we return week('(year-1)-12-31'),
	// which will be either 52 or 53 as the 1-53 modes consider such a date
	// as being in the previous year. This means that
	// WrapInWeekCalculation must be recursive, which is why it is
	// separated from the FuncToAggregation for weekFunc. Note that the
	// recursive step is, at most, depth 1, because only used in 1-53
	// modes, but recursively calls with a 0-53 mode.
	WrapInZeroCheck := func(body ast.Expr, m int64) *ast.Let {
		lastDayLastYear := ast.NewFunction(bsonutil.OpDateFromParts, ast.NewDocument(
			ast.NewDocumentElement("year", ast.NewBinary(bsonutil.OpSubtract, year, OneInt32Literal)),
			ast.NewDocumentElement("month", Int32Value(12)),
			ast.NewDocumentElement("day", Int32Value(31)),
		))

		output := ast.NewVariableRef("output")
		letAssignment := []*ast.LetVariable{
			ast.NewLetVariable("output", body),
		}

		return ast.NewLet(letAssignment,
			WrapInCond(
				output,
				WrapInWeekCalculation(lastDayLastYear, m),
				ast.NewBinary(bsonutil.OpNeq, output, ZeroInt32Literal),
			),
		)
	}

	// wrapInFiftyThreeCheck is used to handle cases where the last week of a
	// year may actually map as the first week of the next year. This is
	// only possible in the cases where the first week is defined by having
	// 4 days in the year, and where 0 weeks are not allowed, so that is
	// modes 3 and 6. In these modes it is possible that 12-31, 12-30, and even
	// 12-29 map to week 1 of the next year. This is similar in design to
	// zeroCheck, except that it is only needed in the modes with 4 days
	// used to decide the first week of the month. We only need to check
	// the day if our computeDaySubtract results in week 53, giving us
	// faster common cases. janOneDaysOfWeek are the days of the week
	// for the next Jan-1 that result in one of the last three days
	// of the year potentially mapping to the next year. Note that
	// MongoDB aggregation pipeline numbers days 1-7, with 1 being Sunday.
	wrapInFiftyThreeCheck := func(body ast.Expr, janOneDaysOfWeek ...int32) *ast.Let {
		output, day := ast.NewVariableRef("output"), ast.NewVariableRef("day")
		outputLetAssignment := []*ast.LetVariable{
			ast.NewLetVariable("output", body),
		}
		dayLetAssignment := []*ast.LetVariable{
			ast.NewLetVariable("day", ast.NewFunction(bsonutil.OpDayOfMonth, date)),
		}

		// Day Of Week for Jan 1  |  First Day In December Mapping to Next Year
		// --------------------------------------------------------------------
		// janOneDaysOfWeek[0]    |  29
		// janOneDaysOfWeek[1]    |  30
		// janOneDaysOfWeek[2]    |  31
		nextJan1DayOfWeek := ast.NewFunction(bsonutil.OpDayOfWeek, getNextJan1())
		return ast.NewLet(outputLetAssignment,
			WrapInCond(
				ast.NewLet(dayLetAssignment,
					WrapInSwitch(
						Int32Value(53),
						WrapInEqCase(
							nextJan1DayOfWeek,
							Int32Value(janOneDaysOfWeek[0]),
							WrapInCond(
								OneInt32Literal,
								Int32Value(53),
								ast.NewBinary(bsonutil.OpGte, day, Int32Value(29)),
							),
						),
						WrapInEqCase(
							nextJan1DayOfWeek,
							Int32Value(janOneDaysOfWeek[1]),
							WrapInCond(
								OneInt32Literal,
								Int32Value(53),
								ast.NewBinary(bsonutil.OpGte, day, Int32Value(30)),
							),
						),
						WrapInEqCase(
							nextJan1DayOfWeek,
							Int32Value(janOneDaysOfWeek[2]),
							WrapInCond(
								OneInt32Literal,
								Int32Value(53),
								ast.NewBinary(bsonutil.OpGte, day, Int32Value(31)),
							),
						),
					),
				),
				output,
				ast.NewBinary(bsonutil.OpEq, output, Int32Value(53)),
			),
		)
	}

	var body ast.Expr
	switch mode {
	// First day of week: Sunday, with a Sunday in this year.
	// This is what MongoDB's $week function does, so we use it.
	case 0, 2:
		body = WrapSingleArgFuncWithNullCheck(bsonutil.OpWeek, date)
		if mode == 2 {
			body = WrapInZeroCheck(body, 0)
		}
	// First day of week: Monday, with 4 days in this year.
	case 1, 3:
		body = generate4DaysBody(2)
		if mode == 3 {
			body = WrapInZeroCheck(body, 1)
			body = wrapInFiftyThreeCheck(body, 5, 4, 3)
		}
	// First day of week: Sunday, with 4 days in this year.
	case 4, 6:
		body = generate4DaysBody(1)
		if mode == 6 {
			body = WrapInZeroCheck(body, 4)
			body = wrapInFiftyThreeCheck(body, 4, 3, 2)
		}
	// First day of week: Monday, with a Monday in this year.
	case 5, 7:
		body = generateMondayBody()
		if mode == 7 {
			body = WrapInZeroCheck(body, 5)
		}
	}

	// Bind expressions that would be expensive to recompute.
	dateLetAssignment := []*ast.LetVariable{
		ast.NewLetVariable("date", expr),
	}
	yearLetAssignment := []*ast.LetVariable{
		ast.NewLetVariable("year", ast.NewFunction(bsonutil.OpYear, date)),
	}
	return ast.NewLet(dateLetAssignment,
		ast.NewLet(yearLetAssignment, body),
	)
}

// WrapSingleArgFuncWithNullCheck returns a null checked version
// of the arg passed to name.
func WrapSingleArgFuncWithNullCheck(name string, arg ast.Expr) ast.Expr {
	return WrapInNullCheckedCond(NullLiteral, ast.NewFunction(name, arg), arg)
}
