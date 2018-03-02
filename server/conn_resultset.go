package server

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"encoding/hex"
	"unicode/utf8"

	"github.com/10gen/mongo-go-driver/bson"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
	"github.com/shopspring/decimal"
)

// dataFormatter holds information necessary for formatting
// the row data, for a particular column type, based on its value.
type dataFormatter struct {
	columnType           schema.BSONSpecType
	valueType            schema.BSONSpecType
	uuidSubtype          schema.BSONSpecType
	charSet              *collation.Charset
	mongoDBVarcharLength int
	data                 []byte
}

func newDataFormatter(columnType, valueType, uuidSubtype schema.BSONSpecType,
	charSet *collation.Charset, mongoDBVarcharLength int, data []byte) dataFormatter {
	return dataFormatter{columnType, valueType, uuidSubtype, charSet,
		mongoDBVarcharLength, data}
}

// fastFormat formats values quickly. It returns data encoded in MySQL's wire
// protocol as a []byte, or an error if formatting failed.
func fastFormat(f dataFormatter) ([]byte, error) {
	columnType, valueType, uuidSubtype := f.columnType, f.valueType, f.uuidSubtype
	// Fast track NULLs, as NULLs are fairly common and easy to serialize:
	if valueType == schema.BSONNull {
		return []byte{0xfb}, nil
	}
	// If we can serialize directly from MongoDB to MySQL's wire protocol, do
	// that. There are a few conditions under which we can do that: first,
	// the uuidSubtype must be 0, which is the case for anything that is
	// not a UUID, as well as most UUIDs. If the columnType is equal to
	// the valueType, we do not need to go through any type conversions, so
	// we can forward straight from MongoDB to the MySQL wire protocol.
	// Another case is if the MongoDB type is ObjectID and the column type
	// is SQLVarchar (schema.BSONString), because ObjectID's are serialized
	// as strings by default. The last case is if the columnType is
	// schema.BSONNone because that is used when the BIC does not care
	// about the output type.
	if uuidSubtype == 0 && (columnType == valueType ||
		(valueType == schema.BSONObjectID && columnType == schema.BSONString) ||
		columnType == schema.BSONNone) {
		return fastCleanFormat(valueType, f.charSet, f.mongoDBVarcharLength, f.data)
	}

	sqlVal, err := evaluator.BSONValueToSQLValue(valueType, uuidSubtype, f.data)
	if err != nil {
		return nil, err
	}
	converted := sqlVal.ConvertTo(columnType)
	bytes, err := converted.MySQLEncode(f.charSet, f.mongoDBVarcharLength)
	if err != nil {
		return nil, err
	}
	if bytes == nil {
		return []byte{0xfb}, nil
	}
	return putLengthEncodedString(bytes), nil
}

// fastCleanFormat produces byte arrays encoded in MySQL's wire protocol based
// on the passed BSONSpecType and the data. The charSet and
// mongoDBVarcharLength arguments are necessary for handling string encoding
// and string max allowed length, respectively.
func fastCleanFormat(bsonType schema.BSONSpecType, charSet *collation.Charset,
	mongoDBVarcharLength int, data []byte) ([]byte, error) {
	switch bsonType {
	case schema.BSONDouble:
		ret := strconv.AppendFloat(nil,
			math.Float64frombits((uint64(data[0])<<0)|
				(uint64(data[1])<<8)|
				(uint64(data[2])<<16)|
				(uint64(data[3])<<24)|
				(uint64(data[4])<<32)|
				(uint64(data[5])<<40)|
				(uint64(data[6])<<48)|
				(uint64(data[7])<<56),
			),
			'f', -1, 64,
		)
		return putLengthEncodedString(ret), nil
	case schema.BSONString:
		// Subtract 1 from the length because we will
		// not send the trailing null byte, and MySQL
		// expects the length to be exact.
		l := ((uint32(data[0]) << 0) |
			(uint32(data[1]) << 8) |
			(uint32(data[2]) << 16) |
			(uint32(data[3]) << 24)) - 1
		if data[len(data)-1] != '\x00' {
			return nil, fmt.Errorf("corrupted string field: not 0x00 terminated")
		}
		if len(data) != int(l)+5 {
			return nil, fmt.Errorf("corrupted string field: length mismatch")
		}
		if !utf8.Valid(data[4 : len(data)-1]) {
			return nil, fmt.Errorf("corrupted string field: not valid unicode")
		}
		if string(charSet.Name) == "utf8" {
			// Rather than using putLengthEcondedString, which results in a allocation
			// and copy of the entire string, we will use the already allocated
			// data buffer. MySQL uses variable width encoding for its lengths,
			// while MongoDB uses a fixed 32 bits.
			switch {
			case l <= 250:
				data = data[3 : len(data)-1]
				data[0] = byte(l)
				return data, nil
			case l <= 0xffff:
				data = data[1 : len(data)-1]
				data[0], data[1], data[2] = 0xfc, byte(l), byte(l>>8)
				return data, nil
			case l <= 0xffffff:
				data = data[0 : len(data)-1]
				data[0], data[1], data[2], data[3] = 0xfd, byte(l), byte(l>>8), byte(l>>16)
				return data, nil
			}
			// Unfortunately, if we get to this point, MySQL encodes the length using 9 bytes
			// which overflows our data slice. We have no choice but to copy. String
			// this long should be pretty rare (in fact, the are right at the Document size limit).
			data = data[4 : len(data)-1]
			return putLengthEncodedString(data), nil
		}
		data = data[4 : len(data)-1]
		ret := charSet.Encode(data)
		// Varchars are counted by characters, not bytes. Use runes to
		// account for multi-byte characters. Since we know the number
		// of characters can't be more than the number of bytes, we can
		// skip the character length check, if the byte length is satisfactory.
		if mongoDBVarcharLength != 0 && len(ret) > mongoDBVarcharLength {
			runes := []rune(string(ret))
			if len(runes) > mongoDBVarcharLength {
				runes = runes[:mongoDBVarcharLength]
				ret = []byte(string(runes))
			}
		}
		return putLengthEncodedString(ret), nil
	case schema.BSONUUID:
		l := ((uint32(data[0]) << 0) |
			(uint32(data[1]) << 8) |
			(uint32(data[2]) << 16) |
			(uint32(data[3]) << 24))
		subType := data[4]
		data = data[5:]
		if len(data) != int(l) {
			return nil, fmt.Errorf("corrupted binary field")
		}
		if !(subType == 0x03 || subType == 0x04) {
			return nil, fmt.Errorf("UUID types 0x3 and 0x4 are the only supported binary "+
				"subtybes, not %#02x", subType)
		}
		str := hex.EncodeToString(data)
		return putLengthEncodedString([]byte(str[0:8] +
			"-" + str[8:12] +
			"-" + str[12:16] +
			"-" + str[16:20] +
			"-" + str[20:])), nil
	case schema.BSONObjectID:
		return putLengthEncodedString([]byte(hex.EncodeToString(data))), nil
	case schema.BSONBoolean:
		data[0] += 48
		return putLengthEncodedString(data), nil
	case schema.BSONTimestamp:
		i := int64((uint64(data[0]) << 0) |
			(uint64(data[1]) << 8) |
			(uint64(data[2]) << 16) |
			(uint64(data[3]) << 24) |
			(uint64(data[4]) << 32) |
			(uint64(data[5]) << 40) |
			(uint64(data[6]) << 48) |
			(uint64(data[7]) << 56))
		var t time.Time
		if i == -62135596800000 {
			t = time.Time{}.In(schema.DefaultLocale)
		} else {
			t = time.Unix(i/1e3, i%1e3*1e6).In(schema.DefaultLocale)
		}
		if t.Nanosecond() != 0 {
			return putLengthEncodedString(util.Slice(t.Format(schema.TimestampFormatMicros))), nil
		}
		return putLengthEncodedString(util.Slice(t.Format(schema.TimestampFormat))), nil
	case schema.BSONNull:
		return []byte{0xfb}, nil
	case schema.BSONInt:
		ret := strconv.AppendInt(nil,
			int64(int32((uint32(data[0])<<0)|
				(uint32(data[1])<<8)|
				(uint32(data[2])<<16)|
				(uint32(data[3])<<24),
			)),
			10,
		)
		return putLengthEncodedString(ret), nil
	case schema.BSONInt64:
		ret := strconv.AppendInt(nil,
			int64((uint64(data[0])<<0)|
				(uint64(data[1])<<8)|
				(uint64(data[2])<<16)|
				(uint64(data[3])<<24)|
				(uint64(data[4])<<32)|
				(uint64(data[5])<<40)|
				(uint64(data[6])<<48)|
				(uint64(data[7])<<56)),
			10,
		)
		return putLengthEncodedString(ret), nil
	case schema.BSONDecimal128:
		h := (uint64(data[0]) << 0) |
			(uint64(data[1]) << 8) |
			(uint64(data[2]) << 16) |
			(uint64(data[3]) << 24) |
			(uint64(data[4]) << 32) |
			(uint64(data[5]) << 40) |
			(uint64(data[6]) << 48) |
			(uint64(data[7]) << 56)
		l := (uint64(data[8]) << 0) |
			(uint64(data[9]) << 8) |
			(uint64(data[10]) << 16) |
			(uint64(data[11]) << 24) |
			(uint64(data[12]) << 32) |
			(uint64(data[13]) << 40) |
			(uint64(data[14]) << 48) |
			(uint64(data[15]) << 56)
		d := evaluator.NewBSONDecimal128(l, h)
		return putLengthEncodedString([]byte(d.String())), nil
	default:
		return nil, fmt.Errorf("unimplemented bson type: %#x", bsonType)
	}
}

// formatValue formats a SQLValue into MySQL's wire protocol as a []byte,
// or returns an error if formatting failed.
func (c *conn) formatValue(value evaluator.SQLValue) ([]byte, error) {
	switch v := value.(type) {
	case evaluator.SQLVarchar:
		bytes := c.variables.GetCharset(variable.CharacterSetResults).Encode(util.Slice(string(v)))
		// Varchars are counted by characters, not bytes. Use runes to
		// account for multi-byte characters. Since we know the number
		// of characters can't be more than the number of bytes, we can
		// skip the character length check if the byte length is satisfactory.
		mongoDBVarcharLength := c.variables.GetUInt16(variable.MongoDBMaxVarcharLength)
		if mongoDBVarcharLength != 0 && len(bytes) > int(mongoDBVarcharLength) {
			runes := []rune(string(bytes))
			if len(runes) > int(mongoDBVarcharLength) {
				runes = runes[:mongoDBVarcharLength]
				bytes = []byte(string(runes))
			}
		}

		return bytes, nil
	case evaluator.SQLObjectID:
		return []byte(string(v)), nil
	case evaluator.SQLUUID:
		return []byte(v.String()), nil
	case evaluator.SQLInt:
		return strconv.AppendInt(nil, int64(v), 10), nil
	case evaluator.SQLDecimal128:
		return []byte(util.FormatDecimal(decimal.Decimal(v))), nil
	case evaluator.SQLUint32:
		return strconv.AppendUint(nil, uint64(v), 10), nil
	case evaluator.SQLUint64:
		return strconv.AppendUint(nil, uint64(v), 10), nil
	case evaluator.SQLFloat:
		return strconv.AppendFloat(nil, float64(v), 'f', -1, 64), nil
	case evaluator.SQLNullValue, *evaluator.SQLNullValue, evaluator.SQLNoValue:
		return nil, nil
	case evaluator.SQLBool:
		if v.Bool() {
			return []byte{'1'}, nil
		}
		return []byte{'0'}, nil
	case *evaluator.SQLValues:
		return c.formatValue(v.Values[0])
	// SQL time related values
	case evaluator.SQLDate:
		return util.Slice(v.Time.Format(schema.DateFormat)), nil
	case evaluator.SQLTimestamp:
		if strings.Contains(v.Time.String(), ".") {
			return util.Slice(v.Time.Format(schema.TimestampFormatMicros)), nil
		}
		return util.Slice(v.Time.Format(schema.TimestampFormat)), nil
	default:
		return nil, mysqlerrors.Unknownf("unsupported type %T for result set", value)
	}
}

func formatField(variables *variable.Container, collationID uint16, field *Field,
	value evaluator.SQLValue) error {
	switch typedV := value.(type) {

	case evaluator.SQLFloat:
		field.Charset = collationID
		field.Type = MySQLTypeDouble
		field.Decimal = 0x1f
		field.Flag = BinaryFlag
	case evaluator.SQLDecimal128:
		field.Charset = collationID
		field.Type = MySQLTypeNewDecimal
		field.Decimal = 20      // scale
		field.ColumnLength = 67 // precision plus 2 (decimal point and length)
	case evaluator.SQLBool:
		field.Charset = collationID
		field.Type = MySQLTypeTiny
	case evaluator.SQLUint32:
		field.Charset = collationID
		field.Type = MySQLTypeLongLong
		field.Flag = BinaryFlag | UnsignedFlag
	case evaluator.SQLUint64:
		field.Charset = collationID
		field.Type = MySQLTypeLongLong
		field.Flag = BinaryFlag | UnsignedFlag
	case evaluator.SQLInt:
		field.Charset = collationID
		field.Type = MySQLTypeLongLong
		field.Flag = BinaryFlag
	case evaluator.SQLVarchar:
		field.Charset = collationID
		field.Type = MySQLTypeVarString

		length := uint32(variables.GetUInt16(variable.MongoDBMaxVarcharLength))
		if length == 0 {
			length = math.MaxUint16
		}

		field.ColumnLength = length
	case evaluator.SQLUUID:
		field.Charset = collationID
		field.Type = MySQLTypeVarString
		field.ColumnLength = 36 // 6B29FC40-CA47-1067-B31D-00DD010662DA
	case evaluator.SQLObjectID:
		field.Charset = collationID
		field.Type = MySQLTypeVarString
		field.ColumnLength = 24 // 582c98cdea11582c488616ee
	case nil, *evaluator.SQLNullValue, evaluator.SQLNullValue, evaluator.SQLNoValue:
		field.Charset = collationID
		field.Type = MySQLTypeNull
	case evaluator.SQLDate:
		field.Charset = collationID
		field.Type = MySQLTypeDate
	case evaluator.SQLTimestamp:
		field.Charset = collationID
		field.Type = MySQLTypeDatetime
	case *evaluator.SQLValues:
		if len(typedV.Values) != 1 {
			return mysqlerrors.Defaultf(mysqlerrors.ErOperandColumns, 1)
		}
		return formatField(variables, collationID, field, typedV.Values[0])
	default:
		return mysqlerrors.Unknownf("unsupported type %T for result set", value)
	}

	return nil
}

// writeHeaders writes the column headers for a resultset.
func (c *conn) writeHeaders(columns []*evaluator.Column, colID collation.ID) error {
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

	j := 0

	for j < numFields {
		zeroValue := columns[j].SQLType.ZeroValue()
		value, _ := evaluator.NewSQLValue(
			zeroValue,
			columns[j].SQLType,
			schema.SQLNone,
		)

		field := &Field{
			Name:          []byte(columns[j].Name),
			OriginalName:  []byte(columns[j].OriginalName),
			Schema:        []byte(columns[j].Database),
			Table:         []byte(columns[j].Table),
			OriginalTable: []byte(columns[j].OriginalTable),
		}

		err := formatField(c.variables, uint16(colID), field, value)
		if err != nil {
			return err
		}

		data = data[0:4]
		data = append(data, field.Dump(c.variables.GetCharset(variable.CharacterSetResults))...)

		// Write a column definition packet for each
		// column in the resultset.
		if err = c.writePacket(data); err != nil {
			return err
		}
		j++
	}
	// End the column definitions with an EOF packet.
	return c.writeEOF(status)
}

// streamRows receives packets from a packet producer across the packetChan and writes them to the
// conn. It is also responsible for closing the Iter once the packetChan is closed.
func (c *conn) streamRows(packetChan chan []byte, errChan chan error, columns []*evaluator.Column,
	iter evaluator.ErrCloser) (err error) {
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

		case <-c.Context().Done():
			return c.Context().Err()

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

	c.logger.Infof(log.Admin, "returned %d %s (%s)", count, util.Pluralize(count, "row", "rows"),
		util.ByteString(totalBytes))

	if err = c.Context().Err(); err != nil {
		return err
	}
	return c.writeEOF(status)
}

// sendPackets is used to produce packets from the Rows returned by
// the passed Iter. The results are passed as a []byte to the
// packetChan channel.
func (c *conn) sendPackets(packetChan chan []byte, iter evaluator.Iter) {
	r := &evaluator.Row{}
	ctx := c.Context()
	for ctx.Err() == nil && iter.Next(r) {
		packet := []byte{0, 0, 0, 0}
		for _, value := range r.Data {
			b, err := c.formatValue(value.Data)
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
func (c *conn) fastSendPackets(packetChan chan []byte, fastIter evaluator.FastIter) {
	columnInfo := fastIter.GetColumnInfo()
	lenColumnFields := len(columnInfo)
	ctx := c.Context()
	charSet := c.variables.GetCharset(variable.CharacterSetResults)
	mongoDBVarcharLength := int(c.variables.GetUInt16(variable.MongoDBMaxVarcharLength))

	doc := &bson.RawD{}

	maybePanic := func(err error) {
		if err != nil {
			close(packetChan)
			panic(err)
		}
	}

	for ctx.Err() == nil && fastIter.Next(doc) {
		packet := []byte{0, 0, 0, 0}
		values := *doc

		lenValues := len(values)
		if lenValues == lenColumnFields {
			// No missing values, so we can iterate fast without checking key names.
			for i, value := range values {
				columnType := columnInfo[i].Type
				uuidSubtype := columnInfo[i].UUIDSubtype
				df := newDataFormatter(columnType,
					schema.BSONSpecType(value.Value.Kind), uuidSubtype,
					charSet, mongoDBVarcharLength, value.Value.Data)
				b, err := fastFormat(df)
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
				if numMissingValues > 0 && i < len(values) {
					if fieldName == values[i].Name {
						value := values[i].Value
						// If this is the correct fieldName, output the value.
						df := newDataFormatter(columnType,
							schema.BSONSpecType(value.Kind), uuidSubtype,
							charSet, mongoDBVarcharLength, value.Data)
						b, err := fastFormat(df)
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
				} else if i < len(values) {
					// We have found all the missing values, default to the faster mode.
					value := values[i].Value
					i++
					df := newDataFormatter(columnType,
						schema.BSONSpecType(value.Kind), uuidSubtype,
						charSet, mongoDBVarcharLength, value.Data)
					b, err := fastFormat(df)
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

// fastSendPackets is used to produce packets from Documents returned from the
// passed fastIter. The results are returned as a []byte across the packetChan
// channel. The Documents returned by the fastIter do not have a guaranteed
// field order.
func (c *conn) fastSendPackets32(packetChan chan []byte, fastIter evaluator.FastIter) {
	columnInfo := fastIter.GetColumnInfo()
	ctx := c.Context()
	charSet := c.variables.GetCharset(variable.CharacterSetResults)
	mongoDBVarcharLength := int(c.variables.GetUInt16(variable.MongoDBMaxVarcharLength))

	fieldMap := make(map[string]*bson.Raw, len(columnInfo))
	doc := &bson.RawD{}
	for ctx.Err() == nil && fastIter.Next(doc) {
		packet := []byte{0, 0, 0, 0}
		values := *doc
		// Clear out previous fieldMap. We do not want to allocate
		// memory in a tight loop.
		for key := range fieldMap {
			delete(fieldMap, key)
		}
		// We can't rely on field ordering in 3.2.
		for i := range values {
			fieldMap[values[i].Name] = &(values[i].Value)
		}
		for _, info := range columnInfo {
			if value, ok := fieldMap[info.Field]; ok {
				df := dataFormatter{
					columnType:           info.Type,
					valueType:            schema.BSONSpecType(value.Kind),
					uuidSubtype:          info.UUIDSubtype,
					charSet:              charSet,
					mongoDBVarcharLength: mongoDBVarcharLength,
					data:                 value.Data,
				}
				b, err := fastFormat(df)
				if err != nil {
					close(packetChan)
					panic(err)
				}
				packet = append(packet, b...)
			} else {
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

// streamResultset implements COM_QUERY response.
// More at https://dev.mysql.com/doc/internals/en/com-query-response.html
//
// It uses a producer function defined by the type of iterator and possibly server
// version, which formats data to byte packets, and a consumer that actually
// writes those byte packets to the client. This producer consumer relation
// allows for query cancellation.
func (c *conn) streamResultset(columns []*evaluator.Column, iter evaluator.ErrCloser) (err error) {
	defer func() {
		if ctxErr := c.Context().Err(); ctxErr != nil {
			c.refreshContext()
			err = ctxErr
		}
	}()

	packetChan := make(chan []byte, 1)
	errChan := make(chan error, 1)

	errorHandler := func(err interface{}) {
		errChan <- fmt.Errorf("iterating error: %v", err)
	}

	if len(columns) == 0 {
		err = c.writeOK(nil)
	}

	c.affectedRows = int64(-1)

	var asyncPacketSender func()
	switch typedIter := iter.(type) {
	case evaluator.Iter:
		asyncPacketSender = func() { c.sendPackets(packetChan, typedIter) }
	case evaluator.FastIter:
		if c.Variables().MongoDBInfo.VersionAtLeast(3, 4, 0) {
			asyncPacketSender = func() { c.fastSendPackets(packetChan, typedIter) }
		} else {
			// For server < 3.4, we cannot rely on document field ordering.
			asyncPacketSender = func() { c.fastSendPackets32(packetChan, typedIter) }
		}

	}

	util.PanicSafeGo(asyncPacketSender, errorHandler)
	return c.streamRows(packetChan, errChan, columns, iter)
}
