package server

import "github.com/10gen/sqlproxy/collation"

type FieldData []byte

type Field struct {
	Data         FieldData
	Schema       []byte
	Table        []byte
	OrgTable     []byte
	Name         []byte
	OrgName      []byte
	Charset      uint16
	ColumnLength uint32
	Type         uint8
	Flag         uint16
	Decimal      uint8

	DefaultValueLength uint64
	DefaultValue       []byte
}

func (f *Field) Dump(cs *collation.Charset) []byte {
	if f.Data != nil {
		return []byte(f.Data)
	}

	l := len(f.Schema) + len(f.Table) + len(f.OrgTable) + len(f.Name) + len(f.OrgName) + len(f.DefaultValue) + 48

	data := make([]byte, 0, l)

	data = append(data, putLengthEncodedString(cs.Decode([]byte("def")))...)

	data = append(data, putLengthEncodedString(cs.Decode(f.Schema))...)

	data = append(data, putLengthEncodedString(cs.Decode(f.Table))...)
	data = append(data, putLengthEncodedString(cs.Decode(f.OrgTable))...)

	data = append(data, putLengthEncodedString(cs.Decode(f.Name))...)
	data = append(data, putLengthEncodedString(cs.Decode(f.OrgName))...)

	data = append(data, 0x0c)

	data = append(data, uint16ToBytes(f.Charset)...)
	data = append(data, uint32ToBytes(f.ColumnLength)...)
	data = append(data, f.Type)
	data = append(data, uint16ToBytes(f.Flag)...)
	data = append(data, f.Decimal)
	data = append(data, 0, 0)

	if f.DefaultValue != nil {
		data = append(data, uint64ToBytes(f.DefaultValueLength)...)
		data = append(data, f.DefaultValue...)
	}

	return data
}
