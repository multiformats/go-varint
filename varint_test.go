package varint

import (
	"encoding/binary"
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

func TestUnderflow(t *testing.T) {
	i, n, err := FromUvarint([]byte{0x80})
	if err == nil {
		t.Error("expected an error")
	}
	if n != 0 {
		t.Error("expected n = 0")
	}
	if i != 0 {
		t.Error("expected i = 0")
	}
}
