// Copyright 2020 ConsenSys AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"crypto/sha256"
	"errors"
)

// ExpandMsgXmd expands msg to a slice of lenInBytes bytes.
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-5
// https://tools.ietf.org/html/rfc8017#section-4.1 (I2OSP/O2ISP)
func ExpandMsgXmd(msg, dst []byte, lenInBytes int) ([]byte, error) {

	h := sha256.New()
	ell := (lenInBytes + h.Size() - 1) / h.Size() // ceil(len_in_bytes / b_in_bytes)
	if ell > 255 {
		return nil, errors.New("Invalid lenInBytes")
	}
	if len(dst) > 255 {
		return nil, errors.New("Invalid domain size (>255 bytes)")
	}
	sizeDomain := uint8(len(dst))

	// Z_pad = I2OSP(0, r_in_bytes)
	// l_i_b_str = I2OSP(len_in_bytes, 2)
	// DST_prime = I2OSP(len(DST), 1) || DST
	// b_0 = H(Z_pad || msg || l_i_b_str || I2OSP(0, 1) || DST_prime)
	h.Reset()
	h.Write(make([]byte, h.BlockSize()))
	h.Write(msg)
	h.Write([]byte{uint8(lenInBytes >> 8), uint8(lenInBytes), uint8(0)})
	h.Write(dst)
	h.Write([]byte{sizeDomain})
	b0 := h.Sum(nil)

	// b_1 = H(b_0 || I2OSP(1, 1) || DST_prime)
	h.Reset()
	h.Write(b0)
	h.Write([]byte{uint8(1)})
	h.Write(dst)
	h.Write([]byte{sizeDomain})
	b1 := h.Sum(nil)

	res := make([]byte, lenInBytes)
	copy(res[:h.Size()], b1)

	for i := 2; i <= ell; i++ {
		// b_i = H(strxor(b_0, b_(i - 1)) || I2OSP(i, 1) || DST_prime)
		h.Reset()
		strxor := make([]byte, h.Size())
		for j := 0; j < h.Size(); j++ {
			strxor[j] = b0[j] ^ b1[j]
		}
		h.Write(strxor)
		h.Write([]byte{uint8(i)})
		h.Write(dst)
		h.Write([]byte{sizeDomain})
		b1 = h.Sum(nil)
		copy(res[h.Size()*(i-1):h.Size()*i], b1)
	}
	return res, nil
}
