package varint

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
)

func checkVarint(t *testing.T, x uint64) {
	buf := make([]byte, binary.MaxVarintLen64)
	expected := binary.PutUvarint(buf, x)

	size := UvarintSize(x)
	if size != expected {
		t.Fatalf("expected varintsize of %d to be %d, got %d", x, expected, size)
	}
}

func TestVarintSize(t *testing.T) {
	var max uint64 = 1 << 16
	for x := uint64(0); x < max; x++ {
		checkVarint(t, x)
	}
}

func TestOverflow(t *testing.T) {
	i, n, err := FromUvarint(
		[]byte{
			0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0x00,
		},
	)
	if err != ErrOverflow {
		t.Error("expected an error")
	}
	if n != 0 {
		t.Error("expected n = 0")
	}
	if i != 0 {
		t.Error("expected i = 0")
	}
}

func TestNotMinimal(t *testing.T) {
	i, n, err := FromUvarint([]byte{0x80, 0x01})
	if err != ErrNotMinimal {
		t.Error("expected an error")
	}
	if n != 0 {
		t.Error("expected n = 0")
	}
	if i != 0 {
		t.Error("expected i = 0")
	}
}

func TestUnderflow(t *testing.T) {
	i, n, err := FromUvarint([]byte{0x81, 0x81})
	if err != ErrUnderflow {
		t.Error("expected an underflow")
	}
	if n != 0 {
		t.Error("expected n = 0")
	}
	if i != 0 {
		t.Error("expected i = 0")
	}
}

func TestEOF(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	n, err := ReadUvarint(buf)
	if err != io.EOF {
		t.Error("expected an EOF")
	}
	if n != 0 {
		t.Error("expected n = 0")
	}
}

func TestUnexpectedEOF(t *testing.T) {
	buf := bytes.NewBuffer([]byte{0x81, 0x81})
	n, err := ReadUvarint(buf)
	if err != io.ErrUnexpectedEOF {
		t.Error("expected an EOF")
	}
	if n != 0 {
		t.Error("expected n = 0")
	}
}
