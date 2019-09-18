package varint

import (
	"encoding/binary"
	"errors"
	"math/bits"
)

var (
	ErrOverflow   = errors.New("varints larger than uint64 not yet supported")
	ErrNotMinimal = errors.New("varint not minimally encoded")
)

// UvarintSize returns the size (in bytes) of `num` encoded as a unsigned varint.
func UvarintSize(num uint64) int {
	bits := bits.Len64(num)
	q, r := bits/7, bits%7
	size := q
	if r > 0 || size == 0 {
		size++
	}
	return size
}

// ToUvarint converts an unsigned integer to a varint-encoded []byte
func ToUvarint(num uint64) []byte {
	buf := make([]byte, UvarintSize(num))
	n := binary.PutUvarint(buf, uint64(num))
	return buf[:n]
}

// FromUvarint reads an unsigned varint from the beginning of buf, returns the
// varint, and the number of bytes read.
func FromUvarint(buf []byte) (uint64, int, error) {
	num, n := binary.Uvarint(buf)
	if n < 0 {
		return 0, 0, ErrOverflow
	}
	if n > UvarintSize(num) {
		return 0, 0, ErrNotMinimal
	}
	return num, n, nil
}
