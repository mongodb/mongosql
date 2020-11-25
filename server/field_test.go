package server

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/10gen/sqlproxy/collation"
	"github.com/stretchr/testify/require"
)

func TestDump(t *testing.T) {
	req := require.New(t)

	f := &Field{
		Schema:             []byte("INFORMATION_SCHEMA"),
		Table:              []byte("foo"),
		OriginalTable:      []byte("ofoo"),
		Name:               []byte("bar"),
		OriginalName:       []byte("obar"),
		Charset:            uint16(45),    // utf8mb4_general_ci
		ColumnLength:       uint32(65535), // 0x0000FFFF
		Type:               uint8(253),    // 0xFD - MYSQL_TYPE_VAR_STRING
		Flag:               uint16(0),
		Decimal:            uint8(0),
		DefaultValueLength: uint64(0),
		DefaultValue:       nil,
	}

	charset, err := collation.GetCharset("utf8mb4")
	req.Nilf(err, "unable to get charset")

	payloadString := "03 64 65 66" + // def
		"12 49 4E 46 4F 52 4D 41 54 49 4F 4E 5F 53 43 48 45 4D 41" + // INFORMATION_SCHEMA
		"03 66 6F 6F" + // foo
		"04 6F 66 6F 6F" + // ofoo
		"03 62 61 72" + // bar
		"04 6F 62 61 72" + // obar
		"0C" + // length of fixed-length fields (always 0x0C)
		"2D 00" + // charset
		"FF FF 00 00" + // column length
		"FD" + // type
		"00 00" + // flags
		"00" + // decimals
		"00 00" // filler

	regularPayload, err := hex.DecodeString(strings.ReplaceAll(payloadString, " ", ""))
	req.Nilf(err, "unable to decode payload string")

	regularDump := f.Dump(charset, false)
	req.Equalf(regularDump, regularPayload, "unexpected field dump")

	comFieldListPayload := append(regularPayload, 0xfb)

	comFieldListDump := f.Dump(charset, true)
	req.Equalf(comFieldListDump, comFieldListPayload, "unexpected field dump for COM_FIELD_LIST")
}
