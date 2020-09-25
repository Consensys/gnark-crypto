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

package bls381

import (
	"github.com/consensys/gurvy/bls381/fp"
	"github.com/consensys/gurvy/utils"
)

// hashToFp hashes msg to count prime field elements.
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-5.2
func hashToFp(msg, dst []byte, count int) ([]fp.Element, error) {

	// 128 bits of security
	// L = ceil((ceil(log2(p)) + k) / 8), where k is the security parameter = 128
	L := 64

	lenInBytes := count * L
	pseudoRandomBytes, err := utils.ExpandMsgXmd(msg, dst, lenInBytes)
	if err != nil {
		return nil, err
	}

	res := make([]fp.Element, count)
	for i := 0; i < count; i++ {
		res[i].SetBytes(pseudoRandomBytes[i*L : (i+1)*L])
	}
	return res, nil
}

// hashToFp2 hashes msg to count fp2 elements.
// https://tools.ietf.org/html/draft-irtf-cfrg-hash-to-curve-06#section-5.2
func hashToFp2(msg, dst []byte, count int) ([]e2, error) {

	// 128 bits of security
	// L = ceil((ceil(log2(p)) + k) / 8), where k is the security parameter = 128
	L := 64

	lenInBytes := 2 * count * L
	pseudoRandomBytes, err := utils.ExpandMsgXmd(msg, dst, lenInBytes)
	if err != nil {
		return nil, err
	}

	res := make([]e2, count)
	for i := 0; i < count; i++ {
		res[i].A0.SetBytes(pseudoRandomBytes[i*2*L : (2*i+1)*L])
		res[i].A1.SetBytes(pseudoRandomBytes[(2*i+1)*L : 2*(i+1)*L])
	}
	return res, nil
}
