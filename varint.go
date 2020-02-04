package varint

import (
	"encoding/binary"
	"errors"
	"io"
	"math/bits"
)

var (
	ErrOverflow   = errors.New("varints larger than uint64 not yet supported")
	ErrUnderflow  = errors.New("varints malformed, could not reach the end")
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
	var x uint64
	var s uint
	for i, b := range buf {
		if b < 0x80 {
			if i > 9 || i == 9 && b > 1 {
				return 0, 0, ErrOverflow
			}
			return x | uint64(b)<<s, i + 1, nil
		} else if b == 0x80 && x == 0 {
			return 0, 0, ErrNotMinimal
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
	return 0, 0, ErrUnderflow
}

// ReadUvarint reads a unsigned varint from the given reader.
func ReadUvarint(r io.ByteReader) (uint64, error) {
	var x uint64
	var s uint
	for i := 0; ; i++ {
		b, err := r.ReadByte()
		switch err {
		case nil:
		case io.EOF:
			// "eof" will look like a success.
			return 0, io.ErrUnexpectedEOF
		default:
			return 0, err
		}
		if b < 0x80 {
			if i > 9 || i == 9 && b > 1 {
				return 0, ErrOverflow
			}
			return x | uint64(b)<<s, nil
		} else if b == 0x80 && x == 0 {
			return 0, ErrNotMinimal
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
}

// PutUvarint is an alias for binary.PutUvarint.
//
// This is provided for convenience so users of this library can avoid built-in
// varint functions and easily audit code for uses of those functions.
func PutUvarint(buf []byte, x uint64) int {
	return binary.PutUvarint(buf, x)
}
