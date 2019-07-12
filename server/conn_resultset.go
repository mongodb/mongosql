package server

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"encoding/hex"
	"unicode/utf8"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/evaluator/results"
	"github.com/10gen/sqlproxy/evaluator/types"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/evaluator/variable"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/internal/strutil"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/schema"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// dataFormatter holds information necessary for formatting
// the row data, for a particular column type, based on its value.
type dataFormatter struct {
	fieldName            string
	columnType           types.EvalType
	evalType             types.EvalType
	uuidSubtype          types.EvalType
	charSet              *collation.Charset
	mongoDBVarcharLength int
	data                 []byte
}

func newDataFormatter(fieldName string, columnType, evalType, uuidSubtype types.EvalType,
	charSet *collation.Charset, mongoDBVarcharLength int, data []byte) dataFormatter {
	return dataFormatter{fieldName, columnType, evalType, uuidSubtype, charSet,
		mongoDBVarcharLength, data}
}

// fastFormat formats values quickly. It returns data encoded in MySQL's wire
// protocol as a []byte, or an error if formatting failed.
func fastFormat(f dataFormatter, valueKind values.SQLValueKind) ([]byte, error) {
	columnType, evalType, uuidSubtype := f.columnType, f.evalType, f.uuidSubtype

	// Null do not need special treatment - regardless of the SQL column type.
	if evalType == types.EvalNull {
		return []byte{0xfb}, nil
	}

	// If we can serialize directly from MongoDB to MySQL's wire protocol, do that. There are a few
	// conditions under which we can do that. For all conditions, the UUID subtype must be
	// EvalBinary.
	//
	// 1. If the columnType is equal to the evalType, we do not need to go through any type
	//	   conversions, so we can send data straight from MongoDB to the MySQL wire protocol.
	// 2. If the columnType is types.EvalPolymorphic because that is used when the BIC does not care
	// 	  about the output type.
	// 3. If the MongoDB type is ObjectID and the column type is SQLVarchar (types.EvalString),
	//	  since ObjectID's are serialized as strings by default.

	isSameEvalType := columnType == evalType
	isEvalPolymorphicType := columnType == types.EvalPolymorphic
	isObjectIDType := columnType == types.EvalString && evalType == types.EvalObjectID
	hasStandardUUIDSubtype := uuidSubtype == types.EvalBinary

	if hasStandardUUIDSubtype && (isSameEvalType || isEvalPolymorphicType || isObjectIDType) {
		return fastCleanFormat(columnType, evalType, &f)
	}

	sqlVal, err := values.BSONValueToSQLValue(valueKind, evalType, uuidSubtype, f.data)
	if err != nil {
		readableColumnType := types.EvalTypeToMongoType(columnType)
		return nil, fmt.Errorf("%s, expected type '%s' for field '%s'", err,
			readableColumnType, f.fieldName)
	}

	converted := values.ConvertTo(sqlVal, columnType)
	bytes, err := converted.WireProtocolEncode(f.charSet, f.mongoDBVarcharLength)
	if err != nil {
		return nil, err
	}

	if bytes == nil {
		return []byte{0xfb}, nil
	}
	return putLengthEncodedString(bytes), nil
}

// fastCleanFormat produces byte arrays encoded in MySQL's wire protocol based
// on the passed EvalType and the data. The dataFormatter's charSet and
// mongoDBVarcharLength fields are necessary for handling string encoding
// and string max allowed length, respectively.
func fastCleanFormat(columnType, evalType types.EvalType, f *dataFormatter) ([]byte, error) {
	switch evalType {
	case types.EvalDouble:
		ret := strconv.AppendFloat(nil,
			math.Float64frombits((uint64(f.data[0])<<0)|
				(uint64(f.data[1])<<8)|
				(uint64(f.data[2])<<16)|
				(uint64(f.data[3])<<24)|
				(uint64(f.data[4])<<32)|
				(uint64(f.data[5])<<40)|
				(uint64(f.data[6])<<48)|
				(uint64(f.data[7])<<56),
			),
			'f', -1, 64,
		)
		return putLengthEncodedString(ret), nil
	case types.EvalString:
		// Subtract 1 from the length because we will
		// not send the trailing null byte, and MySQL
		// expects the length to be exact.
		l := ((uint32(f.data[0]) << 0) |
			(uint32(f.data[1]) << 8) |
			(uint32(f.data[2]) << 16) |
			(uint32(f.data[3]) << 24)) - 1
		if f.data[len(f.data)-1] != '\x00' {
			return nil, fmt.Errorf("corrupted string field: not 0x00 terminated")
		}
		if len(f.data) != int(l)+5 {
			return nil, fmt.Errorf("corrupted string field: length mismatch")
		}
		if !utf8.Valid(f.data[4 : len(f.data)-1]) {
			return nil, fmt.Errorf("corrupted string field: not valid unicode")
		}
		if string(f.charSet.Name) == "utf8" {
			// Rather than using putLengthEncodedString, which results in a allocation
			// and copy of the entire string, we will use the already allocated
			// f.data buffer. MySQL uses variable width encoding for its lengths,
			// while MongoDB uses a fixed 32 bits.
			switch {
			case l <= 250:
				f.data = f.data[3 : len(f.data)-1]
				f.data[0] = byte(l)
				return f.data, nil
			case l <= 0xffff:
				f.data = f.data[1 : len(f.data)-1]
				f.data[0], f.data[1], f.data[2] = 0xfc, byte(l), byte(l>>8)
				return f.data, nil
			case l <= 0xffffff:
				f.data = f.data[0 : len(f.data)-1]
				f.data[0], f.data[1], f.data[2], f.data[3] = 0xfd, byte(l), byte(l>>8), byte(l>>16)
				return f.data, nil
			}
			// Unfortunately, if we get to this point, MySQL encodes the length using 9 bytes
			// which overflows our f.data slice. We have no choice but to copy. String
			// this long should be pretty rare (in fact, the are right at the Document size limit).
			f.data = f.data[4 : len(f.data)-1]
			return putLengthEncodedString(f.data), nil
		}
		f.data = f.data[4 : len(f.data)-1]
		ret := f.charSet.Encode(f.data)
		// Varchars are counted by characters, not bytes. Use runes to
		// account for multi-byte characters. Since we know the number
		// of characters can't be more than the number of bytes, we can
		// skip the character length check, if the byte length is satisfactory.
		if f.mongoDBVarcharLength != 0 && len(ret) > f.mongoDBVarcharLength {
			runes := []rune(string(ret))
			if len(runes) > f.mongoDBVarcharLength {
				runes = runes[:f.mongoDBVarcharLength]
				ret = []byte(string(runes))
			}
		}
		return putLengthEncodedString(ret), nil
	case types.EvalBinary:
		l := ((uint32(f.data[0]) << 0) |
			(uint32(f.data[1]) << 8) |
			(uint32(f.data[2]) << 16) |
			(uint32(f.data[3]) << 24))
		subType := f.data[4]
		f.data = f.data[5:]
		if len(f.data) != int(l) {
			return nil, fmt.Errorf("corrupted binary field")
		}
		if !(subType == 0x03 || subType == 0x04) {
			return nil, fmt.Errorf("UUID types 0x3 and 0x4 are the only supported binary "+
				"subtybes, not %#02x", subType)
		}
		str := hex.EncodeToString(f.data)
		ret := str[0:8] + "-" + str[8:12] + "-" + str[12:16] + "-" + str[16:20] + "-" + str[20:]
		if string(f.charSet.Name) == "utf8" {
			return putLengthEncodedString([]byte(ret)), nil
		}
		return putLengthEncodedString(f.charSet.Encode([]byte(ret))), nil
	case types.EvalObjectID:
		str := []byte(hex.EncodeToString(f.data))
		ret := f.charSet.Encode(str)
		if string(f.charSet.Name) == "utf8" {
			return putLengthEncodedString(ret), nil
		}
		return putLengthEncodedString(f.charSet.Encode(ret)), nil
	case types.EvalBoolean:
		f.data[0] += 48
		return putLengthEncodedString(f.data), nil
	case types.EvalDatetime:
		i := int64((uint64(f.data[0]) << 0) |
			(uint64(f.data[1]) << 8) |
			(uint64(f.data[2]) << 16) |
			(uint64(f.data[3]) << 24) |
			(uint64(f.data[4]) << 32) |
			(uint64(f.data[5]) << 40) |
			(uint64(f.data[6]) << 48) |
			(uint64(f.data[7]) << 56))
		var t time.Time
		if i == -62135596800000 {
			t = time.Time{}.In(schema.DefaultLocale)
		} else {
			t = time.Unix(i/1e3, i%1e3*1e6).In(schema.DefaultLocale)
		}
		if t.Nanosecond() != 0 {
			return putLengthEncodedString(strutil.Slice(t.Format(schema.TimestampFormatMicros))), nil
		}
		return putLengthEncodedString(strutil.Slice(t.Format(schema.TimestampFormat))), nil
	case types.EvalNull:
		return []byte{0xfb}, nil
	case types.EvalInt32:
		ret := strconv.AppendInt(nil,
			int64(int32((uint32(f.data[0])<<0)|
				(uint32(f.data[1])<<8)|
				(uint32(f.data[2])<<16)|
				(uint32(f.data[3])<<24),
			)),
			10,
		)
		return putLengthEncodedString(ret), nil
	case types.EvalInt64:
		ret := strconv.AppendInt(nil,
			int64((uint64(f.data[0])<<0)|
				(uint64(f.data[1])<<8)|
				(uint64(f.data[2])<<16)|
				(uint64(f.data[3])<<24)|
				(uint64(f.data[4])<<32)|
				(uint64(f.data[5])<<40)|
				(uint64(f.data[6])<<48)|
				(uint64(f.data[7])<<56)),
			10,
		)
		return putLengthEncodedString(ret), nil
	case types.EvalDecimal128:
		l := (uint64(f.data[0]) << 0) |
			(uint64(f.data[1]) << 8) |
			(uint64(f.data[2]) << 16) |
			(uint64(f.data[3]) << 24) |
			(uint64(f.data[4]) << 32) |
			(uint64(f.data[5]) << 40) |
			(uint64(f.data[6]) << 48) |
			(uint64(f.data[7]) << 56)
		h := (uint64(f.data[8]) << 0) |
			(uint64(f.data[9]) << 8) |
			(uint64(f.data[10]) << 16) |
			(uint64(f.data[11]) << 24) |
			(uint64(f.data[12]) << 32) |
			(uint64(f.data[13]) << 40) |
			(uint64(f.data[14]) << 48) |
			(uint64(f.data[15]) << 56)
		d := primitive.NewDecimal128(h, l)
		return putLengthEncodedString([]byte(d.String())), nil
	default:
		readableBSONType := string(types.EvalTypeToMongoType(evalType))
		readableColumnType := string(types.EvalTypeToMongoType(columnType))

		return nil, fmt.Errorf("unexpected bson type: found %s but expected %s for field %s",
			readableBSONType, readableColumnType, f.fieldName)
	}
}

func formatHeaderField(variables *variable.Container, field *Field,
	value values.SQLValue) error {
	switch value.(type) {
	case values.SQLFloat:
		field.Type = MySQLTypeDouble
		field.Decimal = 0x1f
		field.Flag = BinaryFlag
	case values.SQLDecimal128:
		field.Type = MySQLTypeNewDecimal
		field.Decimal = 20      // scale
		field.ColumnLength = 67 // precision plus 2 (decimal point and length)
	case values.SQLBool:
		field.Type = MySQLTypeTiny
	case values.SQLUint64:
		field.Type = MySQLTypeLongLong
		field.Flag = BinaryFlag | UnsignedFlag
	case values.SQLInt64:
		field.Type = MySQLTypeLongLong
		field.Flag = BinaryFlag
	case values.SQLObjectID, values.SQLVarchar:
		field.Type = MySQLTypeVarString
		length := uint32(variables.GetUint64(variable.MongoDBMaxVarcharLength))
		if length == 0 {
			length = math.MaxUint16
		}
		field.ColumnLength = length
	case values.SQLDate:
		field.Type = MySQLTypeDate
	case values.SQLTimestamp:
		field.Type = MySQLTypeDatetime
	case nil:
		field.Type = MySQLTypeNull
	default:
		return mysqlerrors.Unknownf("unsupported type %T for result set", value)
	}

	return nil
}

// writeHeaders writes the column headers for a resultset.
func (c *conn) writeHeaders(columns []*results.Column, colID collation.ID) error {
	columnLen := putLengthEncodedInt(uint64(len(columns)))

	data := make([]byte, 4, 1024)

	data = append(data, columnLen...)

	// Write the column count.
	if err := c.writePacket(data); err != nil {
		return err
	}
	status := c.status()

	numFields := len(columns)

	if numFields == 0 {
		return mysqlerrors.Unknownf("No columns found in result set")
	}

	valueKind := evaluator.GetSQLValueKind(c.variables)
	for j := 0; j < numFields; j++ {
		field := &Field{
			Name:          []byte(columns[j].Name),
			OriginalName:  []byte(columns[j].OriginalName),
			Schema:        []byte(columns[j].Database),
			Table:         []byte(columns[j].Table),
			OriginalTable: []byte(columns[j].OriginalTable),
			Charset:       uint16(colID),
		}

		err := formatHeaderField(c.variables, field, evaluator.ZeroValue(columns[j].EvalType, valueKind))
		if err != nil {
			return err
		}

		data = data[0:4]
		data = append(data, field.Dump(c.variables.GetCharset(variable.CharacterSetResults))...)

		// Write a column definition packet for each column in the resultset.
		if err = c.writePacket(data); err != nil {
			return err
		}
	}
	// End the column definitions with an EOF packet.
	return c.writeEOF(status)
}

// streamRows receives packets from a packet producer across the packetChan and writes them to the
// conn. It is also responsible for closing the Iter once the packetChan is closed.
func (c *conn) streamRows(ctx context.Context, packetChan chan []byte, errChan chan error, columns []*results.Column,
	iter evaluator.Iter) (err error) {
	status := c.status()

	var col *collation.Collation
	col, err = collation.Get(
		c.variables.GetCharset(variable.CharacterSetResults).DefaultCollationName)
	if err != nil {
		return err
	}
	var wroteHeaders bool
	count := 0
	totalBytes := uint64(0)
streamer:
	for {
		select {
		case packet, ok := <-packetChan:
			if !ok {
				break streamer
			}

			// Write the headers once.
			if !wroteHeaders {
				if err = c.writeHeaders(columns, col.ID); err != nil {
					return err
				}
				wroteHeaders = true
			}

			// Write each packet from the producer.
			if err = c.writePacket(packet); err != nil {
				return err
			}
			count++
			totalBytes += uint64(len(packet))

		case <-ctx.Done():
			return ctx.Err()

		case err = <-errChan:
			_ = iter.Close()
			return err
		}
	}

	if err = iter.Close(); err != nil {
		c.logger.Errf(log.Dev, "iterator close err: %v", err)
		return err
	}

	if err = iter.Err(); err != nil {
		c.logger.Errf(log.Dev, "iterator err: %v", err)
		return err
	}

	if !wroteHeaders {
		if err = c.writeHeaders(columns, col.ID); err != nil {
			return err
		}
	}

	c.logger.Infof(log.Admin, "returned %d %s (%s)", count, strutil.Pluralize(count, "row", "rows"),
		strutil.ByteString(totalBytes))

	if err = ctx.Err(); err != nil {
		return err
	}
	return c.writeEOF(status)
}

// sendPackets is used to produce packets from the Rows returned by
// the passed Iter. The results are passed as a []byte to the
// packetChan channel.
func (c *conn) sendPackets(ctx context.Context, packetChan chan []byte, iter evaluator.RowIter) {
	r := &results.Row{}
	charSet := c.variables.GetCharset(variable.CharacterSetResults)
	mongoDBVarcharLength := int(c.variables.GetUint64(variable.MongoDBMaxVarcharLength))
	for ctx.Err() == nil && iter.Next(ctx, r) {
		packet := []byte{0, 0, 0, 0}
		for _, value := range r.Data {
			b, err := value.Data.WireProtocolEncode(charSet, mongoDBVarcharLength)
			if err != nil {
				close(packetChan)
				panic(err)
			}
			if b == nil {
				packet = append(packet, 0xfb)
			} else {
				packet = append(packet, putLengthEncodedString(b)...)
			}
		}
		packetChan <- packet
	}

	if ctx.Err() != nil {
		_ = iter.Close()
	}
	close(packetChan)
}

// fastSendPackets is used to produce packets from Documents returned from the
// passed fastIter that have a guaranteed field order, which allows for more
// optimally finding fields in the Document. The results are returned as a
// []byte across the packetChan channel.
func (c *conn) fastSendPackets(ctx context.Context, packetChan chan []byte, fastIter evaluator.DocIter) {
	valueKind := evaluator.GetSQLValueKind(c.variables)
	charSet := c.variables.GetCharset(variable.CharacterSetResults)
	mongoDBVarcharLength := int(c.variables.GetUint64(variable.MongoDBMaxVarcharLength))

	doc := &bson.Raw{}

	maybePanic := func(err error) {
		if err != nil {
			close(packetChan)
			panic(err)
		}
	}

	columnInfo := fastIter.GetColumnInfo()
	lenColumnFields := len(columnInfo)
	for ctx.Err() == nil && fastIter.Next(ctx, doc) {
		columnInfo = fastIter.GetColumnInfo()
		packet := []byte{0, 0, 0, 0}
		vs, err := doc.Elements()
		maybePanic(err)
		lenValues := len(vs)
		if lenValues == lenColumnFields {
			// No missing values, so we can iterate fast without checking key names.
			for i, v := range vs {
				fieldName := columnInfo[i].Field
				columnType := columnInfo[i].Type
				uuidSubtype := columnInfo[i].UUIDSubtype
				value := v.Value()
				df := newDataFormatter(fieldName, columnType,
					types.EvalType(value.Type), uuidSubtype,
					charSet, mongoDBVarcharLength, value.Value)
				b, err := fastFormat(df, valueKind)
				maybePanic(err)
				packet = append(packet, b...)
			}
		} else {
			// If we have missing fields, we need to check key names. Until
			// we have determined all the missing fields.
			numMissingValues := lenColumnFields - lenValues
			i := 0
			for _, info := range columnInfo {
				fieldName := info.Field
				columnType := info.Type
				uuidSubtype := info.UUIDSubtype
				if numMissingValues > 0 && i < len(vs) {
					if fieldName == vs[i].Key() {
						value := vs[i].Value()
						// If this is the correct fieldName, output the value.
						df := newDataFormatter(fieldName, columnType,
							types.EvalType(value.Type), uuidSubtype,
							charSet, mongoDBVarcharLength, value.Value)
						b, err := fastFormat(df, valueKind)
						maybePanic(err)
						packet = append(packet, b...)
						// increment i so that we consider the next value.
						i++
					} else {
						// If the fieldName is wrong, this field must be missing, output
						// a NULL, decrement numMissingValues (because we found one), but do NOT
						// touch i because we want the same position in the values next
						// iteration.
						packet = append(packet, 0xfb)
						numMissingValues--
					}
				} else if i < len(vs) {
					// We have found all the missing values, default to the faster mode.
					value := vs[i].Value()
					i++
					df := newDataFormatter(fieldName, columnType,
						types.EvalType(value.Type), uuidSubtype,
						charSet, mongoDBVarcharLength, value.Value)
					b, err := fastFormat(df, valueKind)
					maybePanic(err)
					packet = append(packet, b...)
				} else {
					// i >= len(values), break to where we add NULLS, if necessary.
					break
				}
			}
			// We ran out of values, all values after this point must be missing.
			for ; numMissingValues != 0; numMissingValues-- {
				packet = append(packet, 0xfb)
			}
		}
		packetChan <- packet
	}
	if ctx.Err() != nil {
		_ = fastIter.Close()
	}
	close(packetChan)
}

// fastSendPackets32 is used to produce packets from Documents returned from the
// passed fastIter. The results are returned as a []byte across the packetChan
// channel. The Documents returned by the fastIter do not have a guaranteed
// field order.
func (c *conn) fastSendPackets32(ctx context.Context, packetChan chan []byte, fastIter evaluator.DocIter) {
	valueKind := evaluator.GetSQLValueKind(c.variables)
	charSet := c.variables.GetCharset(variable.CharacterSetResults)
	mongoDBVarcharLength := int(c.variables.GetUint64(variable.MongoDBMaxVarcharLength))

	doc := &bson.Raw{}
	columnInfo := fastIter.GetColumnInfo()
	lenColumnInfo := len(columnInfo)
	// We will use one nullField value to represent all NULLs that will result
	// from missing fields.
	nullField := &bson.RawValue{Type: bson.TypeNull, Value: []byte{}}
	fieldMap := make(map[string]*bson.RawValue, lenColumnInfo)
	// Set the value for all columns to null so we can avoid
	// a branch in the forloop below.
	for _, info := range columnInfo {
		fieldMap[info.Field] = nullField
	}
	for ctx.Err() == nil && fastIter.Next(ctx, doc) {
		columnInfo = fastIter.GetColumnInfo()
		packet := []byte{0, 0, 0, 0}
		vs, err := doc.Elements()
		if err != nil {
			close(packetChan)
			panic(err)
		}
		// We can't rely on field ordering in 3.2.
		for _, v := range vs {
			value := v.Value()
			fieldMap[v.Key()] = &value
		}
		for _, info := range columnInfo {
			value := fieldMap[info.Field]
			df := newDataFormatter(info.Field, info.Type,
				types.EvalType(value.Type), info.UUIDSubtype,
				charSet, mongoDBVarcharLength, value.Value)
			b, err := fastFormat(df, valueKind)
			if err != nil {
				close(packetChan)
				panic(err)
			}
			packet = append(packet, b...)
			// reset the fields to the nullField for the next iteration.
			fieldMap[info.Field] = nullField
		}
		packetChan <- packet
	}
	if ctx.Err() != nil {
		_ = fastIter.Close()
	}
	close(packetChan)

}

// streamRowResultset implements COM_QUERY response.
// More at https://dev.mysql.com/doc/internals/en/com-query-response.html
//
// It uses a producer function defined by the type of iterator and possibly server
// version, which formats data to byte packets, and a consumer that actually
// writes those byte packets to the client. This producer consumer relation
// allows for query cancellation. It operates on a RowIter, which means the results
// are coming from in memory evaluation.
func (c *conn) streamRowResultset(ctx context.Context, columns []*results.Column, rowIter evaluator.RowIter) (err error) {
	packetChan := make(chan []byte, 1)
	errChan := make(chan error, 1)

	errorHandler := func(err interface{}) {
		errChan <- fmt.Errorf("iterating error: %v", err)
	}

	if len(columns) == 0 {
		err := c.writeOK(nil)
		if err != nil {
			return err
		}
	}

	c.affectedRows = int64(-1)
	asyncPacketSender := func() { c.sendPackets(ctx, packetChan, rowIter) }

	procutil.PanicSafeGo(asyncPacketSender, errorHandler)
	return c.streamRows(ctx, packetChan, errChan, columns, rowIter)
}

// streamDocRowResultset implements COM_QUERY response.
// More at https://dev.mysql.com/doc/internals/en/com-query-response.html
//
// It uses a producer function defined by the type of iterator and possibly server
// version, which formats data to byte packets, and a consumer that actually
// writes those byte packets to the client. This producer consumer relation
// allows for query cancellation. It operators on a DocIter, which means the results
// are coming straight from MongoDB.
func (c *conn) streamDocResultset(ctx context.Context, columns []*results.Column, docIter evaluator.DocIter) (err error) {
	packetChan := make(chan []byte, 1)
	errChan := make(chan error, 1)

	errorHandler := func(err interface{}) {
		errChan <- fmt.Errorf("iterating error: %v", err)
	}

	if len(columns) == 0 {
		err := c.writeOK(nil)
		if err != nil {
			return err
		}
	}

	var asyncPacketSender func()
	if c.mongoDBInfo.VersionAtLeast(3, 4, 0) {
		asyncPacketSender = func() { c.fastSendPackets(ctx, packetChan, docIter) }
	} else {
		// For server < 3.4, we cannot rely on document field ordering.
		asyncPacketSender = func() { c.fastSendPackets32(ctx, packetChan, docIter) }
	}

	procutil.PanicSafeGo(asyncPacketSender, errorHandler)
	return c.streamRows(ctx, packetChan, errChan, columns, docIter)
}
