package bitwise

import (
	"math/big"
)

func XOR256(a, b [32]byte) *big.Int {
	var out [32]byte
	for i := 0; i < 32; i++ {
		out[i] = a[i] ^ b[i]
	}
	return new(big.Int).SetBytes(out[:])
}

func OR256(a, b [32]byte) *big.Int {
	var out [32]byte
	for i := 0; i < 32; i++ {
		out[i] = a[i] | b[i]
	}
	return new(big.Int).SetBytes(out[:])
}

func AND256(a, b [32]byte) *big.Int {
	var out [32]byte
	for i := 0; i < 32; i++ {
		out[i] = a[i] & b[i]
	}
    return new(big.Int).SetBytes(out[:])
}

func NOT256(a [32]byte) *big.Int {
	var out [32]byte
	for i := 0; i < 32; i++ {
		out[i] = ^a[i]
	}
	return new(big.Int).SetBytes(out[:])
}

func XOR128(a, b [16]byte) *big.Int {
	var out [16]byte
	for i := 0; i < 16; i++ {
		out[i] = a[i] ^ b[i]
	}
	return new(big.Int).SetBytes(out[:])
}

func OR128(a, b [16]byte) *big.Int {
	var out [16]byte
	for i := 0; i < 16; i++ {
		out[i] = a[i] | b[i]
	}
	return new(big.Int).SetBytes(out[:])
}

func AND128(a, b [16]byte) *big.Int {
	var out [16]byte
	for i := 0; i < 16; i++ {
		out[i] = a[i] & b[i]
	}
	return new(big.Int).SetBytes(out[:])
}

func NOT128(a [16]byte) *big.Int {
	// Interpret a as an unsigned 128-bit integer (big-endian)
	x := new(big.Int).SetBytes(a[:])
	// mask = 2^128 - 1 (all 128 bits set to 1)
	mask := new(big.Int).Lsh(big.NewInt(1), 128)
	mask.Sub(mask, big.NewInt(1))
	// x = ^a == a XOR mask  (bitwise NOT over 128 bits)
	x.Xor(x, mask)
	return x
}

func XOR64(a, b [8]byte) *big.Int {
	// Interpret inputs as big-endian uint64
	va := uint64(a[0])<<56 |
		uint64(a[1])<<48 |
		uint64(a[2])<<40 |
		uint64(a[3])<<32 |
		uint64(a[4])<<24 |
		uint64(a[5])<<16 |
		uint64(a[6])<<8 |
		uint64(a[7])
	vb := uint64(b[0])<<56 |
		uint64(b[1])<<48 |
		uint64(b[2])<<40 |
		uint64(b[3])<<32 |
		uint64(b[4])<<24 |
		uint64(b[5])<<16 |
		uint64(b[6])<<8 |
		uint64(b[7])
	// Convert to *big.Int
	return new(big.Int).SetUint64(va ^ vb)
}

func OR64(a, b [8]byte) *big.Int {
	// Interpret inputs as big-endian uint64
	va := uint64(a[0])<<56 |
		uint64(a[1])<<48 |
		uint64(a[2])<<40 |
		uint64(a[3])<<32 |
		uint64(a[4])<<24 |
		uint64(a[5])<<16 |
		uint64(a[6])<<8 |
		uint64(a[7])
	vb := uint64(b[0])<<56 |
		uint64(b[1])<<48 |
		uint64(b[2])<<40 |
		uint64(b[3])<<32 |
		uint64(b[4])<<24 |
		uint64(b[5])<<16 |
		uint64(b[6])<<8 |
		uint64(b[7])
	// Convert to *big.Int
	return new(big.Int).SetUint64(va | vb)
}

func AND64(a, b [8]byte) *big.Int {
	// Interpret inputs as big-endian uint64
	va := uint64(a[0])<<56 |
		uint64(a[1])<<48 |
		uint64(a[2])<<40 |
		uint64(a[3])<<32 |
		uint64(a[4])<<24 |
		uint64(a[5])<<16 |
		uint64(a[6])<<8 |
		uint64(a[7])
	vb := uint64(b[0])<<56 |
		uint64(b[1])<<48 |
		uint64(b[2])<<40 |
		uint64(b[3])<<32 |
		uint64(b[4])<<24 |
		uint64(b[5])<<16 |
		uint64(b[6])<<8 |
		uint64(b[7])
	// Convert to *big.Int
	return new(big.Int).SetUint64(va & vb)
}

func NOT64(a [8]byte) *big.Int {
	// Interpret a as big-endian uint64
	v := uint64(a[0])<<56 |
		uint64(a[1])<<48 |
		uint64(a[2])<<40 |
		uint64(a[3])<<32 |
		uint64(a[4])<<24 |
		uint64(a[5])<<16 |
		uint64(a[6])<<8 |
		uint64(a[7])
	// Convert to *big.Int
	return big.NewInt(0).SetUint64(^v)
}

func XOR32(a, b [4]byte) *big.Int {
	// interpret inputs as big-endian uint32
	va := uint32(a[0])<<24 | uint32(a[1])<<16 | uint32(a[2])<<8 | uint32(a[3])
	vb := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	// convert to *big.Int
	return big.NewInt(int64(va ^ vb))
}

func OR32(a, b [4]byte) *big.Int {
	// interpret inputs as big-endian uint32
	va := uint32(a[0])<<24 | uint32(a[1])<<16 | uint32(a[2])<<8 | uint32(a[3])
	vb := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	// convert to *big.Int
	return big.NewInt(int64(va | vb))
}

func AND32(a, b [4]byte) *big.Int {
	// interpret inputs as big-endian uint32
	va := uint32(a[0])<<24 | uint32(a[1])<<16 | uint32(a[2])<<8 | uint32(a[3])
	vb := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	// convert to *big.Int
	return big.NewInt(int64(va & vb))
}

func NOT32(a [4]byte) *big.Int {
	// Interpret a as big-endian uint32
	v := uint32(a[0])<<24 |
		uint32(a[1])<<16 |
		uint32(a[2])<<8 |
		uint32(a[3])
	// Convert to *big.Int
	return big.NewInt(int64(^v & 0xFFFFFFFF))
}

func XOR16(a, b [2]byte) *big.Int {
	// interpret inputs as big-endian uint16
	va := uint16(a[0])<<8 | uint16(a[1])
	vb := uint16(b[0])<<8 | uint16(b[1])
	// convert to *big.Int
	return big.NewInt(int64(va ^ vb))
}

func OR16(a, b [2]byte) *big.Int {
	// interpret inputs as big-endian uint16
	va := uint16(a[0])<<8 | uint16(a[1])
	vb := uint16(b[0])<<8 | uint16(b[1])
	// convert to *big.Int
	return big.NewInt(int64(va | vb))
}

func AND16(a, b [2]byte) *big.Int {
	// interpret inputs as big-endian uint16
	va := uint16(a[0])<<8 | uint16(a[1])
	vb := uint16(b[0])<<8 | uint16(b[1])
	// convert to *big.Int
	return big.NewInt(int64(va & vb))
}

func NOT16(a [2]byte) *big.Int {
	// Interpret a as a big-endian uint16
	v := uint16(a[0])<<8 | uint16(a[1])
	// Bitwise NOT on 16 bits
	v = ^v & 0xFFFF
	// Convert to *big.Int
	return big.NewInt(int64(v))
}

func XOR8(a, b [1]byte) *big.Int {
	return big.NewInt(int64(a[0] ^ b[0]))
}

func OR8(a, b [1]byte) *big.Int {
	return big.NewInt(int64(a[0] | b[0]))
}

func AND8(a, b [1]byte) *big.Int {
	return big.NewInt(int64(a[0] & b[0]))
}

func NOT8(a [1]byte) *big.Int {
	return big.NewInt(int64(^a[0]))
}

func Xor4Bits(arg1, arg2 uint8) *big.Int {
	// mask to keep only 4 bits
	const mask uint8 = 0x0F

	res := (arg1 & mask) ^ (arg2 & mask)
	return new(big.Int).SetUint64(uint64(res))
}

func Or4Bits(arg1, arg2 uint8) *big.Int {
	const mask uint8 = 0x0F
	res := (arg1 & mask) | (arg2 & mask)
	return new(big.Int).SetUint64(uint64(res))
}

func And4Bits(arg1, arg2 uint8) *big.Int {
	const mask uint8 = 0x0F
	res := (arg1 & mask) & (arg2 & mask)
	return new(big.Int).SetUint64(uint64(res))
}

func Not4Bits(arg1 uint8) *big.Int {
	const mask uint8 = 0x0F // 0b11, keep only 2 bits
	res := (^arg1) & mask   // invert, then mask to 2 bits
	return new(big.Int).SetUint64(uint64(res))
}

func Xor2Bits(arg1, arg2 uint8) *big.Int {
	const mask uint8 = 0x03 // 0b11
	res := (arg1 & mask) ^ (arg2 & mask)
	return new(big.Int).SetUint64(uint64(res))
}

func Or2Bits(arg1, arg2 uint8) *big.Int {
	const mask uint8 = 0x03 // 0b11
	res := (arg1 & mask) | (arg2 & mask)
	return new(big.Int).SetUint64(uint64(res))
}

func And2Bits(arg1, arg2 uint8) *big.Int {
	const mask uint8 = 0x03 // 0b11
	res := (arg1 & mask) & (arg2 & mask)
	return new(big.Int).SetUint64(uint64(res))
}

func Not2Bits(arg1 uint8) *big.Int {
	const mask uint8 = 0x03 // 0b11, keep only 2 bits
	res := (^arg1) & mask   // invert, then mask to 2 bits
	return new(big.Int).SetUint64(uint64(res))
}

func BigIntTo32Bytes(n *big.Int) [32]byte {
	var out [32]byte

	if n == nil {
		return out
	}

	b := n.Bytes() // big-endian, no leading zeros
	// If longer than 32, keep only the last 32 bytes (lowest 256 bits)
	if len(b) > 32 {
		b = b[len(b)-32:]
	}

	// Copy into the end of out to left-pad with zeros
	copy(out[32-len(b):], b)
	return out
}

func BigIntTo16Bytes(n *big.Int) [16]byte {
	var out [16]byte

	if n == nil {
		return out
	}

	b := n.Bytes() // big-endian, no leading zeros
	if len(b) > 16 {
		b = b[len(b)-16:]
	}

	copy(out[16-len(b):], b)
	return out
}

func BigIntTo8Bytes(n *big.Int) [8]byte {
	var out [8]byte

	if n == nil {
		return out
	}

	b := n.Bytes() // big-endian, no leading zeros
	if len(b) > 8 {
		b = b[len(b)-8:]
	}

	copy(out[8-len(b):], b)
	return out
}

func BigIntTo4Bytes(n *big.Int) [4]byte {
	var out [4]byte

	if n == nil {
		return out
	}

	b := n.Bytes() // big-endian, no leading zeros
	if len(b) > 4 {
		b = b[len(b)-4:]
	}

	copy(out[4-len(b):], b)
	return out
}

func BigIntTo2Bytes(n *big.Int) [2]byte {
	var out [2]byte

	if n == nil {
		return out
	}

	b := n.Bytes() // big-endian, no leading zeros
	if len(b) > 2 {
		b = b[len(b)-2:]
	}

	copy(out[2-len(b):], b)
	return out
}

func BigIntTo1Bytes(n *big.Int) [1]byte {
	var out [1]byte

	if n == nil {
		return out
	}

	b := n.Bytes() // big-endian, no leading zeros
	if len(b) > 1 {
		b = b[len(b)-1:]
	}

	copy(out[1-len(b):], b)
	return out
}

func BigIntTo4Bits(n *big.Int) uint8 {
	if n == nil {
		return 0
	}
	var mask = big.NewInt(0xF) // 0b1111
	var tmp big.Int
	tmp.And(n, mask)
	return uint8(tmp.Uint64()) // value is guaranteed to be <= 15
}

func BigIntTo2Bits(n *big.Int) uint8 {
	if n == nil {
		return 0
	}
	var mask = big.NewInt(0x3) // 0b11
	var tmp big.Int
	tmp.And(n, mask)
	return uint8(tmp.Uint64()) // guaranteed <= 3
}

func SplitByteInto2BigInt(n [1]byte) (high, low *big.Int) {
	b := n[0]
	h := (b & 0xF0) >> 4 // upper 4 bits
	l := b & 0x0F        // lower 4 bits

	high = big.NewInt(int64(h))
	low = big.NewInt(int64(l))
	return
}

func SplitUint8Into2BigInt(n uint8) (high, low *big.Int) {
	l_hi := (n >> 2) & 0x3
	l_lo := n & 0x03
	high = big.NewInt(int64(l_hi))
	low = big.NewInt(int64(l_lo))
	return
}

// not2Bit returns the bitwise NOT of a 2‑bit value (only lowest 2 bits kept).
func Not2Bit(arg1 uint8) *big.Int {
	out := ^arg1 & 0b11 // or & 3
	return new(big.Int).SetUint64(uint64(out))
}
