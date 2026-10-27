// Package ids is the engine's one ID scheme (Phase 4b decisions 10 and 11).
// An ID is the cell key of the room or region where a thing was created and a
// random 64-bit part, written 0x<cell>:0x<random> in lowercase hex, in the
// style of Google's feature IDs. An ID is made once, when a thing is created
// or instantiated, and kept for good; internal/guards lists the places allowed
// to make one.
package ids

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strconv"
)

// NoCell is the cell of a thing created where there is no geography yet.
// Every room uses it until the geo room key gives rooms a cell.
const NoCell uint64 = 0

// ID identifies one thing for good. The zero ID means none.
type ID struct {
	cell uint64
	rand uint64
}

// New makes the ID for a thing created now in cell. The random part is never
// zero, so a made ID is never the zero ID.
func New(cell uint64) ID {
	var b [8]byte
	for {
		rand.Read(b[:]) // crypto/rand.Read never returns an error (Go 1.24+)
		if r := binary.BigEndian.Uint64(b[:]); r != 0 {
			return ID{cell: cell, rand: r}
		}
	}
}

// IsZero reports whether id is the zero ID (no thing).
func (id ID) IsZero() bool { return id == ID{} }

// Cell is the cell key the thing was created in.
func (id ID) Cell() uint64 { return id.cell }

// String is the text form, 0x<16 hex>:0x<16 hex>, or "" for the zero ID.
func (id ID) String() string {
	if id.IsZero() {
		return ""
	}
	return fmt.Sprintf("0x%016x:0x%016x", id.cell, id.rand)
}

// Parse reads the text form exactly: both halves 0x-prefixed, 16 lowercase
// hex digits each, and a non-zero random half.
func Parse(s string) (ID, error) {
	if len(s) != 37 || s[0:2] != "0x" || s[18] != ':' || s[19:21] != "0x" {
		return ID{}, fmt.Errorf("ids: %q is not 0x<16 hex>:0x<16 hex>", s)
	}
	cell, err1 := parseHex(s[2:18])
	r, err2 := parseHex(s[21:37])
	if err1 != nil || err2 != nil || r == 0 {
		return ID{}, fmt.Errorf("ids: %q is not 0x<16 hex>:0x<16 hex>", s)
	}
	return ID{cell: cell, rand: r}, nil
}

func parseHex(s string) (uint64, error) {
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return 0, fmt.Errorf("not lowercase hex")
		}
	}
	return strconv.ParseUint(s, 16, 64)
}

func (id ID) MarshalText() ([]byte, error) { return []byte(id.String()), nil }

func (id *ID) UnmarshalText(b []byte) error {
	if len(b) == 0 {
		*id = ID{}
		return nil
	}
	got, err := Parse(string(b))
	if err != nil {
		return err
	}
	*id = got
	return nil
}
