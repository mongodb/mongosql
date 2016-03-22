package server

import (
	"encoding/binary"
)

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

func (p FieldData) Parse() (f *Field, err error) {
	f = new(Field)

	f.Data = p

	var n int
	pos := 0
	//skip catelog, always def
	n, err = skipLengthEncodedString(p)
	if err != nil {
		return
	}
	pos += n

	//schema
	f.Schema, _, n, err = lengthEncodedString(p[pos:])
	if err != nil {
		return
	}
	pos += n

	//table
	f.Table, _, n, err = lengthEncodedString(p[pos:])
	if err != nil {
		return
	}
	pos += n

	//org_table
	f.OrgTable, _, n, err = lengthEncodedString(p[pos:])
	if err != nil {
		return
	}
	pos += n

	//name
	f.Name, _, n, err = lengthEncodedString(p[pos:])
	if err != nil {
		return
	}
	pos += n

	//org_name
	f.OrgName, _, n, err = lengthEncodedString(p[pos:])
	if err != nil {
		return
	}
	pos += n

	//skip oc
	pos += 1

	//charset
	f.Charset = binary.LittleEndian.Uint16(p[pos:])
	pos += 2

	//column length
	f.ColumnLength = binary.LittleEndian.Uint32(p[pos:])
	pos += 4

	//type
	f.Type = p[pos]
	pos++

	//flag
	f.Flag = binary.LittleEndian.Uint16(p[pos:])
	pos += 2

	//decimals 1
	f.Decimal = p[pos]
	pos++

	//filter [0x00][0x00]
	pos += 2

	f.DefaultValue = nil
	//if more data, command was field list
	if len(p) > pos {
		//length of default value lenenc-int
		f.DefaultValueLength, _, n = lengthEncodedInt(p[pos:])
		pos += n

		if pos+int(f.DefaultValueLength) > len(p) {
			err = ErrMalformPacket
			return
		}

		//default value string[$len]
		f.DefaultValue = p[pos:(pos + int(f.DefaultValueLength))]
	}

	return
}

func (f *Field) Dump() []byte {
	if f.Data != nil {
		return []byte(f.Data)
	}

	l := len(f.Schema) + len(f.Table) + len(f.OrgTable) + len(f.Name) + len(f.OrgName) + len(f.DefaultValue) + 48

	data := make([]byte, 0, l)

	data = append(data, putLengthEncodedString([]byte("def"))...)

	data = append(data, putLengthEncodedString(f.Schema)...)

	data = append(data, putLengthEncodedString(f.Table)...)
	data = append(data, putLengthEncodedString(f.OrgTable)...)

	data = append(data, putLengthEncodedString(f.Name)...)
	data = append(data, putLengthEncodedString(f.OrgName)...)

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
