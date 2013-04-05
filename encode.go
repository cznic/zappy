// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Copyright 2011 The Snappy-Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the SNAPPY-GO-LICENSE file.

package zappy

import (
	"encoding/binary"
)

// We limit how far copy back-references can go, the same as the snappy C++
// code.
const maxOffset = 1 << 20

// emitLiteral writes a literal chunk and returns the number of bytes written.
func emitLiteral(dst, lit []byte) (n int) {
	n = binary.PutVarint(dst, int64(len(lit)-1))
	n += copy(dst[n:], lit)
	return
}

// emitCopy writes a copy chunk and returns the number of bytes written.
func emitCopy(dst []byte, offset, length int) (n int) {
	n = binary.PutVarint(dst, int64(-length))
	n += binary.PutUvarint(dst[n:], uint64(offset))
	return
}

// Encode returns the encoded form of src. The returned slice may be a sub-
// slice of dst if dst was large enough to hold the entire encoded block.
// Otherwise, a newly allocated slice will be returned.
// It is valid to pass a nil dst.
func Encode(dst, src []byte) ([]byte, error) {
	if n := MaxEncodedLen(len(src)); len(dst) < n {
		dst = make([]byte, n)
	}

	// The block starts with the varint-encoded length of the decompressed bytes.
	d := binary.PutUvarint(dst, uint64(len(src)))

	// Return early if src is short.
	if len(src) <= 4 {
		if len(src) != 0 {
			d += emitLiteral(dst[d:], src)
		}
		return dst[:d], nil
	}

	// Iterate over the source bytes.
	var (
		table [1 << 12]int // Hash table
		s     int          // The iterator position.
		t     int          // The last position with the same hash as s.
		lit   int          // The start position of any pending literal bytes.
	)
	for s+3 < len(src) {
		// Update the hash table.
		b0, b1, b2, b3 := src[s], src[s+1], src[s+2], src[s+3]
		h := uint32(b0) | uint32(b1)<<8 | uint32(b2)<<16 | uint32(b3)<<24
		p := &table[(h*0x1e35a7bd)>>20]
		t, *p = *p, s
		// If t is invalid or src[s:s+4] differs from src[t:t+4], accumulate a literal byte.
		if t == 0 || s-t >= maxOffset || b0 != src[t] || b1 != src[t+1] || b2 != src[t+2] || b3 != src[t+3] {
			s++
			continue
		}

		// Otherwise, we have a match. First, emit any pending literal bytes.
		if lit != s {
			d += emitLiteral(dst[d:], src[lit:s])
		}
		// Extend the match to be as long as possible.
		s0 := s
		s, t = s+4, t+4
		for s < len(src) && src[s] == src[t] {
			s++
			t++
		}
		// Emit the copied bytes.
		d += emitCopy(dst[d:], s-t, s-s0)
		lit = s
	}

	// Emit any final pending literal bytes and return.
	if lit != len(src) {
		d += emitLiteral(dst[d:], src[lit:])
	}
	return dst[:d], nil
}

// MaxEncodedLen returns the maximum length of a zappy block, given its
// uncompressed length.
func MaxEncodedLen(srcLen int) int {
	return 10 + srcLen
}
