package bytesize

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// ByteSize represents size in bytes.
type ByteSize uint64

// Constants
const (
	Byte ByteSize = 1
	KiB           = Byte * 1024
	MiB           = KiB << 10
	GiB           = MiB << 10
	TiB           = GiB << 10
	PiB           = TiB << 10
	KB            = Byte * 1000
	MB            = KB * 1000
	GB            = MB * 1000
	TB            = GB * 1000
	PB            = TB * 1000
)

// Parse a string into a ByteSize.
func Parse(s string) (ByteSize, error) {
	var b ByteSize
	_, err := fmt.Sscan(s, &b)
	return b, err
}

// Bytes is the number of bytes.
func (b ByteSize) Bytes() uint64 {
	return uint64(b)
}

// String implements the fmt.Stringer interface. It
// returns the string form using binary units.
func (b ByteSize) String() string {
	return b.binaryString('f', -1)
}

func (b ByteSize) binaryString(fmt byte, precision int) string {
	switch {
	case b < KiB:
		return strconv.FormatFloat(float64(b), fmt, precision, 64) + " B"
	case b < MiB:
		return strconv.FormatFloat(b.KiB(), fmt, precision, 64) + " KiB"
	case b < GiB:
		return strconv.FormatFloat(b.MiB(), fmt, precision, 64) + " MiB"
	case b < TiB:
		return strconv.FormatFloat(b.GiB(), fmt, precision, 64) + " GiB"
	case b < PiB:
		return strconv.FormatFloat(b.TiB(), fmt, precision, 64) + " TiB"
	default:
		return strconv.FormatFloat(b.PiB(), fmt, precision, 64) + " PiB"
	}
}

// SIString returns the string form using SI units.
func (b ByteSize) SIString() string {
	return b.decimalString('f', -1)
}

func (b ByteSize) decimalString(fmt byte, precision int) string {
	switch {
	case b < KB:
		return strconv.FormatFloat(float64(b), 'f', precision, 64) + " B"
	case b < MB:
		return strconv.FormatFloat(b.KB(), fmt, precision, 64) + " KB"
	case b < GB:
		return strconv.FormatFloat(b.MB(), fmt, precision, 64) + " MB"
	case b < TB:
		return strconv.FormatFloat(b.GB(), fmt, precision, 64) + " GB"
	case b < PB:
		return strconv.FormatFloat(b.TB(), fmt, precision, 64) + " TB"
	default:
		return strconv.FormatFloat(b.PB(), fmt, precision, 64) + " PB"
	}
}

// Format implements the fmt.Formatter interface.
func (b ByteSize) Format(f fmt.State, c rune) {
	p, ok := f.Precision()
	if !ok {
		p = -1
	}

	switch c {
	case 'b', 'v':
		_, _ = f.Write([]byte(b.binaryString('f', p)))
	case 'd':
		_, _ = f.Write([]byte(b.decimalString('f', p)))
	default:
		_, _ = f.Write([]byte(fmt.Sprintf("ERROR: ByteSize unsupported fmt '%c'", c)))
	}
}

// Scan implements the fmt.Scanner interface.
func (b *ByteSize) Scan(state fmt.ScanState, verb rune) error {
	token, err := state.Token(false, func(r rune) bool {
		return unicode.IsDigit(r) || r == '.'
	})
	if err != nil {
		return err
	}

	value, err := strconv.ParseFloat(string(token), 64)
	if err != nil {
		return err
	}

	token, err = state.Token(true, unicode.IsLetter)
	if err != nil {
		return err
	}

	units := strings.ToLower(string(token))
	switch units {
	case "b", "":
		// TODO: throw an error if this is not a whole number?
		*b = ByteSize(value)
	case "kb":
		*b = ByteSize(value * float64(KB))
	case "mb":
		*b = ByteSize(value * float64(MB))
	case "gb":
		*b = ByteSize(value * float64(GB))
	case "tb":
		*b = ByteSize(value * float64(TB))
	case "pb":
		*b = ByteSize(value * float64(PB))
	case "kib":
		*b = ByteSize(value * float64(KiB))
	case "mib":
		*b = ByteSize(value * float64(MiB))
	case "gib":
		*b = ByteSize(value * float64(GiB))
	case "tib":
		*b = ByteSize(value * float64(TiB))
	case "pib":
		*b = ByteSize(value * float64(PiB))
	default:
		return fmt.Errorf("%s is an unrecognized unit", units)
	}

	return nil
}

// Idea taken from time package:
// These methods return float64 because the dominant
// use case is for printing a floating point number like 1.5s, and
// a truncation to integer would make them not useful in those cases.
// Splitting the integer and fraction ourselves guarantees that
// converting the returned float64 to an integer rounds the same
// way that a pure integer conversion would have, even in cases
// where, say, float64(d.Nanoseconds())/1e9 would have rounded
// differently.

// KiB is the number of kibibytes.
func (b ByteSize) KiB() float64 {
	w := b / KiB
	r := b % KiB
	return float64(w) + float64(r)/float64(KiB)
}

// MiB is the number of mebibytes.
func (b ByteSize) MiB() float64 {
	w := b / MiB
	r := b % MiB
	return float64(w) + float64(r)/float64(MiB)
}

// GiB is the number of gibibytes.
func (b ByteSize) GiB() float64 {
	w := b / GiB
	r := b % GiB
	return float64(w) + float64(r)/float64(GiB)
}

// TiB is the number of tebibytes.
func (b ByteSize) TiB() float64 {
	w := b / TiB
	r := b % TiB
	return float64(w) + float64(r)/float64(TiB)
}

// PiB is the number of pebibytes.
func (b ByteSize) PiB() float64 {
	w := b / PiB
	r := b % PiB
	return float64(w) + float64(r)/float64(PiB)
}

// KB is the number of kilobytes.
func (b ByteSize) KB() float64 {
	w := b / KB
	r := b % KB
	return float64(w) + float64(r)/float64(KB)
}

// MB is the number of megabytes.
func (b ByteSize) MB() float64 {
	w := b / MB
	r := b % MB
	return float64(w) + float64(r)/float64(MB)
}

// GB is the number of gigabytes.
func (b ByteSize) GB() float64 {
	w := b / GB
	r := b % GB
	return float64(w) + float64(r)/float64(GB)
}

// TB is the number of terabytes.
func (b ByteSize) TB() float64 {
	w := b / TB
	r := b % TB
	return float64(w) + float64(r)/float64(TB)
}

// PB is the number of petabytes.
func (b ByteSize) PB() float64 {
	w := b / PB
	r := b % PB
	return float64(w) + float64(r)/float64(PB)
}
