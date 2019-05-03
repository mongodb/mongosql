package astprint

import (
	"encoding/base64"
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/10gen/mongoast/internal/jsonutil"
	"github.com/10gen/mongoast/parser"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// Copied from go driver.
type sortableString []rune

func (ss sortableString) Len() int {
	return len(ss)
}

func (ss sortableString) Less(i, j int) bool {
	return ss[i] < ss[j]
}

func (ss sortableString) Swap(i, j int) {
	oldI := ss[i]
	ss[i] = ss[j]
	ss[j] = oldI
}

func sortStringAlphebeticAscending(s string) string {
	ss := sortableString([]rune(s))
	sort.Sort(ss)
	return string([]rune(ss))
}

// ShellPrintConstant prints a constant with shell formatting.
func ShellPrintConstant(w io.Writer, v bsoncore.Value) {
	switch v.Type {
	case bsontype.Double:
		f64 := v.Double()
		fmt.Fprintf(w, `%v`, formatDouble(f64))
	case bsontype.String:
		str := v.StringValue()
		fmt.Fprint(w, escapeString(str))
	case bsontype.EmbeddedDocument:
		doc, err := v.Document().Elements()
		if err != nil {
			panic(err)
		}
		if len(doc) == 0 {
			fmt.Fprint(w, "{}")
			return
		}
		fmt.Fprint(w, "{")
		l := len(doc) - 1
		for i := 0; i < l; i++ {
			fmt.Fprintf(w, `%s: `, escapeString(doc[i].Key()))
			ShellPrintConstant(w, doc[i].Value())
			fmt.Fprint(w, ",")
		}
		fmt.Fprintf(w, `%s: `, escapeString(doc[l].Key()))
		ShellPrintConstant(w, doc[l].Value())
		fmt.Fprint(w, "}")

	case bsontype.Array:
		arr, err := v.Array().Elements()
		if err != nil {
			panic(err)
		}
		if len(arr) == 0 {
			fmt.Fprint(w, "[]")
			return
		}
		fmt.Fprint(w, "[")
		l := len(arr) - 1
		for i := 0; i < l; i++ {
			ShellPrintConstant(w, arr[i].Value())
			fmt.Fprint(w, ",")
		}
		ShellPrintConstant(w, arr[l].Value())
		fmt.Fprint(w, "]")

	case bsontype.Binary:
		subtype, data := v.Binary()
		fmt.Fprintf(w, `BinData(%d, "%s")`, subtype, base64.StdEncoding.EncodeToString(data))
	case bsontype.Undefined:
		fmt.Fprint(w, `undefined`)
	case bsontype.ObjectID:
		oid := v.ObjectID()
		fmt.Fprintf(w, `ObjectId("%s")`, oid.Hex())
	case bsontype.Boolean:
		b := v.Boolean()
		fmt.Fprintf(w, strconv.FormatBool(b))
	case bsontype.DateTime:
		dt := v.Time().UTC()
		fmt.Fprintf(w, `ISODate("%s")`, dt.Format("2006-01-02T15:04:05.999999"))
	case bsontype.Null:
		fmt.Fprint(w, "null")
	case bsontype.Regex:
		pattern, options := v.Regex()
		fmt.Fprintf(w,
			`/%s/%s`,
			escapeRegex(pattern), sortStringAlphebeticAscending(options),
		)
	case bsontype.DBPointer:
		dbp, oid := v.DBPointer()
		fmt.Fprintf(w, `DBPointer("%s", ObjectId("%s"))`, dbp, oid.Hex())
	case bsontype.JavaScript:
		code := v.JavaScript()
		fmt.Fprintf(w, `Code("%s")`, code)
	case bsontype.Symbol:
		sym := v.Symbol()
		fmt.Fprintf(w, `Symbol("%s")`, sym)
	case bsontype.CodeWithScope:
		code, scope := v.CodeWithScope()
		fmt.Fprintf(w, `Code("%s", `, code)
		// This is... circuitous.
		scopeStr := scope.String()
		scopeValue := jsonutil.ParseJSON(scopeStr)
		parsedScope, err := parser.ParseExpr(scopeValue)
		if err != nil {
			panic(fmt.Sprintf("failed to parse scope document: '%s'", scopeStr))
		}
		ShellPrint(w, parsedScope)
		fmt.Fprintf(w, ")")
	case bsontype.Int32:
		i32 := v.Int32()
		fmt.Fprintf(w, `NumberInt("%d")`, i32)
	case bsontype.Timestamp:
		t, i := v.Timestamp()
		fmt.Fprintf(w, `Timestamp(%s,%s)`, strconv.FormatUint(uint64(t), 10), strconv.FormatUint(uint64(i), 10))
	case bsontype.Int64:
		i64 := v.Int64()
		fmt.Fprintf(w, `NumberLong("%d")`, i64)
	case bsontype.Decimal128:
		d128 := v.Decimal128()
		fmt.Fprintf(w, `NumberDecimal("%s")`, d128.String())
	case bsontype.MinKey:
		fmt.Fprint(w, `MinKey`)
	case bsontype.MaxKey:
		fmt.Fprint(w, `MaxKey`)
	default:
		panic(fmt.Sprintf("not currently supporting %s for Mongo Shell printing", v.Type))
	}
}
