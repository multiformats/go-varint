package varint

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"math/bits"
	"testing"
)

func checkVarint(t *testing.T, x uint64) {
	buf := make([]byte, binary.MaxVarintLen64)
	expected := binary.PutUvarint(buf, x)

	size := UvarintSize(x)
	if size != expected {
		t.Fatalf("expected varintsize of %d to be %d, got %d", x, expected, size)
	}
	xi, n, err := FromUvarint(buf)
	if err != nil {
		t.Fatal("decoding error", err)
	}
	if n != size {
		t.Fatal("read the wrong size")
	}
	if xi != x {
		t.Fatal("expected a different result")
	}
}

func TestVarintSize(t *testing.T) {
	var max uint64 = 1 << 16
	for x := uint64(0); x < max; x++ {
		checkVarint(t, x)
	}
}

func TestOverflow_9thSignalsMore(t *testing.T) {
	buf := bytes.NewBuffer([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0x80,
	})

	_, err := ReadUvarint(buf)
	if err != ErrOverflow {
		t.Fatalf("expected ErrOverflow, got: %s", err)
	}
}

func TestOverflow_ReadBuffer(t *testing.T) {
	buf := bytes.NewBuffer([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
	})

	_, err := ReadUvarint(buf)
	if err != ErrOverflow {
		t.Fatalf("expected ErrOverflow, got: %s", err)
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
	varint := []byte{0x81, 0x00}
	i, n, err := FromUvarint(varint)
	if err != ErrNotMinimal {
		t.Error("expected an error")
	}
	if n != 0 {
		t.Error("expected n = 0")
	}
	if i != 0 {
		t.Error("expected i = 0")
	}
	i, n = binary.Uvarint(varint)
	if n != len(varint) {
		t.Error("expected to read entire buffer")
	}
	if i != 1 {
		t.Error("expected varint 1")
	}
}

func TestNotMinimalRead(t *testing.T) {
	varint := bytes.NewBuffer([]byte{0x81, 0x00})
	i, err := ReadUvarint(varint)
	if err != ErrNotMinimal {
		t.Error("expected an error")
	}
	if i != 0 {
		t.Error("expected i = 0")
	}
	varint = bytes.NewBuffer([]byte{0x81, 0x00})
	i, err = binary.ReadUvarint(varint)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if i != 1 {
		t.Error("expected varint 1")
	}
	if err != nil {
		t.Fatal(err)
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

func BenchmarkReadUvarint(t *testing.B) {
	var expected uint64 = 0xffff12
	reader := bytes.NewReader(ToUvarint(expected))
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		result, _ := ReadUvarint(reader)
		if result != expected {
			t.Fatal("invalid result")
		}
		reader.Seek(0, 0)
	}
}

func BenchmarkFromUvarint(t *testing.B) {
	var expected uint64 = 0xffff12
	uvarint := ToUvarint(expected)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		result, _, _ := FromUvarint(uvarint)
		if result != expected {
			t.Fatal("invalid result")
		}
	}
}

// uvarintSizeReference preserves the original implementation for testing and benchmarking
func uvarintSizeReference(num uint64) int {
	bits := bits.Len64(num)
	q, r := bits/7, bits%7
	size := q
	if r > 0 || size == 0 {
		size++
	}
	return size
}

func TestUvarintSizeEdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		value    uint64
		expected int
	}{
		{"zero", 0, 1},

		// Single byte values (0-127)
		{"one", 1, 1},
		{"max_single_byte", 127, 1},

		// Two byte values (128-16383)
		{"min_two_bytes", 128, 2},
		{"max_two_bytes", 16383, 2},

		// Boundary values for each byte count
		{"boundary_1_to_2", (1 << 7) - 1, 1},
		{"boundary_2_start", 1 << 7, 2},
		{"boundary_2_to_3", (1 << 14) - 1, 2},
		{"boundary_3_start", 1 << 14, 3},
		{"boundary_3_to_4", (1 << 21) - 1, 3},
		{"boundary_4_start", 1 << 21, 4},
		{"boundary_4_to_5", (1 << 28) - 1, 4},
		{"boundary_5_start", 1 << 28, 5},
		{"boundary_5_to_6", (1 << 35) - 1, 5},
		{"boundary_6_start", 1 << 35, 6},
		{"boundary_6_to_7", (1 << 42) - 1, 6},
		{"boundary_7_start", 1 << 42, 7},
		{"boundary_7_to_8", (1 << 49) - 1, 7},
		{"boundary_8_start", 1 << 49, 8},
		{"boundary_8_to_9", (1 << 56) - 1, 8},
		{"boundary_9_start", 1 << 56, 9},
		{"boundary_9_to_10", (1 << 63) - 1, 9},
		{"boundary_10_start", 1 << 63, 10},

		// Maximum values
		{"max_uint64", math.MaxUint64, 10},
		{"max_uint64_minus_1", math.MaxUint64 - 1, 10},

		// Powers of 2
		{"power_2_0", 1 << 0, 1},
		{"power_2_6", 1 << 6, 1},
		{"power_2_7", 1 << 7, 2},
		{"power_2_8", 1 << 8, 2},
		{"power_2_13", 1 << 13, 2},
		{"power_2_14", 1 << 14, 3},
		{"power_2_20", 1 << 20, 3},
		{"power_2_21", 1 << 21, 4},
		{"power_2_27", 1 << 27, 4},
		{"power_2_28", 1 << 28, 5},
		{"power_2_34", 1 << 34, 5},
		{"power_2_35", 1 << 35, 6},
		{"power_2_41", 1 << 41, 6},
		{"power_2_42", 1 << 42, 7},
		{"power_2_48", 1 << 48, 7},
		{"power_2_49", 1 << 49, 8},
		{"power_2_55", 1 << 55, 8},
		{"power_2_56", 1 << 56, 9},
		{"power_2_62", 1 << 62, 9},
		{"power_2_63", 1 << 63, 10},

		// Special patterns
		{"all_ones_32bit", 0xFFFFFFFF, 5},
		{"all_ones_48bit", 0xFFFFFFFFFFFF, 7},
		{"alternating_pattern", 0xAAAAAAAAAAAAAAAA, 10},
		{"alternating_pattern2", 0x5555555555555555, 9},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test against the original reference implementation
			original := uvarintSizeReference(tc.value)
			if original != tc.expected {
				t.Errorf("Original implementation wrong for %s (%d): got %d, expected %d",
					tc.name, tc.value, original, tc.expected)
			}

			actualSize := UvarintSize(tc.value)
			if actualSize != tc.expected {
				t.Errorf("Optimised implementation wrong for %s (%d): got %d, expected %d",
					tc.name, tc.value, actualSize, tc.expected)
			}

			// Ensure both implementations agree
			if original != actualSize {
				t.Errorf("Implementations differ for %s (%d): original=%d, actualSize=%d",
					tc.name, tc.value, original, actualSize)
			}

			// Verify against actual encoding length
			buf := make([]byte, 10)
			n := binary.PutUvarint(buf, tc.value)
			if n != tc.expected {
				t.Errorf("Actual encoding length differs for %s (%d): got %d, expected %d",
					tc.name, tc.value, n, tc.expected)
			}

			// Double-check that our UvarintSize matches actual encoding
			if actualSize != n {
				t.Errorf("UvarintSize doesn't match actual encoding for %s (%d): UvarintSize=%d, actual=%d",
					tc.name, tc.value, actualSize, n)
			}
		})
	}
}

// TestUvarintSizeExhaustive tests the first million values exhaustively
func TestUvarintSizeExhaustive(t *testing.T) {
	for i := uint64(0); i < 1000000; i++ {
		original := uvarintSizeReference(i)
		actualSize := UvarintSize(i)

		if original != actualSize {
			t.Errorf("Mismatch at %d: original=%d, actualSize=%d", i, original, actualSize)
		}

		// Also verify against actual encoding
		buf := make([]byte, 10)
		n := binary.PutUvarint(buf, i)
		if actualSize != n {
			t.Errorf("UvarintSize doesn't match actual encoding at %d: UvarintSize=%d, actual=%d",
				i, actualSize, n)
		}
	}
}

// TestUvarintSizeRandom tests random values across the entire uint64 range
func TestUvarintSizeRandom(t *testing.T) {
	// Test values distributed across different ranges
	testValues := []uint64{
		// Random values in each byte-size range
		42, 100, 126, // 1 byte
		200, 1000, 10000, 16000, // 2 bytes
		20000, 100000, 1000000, 2000000, // 3 bytes
		10000000, 100000000, 200000000, // 4 bytes
		1000000000, 10000000000, 30000000000, // 5 bytes
		100000000000, 1000000000000, // 6 bytes
		10000000000000, 100000000000000, // 7 bytes
		1000000000000000, 10000000000000000, // 8 bytes
		100000000000000000, 1000000000000000000, // 9 bytes
		10000000000000000000, // 10 bytes
	}

	// Add more test values using bit shifting for better coverage
	for shift := uint(0); shift < 64; shift++ {
		base := uint64(1) << shift
		testValues = append(testValues, base)
		if base > 2 {
			testValues = append(testValues, base-1)
			testValues = append(testValues, base-2)
		}
		if base < math.MaxUint64-2 {
			testValues = append(testValues, base+1)
			testValues = append(testValues, base+2)
		}
	}

	for _, v := range testValues {
		original := uvarintSizeReference(v)
		actualSize := UvarintSize(v)

		if original != actualSize {
			t.Errorf("Mismatch for %d: original=%d, actualSize=%d", v, original, actualSize)
		}

		// Verify against actual encoding
		buf := make([]byte, 10)
		n := binary.PutUvarint(buf, v)
		if actualSize != n {
			t.Errorf("UvarintSize doesn't match actual encoding for %d: UvarintSize=%d, actual=%d",
				v, actualSize, n)
		}
	}
}

// Benchmark test data
var benchmarkTestVals = func() []uint64 {
	vals := []uint64{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		55, 66, 77, 88, 99, 100,
		127, 128, 255, 256,
		1<<7 - 1, 1 << 7, 1<<14 - 1, 1 << 14,
		1<<21 - 1, 1 << 21, 1<<28 - 1, 1 << 28,
		1<<35 - 1, 1 << 35, 1<<42 - 1, 1 << 42,
		1<<49 - 1, 1 << 49, 1<<56 - 1, 1 << 56,
		1<<63 - 1, 1 << 63,
		123456789, 98765432,
		0xFFFFFFFF, 0xFFFFFFFFFFFFFFFF,
	}
	// Repeat values to get a larger dataset for stable benchmarks
	newslice := make([]uint64, 100*len(vals))
	n := copy(newslice, vals)
	for n < len(newslice) {
		n += copy(newslice[n:], newslice[:n])
	}
	return newslice
}()

func BenchmarkUvarintSizeOriginal(b *testing.B) {
	var total int
	for i := 0; i < b.N; i++ {
		for _, val := range benchmarkTestVals {
			total += uvarintSizeReference(val)
		}
	}
	// Prevent compiler optimisation
	if total == 0 {
		b.Fatal("unexpected")
	}
}

func BenchmarkUvarintSizeCurrent(b *testing.B) {
	var total int
	for i := 0; i < b.N; i++ {
		for _, val := range benchmarkTestVals {
			total += UvarintSize(val)
		}
	}
	// Prevent compiler optimisation
	if total == 0 {
		b.Fatal("unexpected")
	}
}
