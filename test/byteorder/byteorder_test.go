package byteorder_test

import (
	"testing"

	"modbus-sim/internal/byteorder"
	"modbus-sim/internal/config"
)

// TestResolve tests that all byte orders can be resolved.
func TestResolve(t *testing.T) {
	tests := []struct {
		name      string
		byteOrder config.ByteOrder
		wantErr   bool
	}{
		{"ABCD_BigEndian", config.BigEndian, false},
		{"DCBA_LittleEndian", config.LittleEndian, false},
		{"BADC_BigEndianSwap", config.BigEndianSwap, false},
		{"CDAB_LittleEndianSwap", config.LittleEndianSwap, false},
		{"BDAC_MidSwap", config.MidSwap, false},
		{"Invalid", config.ByteOrder("INVALID"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := byteorder.Resolve(tt.byteOrder)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("Resolve() returned nil for valid byte order")
			}
		})
	}
}

// TestByteOrderRoundTrip tests that encoding and decoding are inverse operations.
func TestByteOrderRoundTrip(t *testing.T) {
	orders := []struct {
		name config.ByteOrder
	}{
		{"ABCD"}, {"DCBA"}, {"BADC"}, {"CDAB"}, {"BDAC"},
	}

	testValues32 := []uint32{0x00000000, 0xFFFFFFFF, 0x12345678, 0xAABBCCDD, 0x01020304}
	testValues64 := []uint64{0x0000000000000000, 0xFFFFFFFFFFFFFFFF, 0x0123456789ABCDEF}

	for _, order := range orders {
		orderName := string(order.name)
		t.Run(orderName+"_Uint32", func(t *testing.T) {
			bo, err := byteorder.Resolve(order.name)
			if err != nil {
				t.Fatalf("Resolve(%s) failed: %v", orderName, err)
			}
			for _, v := range testValues32 {
				buf := make([]byte, 4)
				bo.PutUint32(buf, v)
				decoded := bo.Uint32(buf)
				if decoded != v {
					t.Errorf("PutUint32/Uint32 roundtrip failed for %s: got 0x%08X, want 0x%08X",
						orderName, decoded, v)
				}
			}
		})

		t.Run(orderName+"_Uint64", func(t *testing.T) {
			bo, err := byteorder.Resolve(order.name)
			if err != nil {
				t.Fatalf("Resolve(%s) failed: %v", orderName, err)
			}
			for _, v := range testValues64 {
				buf := make([]byte, 8)
				bo.PutUint64(buf, v)
				decoded := bo.Uint64(buf)
				if decoded != v {
					t.Errorf("PutUint64/Uint64 roundtrip failed for %s: got 0x%016X, want 0x%016X",
						orderName, decoded, v)
				}
			}
		})
	}
}

// TestBDACSpecific tests the BDAC byte order encoding specifically.
func TestBDACSpecific(t *testing.T) {
	bo, err := byteorder.Resolve(config.MidSwap)
	if err != nil {
		t.Fatalf("Resolve(MidSwap) failed: %v", err)
	}

	// For value 0x12345678, bytes in big-endian are [0x12][0x34][0x56][0x78]
	// BDAC should produce: [0x34][0x78][0x12][0x56] (swap A,B and C,D)
	buf := make([]byte, 4)
	bo.PutUint32(buf, 0x12345678)

	expected := []byte{0x34, 0x78, 0x12, 0x56}
	for i, b := range expected {
		if buf[i] != b {
			t.Errorf("BDAC PutUint32: buf[%d] = 0x%02X, want 0x%02X", i, buf[i], b)
		}
	}

	// Verify roundtrip
	decoded := bo.Uint32(buf)
	if decoded != 0x12345678 {
		t.Errorf("BDAC roundtrip: got 0x%08X, want 0x12345678", decoded)
	}
}

// TestConfigByteOrderValidation tests byte order validation in config.
func TestConfigByteOrderValidation(t *testing.T) {
	valid := []config.ByteOrder{
		config.BigEndian, config.LittleEndian,
		config.BigEndianSwap, config.LittleEndianSwap, config.MidSwap,
	}
	invalid := config.ByteOrder("XZYW")

	for _, b := range valid {
		if !b.IsValid() {
			t.Errorf("ByteOrder %q should be valid", b)
		}
	}

	if invalid.IsValid() {
		t.Error("ByteOrder XZYW should be invalid")
	}
}
