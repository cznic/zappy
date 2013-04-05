// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Copyright 2011 The Snappy-Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the SNAPPY-GO-LICENSE file.

package zappy

import (
	"encoding/binary"
	"errors"
)

// ErrCorrupt reports that the input is invalid.
var ErrCorrupt = errors.New("zappy: corrupt input")

// DecodedLen returns the length of the decoded block.
func DecodedLen(src []byte) (int, error) {
	v, _, err := decodedLen(src)
	return v, err
}

// decodedLen returns the length of the decoded block and the number of bytes
// that the length header occupied.
func decodedLen(src []byte) (blockLen, headerLen int, err error) {
	v, n := binary.Uvarint(src)
	if n == 0 {
		return 0, 0, ErrCorrupt
	}

	if uint64(int(v)) != v {
		return 0, 0, errors.New("zappy: decoded block is too large")
	}

	return int(v), n, nil
}

// Decode returns the decoded form of src. The returned slice may be a sub-
// slice of dst if dst was large enough to hold the entire decoded block.
// Otherwise, a newly allocated slice will be returned.
// It is valid to pass a nil dst.
func Decode(dst, src []byte) ([]byte, error) {
	dLen, s, err := decodedLen(src)
	if err != nil {
		return nil, err
	}

	if len(dst) < dLen {
		dst = make([]byte, dLen)
	}

	var d, offset, length int
	for s < len(src) {
		n, i := binary.Varint(src[s:])
		if i <= 0 {
			return nil, ErrCorrupt
		}

		s += i
		if n >= 0 {
			length = int(n + 1)
			if length > len(dst)-d || length > len(src)-s {
				return nil, ErrCorrupt
			}

			copy(dst[d:], src[s:s+length])
			d += length
			s += length
			continue
		}

		length = int(-n)
		off64, i := binary.Uvarint(src[s:])
		if i <= 0 {
			return nil, ErrCorrupt
		}

		offset = int(off64)
		s += i
		if s > len(src) {
			return nil, ErrCorrupt
		}

		end := d + length
		if offset > d || end > len(dst) {
			return nil, ErrCorrupt
		}

		for ; d < end; d++ {
			dst[d] = dst[d-offset]
		}
	}
	if d != dLen {
		return nil, ErrCorrupt
	}

	return dst[:d], nil
}
