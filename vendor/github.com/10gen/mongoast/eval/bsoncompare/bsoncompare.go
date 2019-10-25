package bsoncompare

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/10gen/mongoast/internal/decimalutil"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

// Compare compares two values.
func Compare(a bsoncore.Value, b bsoncore.Value) (int, error) {
	// compareDoc and compareObjectID come first because they are used by other
	// comparators. The rest are in alphabetical order.
	compareDoc := func(a, b []bsoncore.Element) (int, error) {
		aLen := len(a)
		bLen := len(b)

		for i := 0; i < aLen; i++ {
			if i >= bLen {
				return 1, nil
			}

			aElem := a[i]
			bElem := b[i]

			if aElem.Key() != bElem.Key() {
				return strings.Compare(aElem.Key(), bElem.Key()), nil
			}

			res, err := Compare(aElem.Value(), bElem.Value())
			if res != 0 || err != nil {
				return res, err
			}
		}

		if aLen < bLen {
			return -1, nil
		}
		return 0, nil
	}

	compareObjectID := func(a, b primitive.ObjectID) int {
		return bytes.Compare(a[:], b[:])
	}

	compareBinary := func(aSubtype byte, aData []byte, bSubtype byte, bData []byte) int {
		if len(aData) < len(bData) {
			return -1
		} else if len(aData) > len(bData) {
			return 1
		}
		return bytes.Compare(append(aData, aSubtype), append(bData, bSubtype))
	}

	compareBool := func(a, b bool) int {
		if a == b {
			return 0
		}
		return -1
	}

	compareCodeWithScope := func(aCode string, aScope bsoncore.Document, bCode string, bScope bsoncore.Document) (int, error) {
		res := strings.Compare(aCode, bCode)
		if res != 0 {
			return res, nil
		}
		aElems, _ := aScope.Elements()
		bElems, _ := bScope.Elements()
		return compareDoc(aElems, bElems)
	}

	compareDBPointer := func(ans string, aid primitive.ObjectID, bns string, bid primitive.ObjectID) int {
		if len(ans) < len(bns) {
			return -1
		} else if len(ans) > len(bns) {
			return 1
		}
		res := strings.Compare(ans, bns)
		if res != 0 {
			return res
		}
		return compareObjectID(aid, bid)
	}

	compareDouble := func(a, b float64) int {
		if a < b {
			return -1
		}
		if a == b {
			return 0
		}
		return 1
	}

	compareInt64 := func(a, b int64) int {
		if a < b {
			return -1
		}
		if a == b {
			return 0
		}
		return 1
	}

	compareRegex := func(aPattern, aOptions, bPattern, bOptions string) int {
		res := strings.Compare(aPattern, bPattern)
		if res != 0 {
			return res
		}
		return strings.Compare(aOptions, bOptions)
	}

	compareTimestamp := func(at, ai, bt, bi uint32) int {
		if at < bt {
			return -1
		} else if at > bt {
			return 1
		} else if ai < bi {
			return -1
		} else if ai > bi {
			return 1
		}
		return 0
	}

	aCanonical := canonicalizeType(a.Type)
	bCanonical := canonicalizeType(b.Type)

	if aCanonical != bCanonical {
		if aCanonical < bCanonical {
			return -1, nil
		}
		return 1, nil
	}

	switch a.Type {
	case bsontype.Array:
		switch b.Type {
		case bsontype.Array:
			aElems, _ := a.Array().Elements()
			bElems, _ := b.Array().Elements()
			return compareDoc(aElems, bElems)
		}

	case bsontype.Binary:
		switch b.Type {
		case bsontype.Binary:
			aSubtype, aData := a.Binary()
			bSubtype, bData := b.Binary()
			return compareBinary(aSubtype, aData, bSubtype, bData), nil
		}

	case bsontype.Boolean:
		switch b.Type {
		case bsontype.Boolean:
			return compareBool(a.Boolean(), b.Boolean()), nil
		}

	case bsontype.CodeWithScope:
		switch b.Type {
		case bsontype.CodeWithScope:
			aCode, aScope := a.CodeWithScope()
			bCode, bScope := b.CodeWithScope()
			return compareCodeWithScope(aCode, aScope, bCode, bScope)
		}

	case bsontype.DBPointer:
		switch b.Type {
		case bsontype.DBPointer:
			ans, aid := a.DBPointer()
			bns, bid := b.DBPointer()
			return compareDBPointer(ans, aid, bns, bid), nil
		}

	case bsontype.DateTime:
		switch b.Type {
		case bsontype.DateTime:
			return compareInt64(a.DateTime(), b.DateTime()), nil
		}

	case bsontype.Double:
		switch b.Type {
		case bsontype.Double:
			return compareDouble(a.Double(), b.Double()), nil
		case bsontype.Int32:
			return compareDouble(a.Double(), float64(b.Int32())), nil
		case bsontype.Int64:
			return compareDouble(a.Double(), float64(b.Int64())), nil
		case bsontype.Decimal128:
			v, err := primitive.ParseDecimal128(fmt.Sprintf("%g", a.Double()))
			if err != nil {
				return 0, err
			}
			return decimalutil.Compare(
				decimalutil.FromPrimitive(v),
				decimalutil.FromPrimitive(b.Decimal128()),
			), nil
		}

	case bsontype.Decimal128:
		switch b.Type {
		case bsontype.Double:
			v, err := primitive.ParseDecimal128(fmt.Sprintf("%g", b.Double()))
			if err != nil {
				return 0, err
			}
			return decimalutil.Compare(
				decimalutil.FromPrimitive(a.Decimal128()),
				decimalutil.FromPrimitive(v),
			), nil
		case bsontype.Int32:
			return decimalutil.Compare(
				decimalutil.FromPrimitive(a.Decimal128()),
				decimalutil.FromInt32(b.Int32()),
			), nil
		case bsontype.Int64:
			return decimalutil.Compare(
				decimalutil.FromPrimitive(a.Decimal128()),
				decimalutil.FromInt64(b.Int64()),
			), nil
		case bsontype.Decimal128:
			return decimalutil.Compare(
				decimalutil.FromPrimitive(a.Decimal128()),
				decimalutil.FromPrimitive(b.Decimal128()),
			), nil
		}

	case bsontype.EmbeddedDocument:
		switch b.Type {
		case bsontype.EmbeddedDocument:
			aElems, _ := a.Document().Elements()
			bElems, _ := b.Document().Elements()
			return compareDoc(aElems, bElems)
		}

	case bsontype.Int32:
		switch b.Type {
		case bsontype.Int32:
			return compareInt64(int64(a.Int32()), int64(b.Int32())), nil
		case bsontype.Int64:
			return compareInt64(int64(a.Int32()), b.Int64()), nil
		case bsontype.Double:
			return compareDouble(float64(a.Int32()), b.Double()), nil
		case bsontype.Decimal128:
			return decimalutil.Compare(
				decimalutil.FromInt32(a.Int32()),
				decimalutil.FromPrimitive(b.Decimal128()),
			), nil
		}

	case bsontype.Int64:
		switch b.Type {
		case bsontype.Int32:
			return compareInt64(a.Int64(), int64(b.Int32())), nil
		case bsontype.Int64:
			return compareInt64(a.Int64(), b.Int64()), nil
		case bsontype.Double:
			return compareDouble(float64(a.Int64()), b.Double()), nil
		case bsontype.Decimal128:
			return decimalutil.Compare(
				decimalutil.FromInt64(a.Int64()),
				decimalutil.FromPrimitive(b.Decimal128()),
			), nil
		}

	case bsontype.JavaScript:
		switch b.Type {
		case bsontype.JavaScript:
			return strings.Compare(a.JavaScript(), b.JavaScript()), nil
		}

	case bsontype.MaxKey:
		switch b.Type {
		case bsontype.MaxKey:
			return 0, nil
		}

	case bsontype.MinKey:
		switch b.Type {
		case bsontype.MinKey:
			return 0, nil
		}

	case bsontype.ObjectID:
		switch b.Type {
		case bsontype.ObjectID:
			return compareObjectID(a.ObjectID(), b.ObjectID()), nil
		}

	case bsontype.Null:
		switch b.Type {
		case bsontype.Null:
			return 0, nil
		}

	case bsontype.Regex:
		switch b.Type {
		case bsontype.Regex:
			aPattern, aOptions := a.Regex()
			bPattern, bOptions := b.Regex()
			return compareRegex(aPattern, aOptions, bPattern, bOptions), nil
		}

	case bsontype.String:
		switch b.Type {
		case bsontype.String:
			return strings.Compare(a.StringValue(), b.StringValue()), nil
		}

	case bsontype.Symbol:
		switch b.Type {
		case bsontype.Symbol:
			return strings.Compare(a.Symbol(), b.Symbol()), nil
		}

	case bsontype.Timestamp:
		switch b.Type {
		case bsontype.Timestamp:
			at, ai := a.Timestamp()
			bt, bi := b.Timestamp()
			return compareTimestamp(at, ai, bt, bi), nil
		}

	case bsontype.Undefined:
		switch b.Type {
		case bsontype.Undefined:
			return 0, nil
		}
	}

	return 0, errors.Errorf("cannot compare %s to %s", a.Type, b.Type)
}

func canonicalizeType(t bsontype.Type) int {
	switch t {
	case bsontype.MinKey:
		return 0
	case bsontype.MaxKey:
		return 255
	case bsontype.Undefined:
		return 0
	case bsontype.Null:
		return 5
	case bsontype.Decimal128:
		return 10
	case bsontype.Double:
		return 10
	case bsontype.Int32:
		return 10
	case bsontype.Int64:
		return 10
	case bsontype.String:
		return 15
	case bsontype.Symbol:
		return 15
	case bsontype.EmbeddedDocument:
		return 20
	case bsontype.Array:
		return 25
	case bsontype.Binary:
		return 30
	case bsontype.ObjectID:
		return 35
	case bsontype.Boolean:
		return 40
	case bsontype.DateTime:
		return 45
	case bsontype.Timestamp:
		return 47
	case bsontype.Regex:
		return 50
	case bsontype.DBPointer:
		return 55
	case bsontype.JavaScript:
		return 60
	case bsontype.CodeWithScope:
		return 65
	}

	panic(fmt.Errorf("unknown type %s %d", t, t))
}
