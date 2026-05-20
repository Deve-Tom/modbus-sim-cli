// Package byteorder provides byte order encoding/decoding for Modbus register values.
// It supports five byte orders: ABCD (BigEndian), DCBA (LittleEndian),
// BADC (BigEndianSwap), CDAB (LittleEndianSwap), and BDAC (MidSwap).
package byteorder

import (
	"encoding/binary"
	"fmt"

	"modbus-sim/internal/config"
)

// Order defines the interface for byte order encoding and decoding.
type Order interface {
	// PutUint32 encodes a uint32 value into a 4-byte buffer.
	PutUint32(buf []byte, v uint32)

	// PutUint64 encodes a uint64 value into an 8-byte buffer.
	PutUint64(buf []byte, v uint64)

	// Uint32 decodes a uint32 value from a 4-byte buffer.
	Uint32(buf []byte) uint32

	// Uint64 decodes a uint64 value from an 8-byte buffer.
	Uint64(buf []byte) uint64
}

// bigEndian uses standard big-endian byte order (ABCD).
type bigEndian struct{}

func (o *bigEndian) PutUint32(buf []byte, v uint32) {
	binary.BigEndian.PutUint32(buf, v)
}

func (o *bigEndian) PutUint64(buf []byte, v uint64) {
	binary.BigEndian.PutUint64(buf, v)
}

func (o *bigEndian) Uint32(buf []byte) uint32 {
	return binary.BigEndian.Uint32(buf)
}

func (o *bigEndian) Uint64(buf []byte) uint64 {
	return binary.BigEndian.Uint64(buf)
}

// littleEndian uses standard little-endian byte order (DCBA).
type littleEndian struct{}

func (o *littleEndian) PutUint32(buf []byte, v uint32) {
	binary.LittleEndian.PutUint32(buf, v)
}

func (o *littleEndian) PutUint64(buf []byte, v uint64) {
	binary.LittleEndian.PutUint64(buf, v)
}

func (o *littleEndian) Uint32(buf []byte) uint32 {
	return binary.LittleEndian.Uint32(buf)
}

func (o *littleEndian) Uint64(buf []byte) uint64 {
	return binary.LittleEndian.Uint64(buf)
}

// bigEndianSwap uses big-endian with word swap (BADC).
// For a 32-bit value [A][B][C][D] -> [B][A][D][C]
type bigEndianSwap struct{}

func (o *bigEndianSwap) PutUint32(buf []byte, v uint32) {
	binary.BigEndian.PutUint32(buf, v)
	buf[0], buf[1] = buf[1], buf[0]
	buf[2], buf[3] = buf[3], buf[2]
}

func (o *bigEndianSwap) PutUint64(buf []byte, v uint64) {
	binary.BigEndian.PutUint64(buf, v)
	buf[0], buf[1] = buf[1], buf[0]
	buf[2], buf[3] = buf[3], buf[2]
	buf[4], buf[5] = buf[5], buf[4]
	buf[6], buf[7] = buf[7], buf[6]
}

func (o *bigEndianSwap) Uint32(buf []byte) uint32 {
	b := []byte{buf[1], buf[0], buf[3], buf[2]}
	return binary.BigEndian.Uint32(b)
}

func (o *bigEndianSwap) Uint64(buf []byte) uint64 {
	b := []byte{buf[1], buf[0], buf[3], buf[2], buf[5], buf[4], buf[7], buf[6]}
	return binary.BigEndian.Uint64(b)
}

// littleEndianSwap uses little-endian with word swap (CDAB).
// For a 32-bit value [A][B][C][D] -> [C][D][A][B]
type littleEndianSwap struct{}

func (o *littleEndianSwap) PutUint32(buf []byte, v uint32) {
	binary.LittleEndian.PutUint32(buf, v)
	buf[0], buf[1] = buf[1], buf[0]
	buf[2], buf[3] = buf[3], buf[2]
}

func (o *littleEndianSwap) PutUint64(buf []byte, v uint64) {
	binary.LittleEndian.PutUint64(buf, v)
	buf[0], buf[1] = buf[1], buf[0]
	buf[2], buf[3] = buf[3], buf[2]
	buf[4], buf[5] = buf[5], buf[4]
	buf[6], buf[7] = buf[7], buf[6]
}

func (o *littleEndianSwap) Uint32(buf []byte) uint32 {
	// CDAB encoding: LE bytes + swap adjacent pairs -> [C][D][A][B]
	// CDAB decoding: swap adjacent pairs + LE interpretation -> original
	// To decode, we swap adjacent pairs of the encoded bytes, then interpret as LE
	b := []byte{buf[2], buf[3], buf[0], buf[1]}
	return binary.BigEndian.Uint32(b)
}

func (o *littleEndianSwap) Uint64(buf []byte) uint64 {
	// CDAB encoding applies swaps in order: (0,1), (2,3), (4,5), (6,7)
	// Decoding should reverse: (6,7), (4,5), (2,3), (0,1)
	b := []byte{buf[6], buf[7], buf[4], buf[5], buf[2], buf[3], buf[0], buf[1]}
	return binary.BigEndian.Uint64(b)
}

// midSwap - BDAC
// For a 32-bit value with bytes [A][B][C][D]:
//
//	After BDAC encoding: [B][D][A][C]
type midSwap struct{}

// PutUint32 encodes a uint32 value using BDAC byte order.
func (o *midSwap) PutUint32(buf []byte, v uint32) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	// [A][B][C][D] -> [B][D][A][C]
	buf[0] = b[1] // B
	buf[1] = b[3] // D
	buf[2] = b[0] // A
	buf[3] = b[2] // C
}

// PutUint64 encodes a uint64 value using BDAC byte order.
func (o *midSwap) PutUint64(buf []byte, v uint64) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	// [A][B][C][D][E][F][G][H] -> [B][D][A][C][F][H][E][G]
	buf[0] = b[1]
	buf[1] = b[3]
	buf[2] = b[0]
	buf[3] = b[2]
	buf[4] = b[5]
	buf[5] = b[7]
	buf[6] = b[4]
	buf[7] = b[6]
}

// Uint32 decodes a uint32 value from a BDAC-encoded buffer.
func (o *midSwap) Uint32(buf []byte) uint32 {
	// Reverse: [B][D][A][C] -> [A][B][C][D]
	b := []byte{buf[2], buf[0], buf[3], buf[1]}
	return binary.BigEndian.Uint32(b)
}

// Uint64 decodes a uint64 value from a BDAC-encoded buffer.
func (o *midSwap) Uint64(buf []byte) uint64 {
	// Reverse: [B][D][A][C][F][H][E][G] -> [A][B][C][D][E][F][G][H]
	b := []byte{buf[2], buf[0], buf[3], buf[1], buf[6], buf[4], buf[7], buf[5]}
	return binary.BigEndian.Uint64(b)
}

// Resolve returns the appropriate Order implementation for the given ByteOrder constant.
func Resolve(byteOrder config.ByteOrder) (Order, error) {
	switch byteOrder {
	case config.BigEndian:
		return &bigEndian{}, nil
	case config.LittleEndian:
		return &littleEndian{}, nil
	case config.BigEndianSwap:
		return &bigEndianSwap{}, nil
	case config.LittleEndianSwap:
		return &littleEndianSwap{}, nil
	case config.MidSwap:
		return &midSwap{}, nil
	default:
		return nil, fmt.Errorf("unsupported byte order: %s", byteOrder)
	}
}
