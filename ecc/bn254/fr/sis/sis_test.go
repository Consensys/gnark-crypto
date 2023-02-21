// Copyright 2020 ConsenSys Software Inc.
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

package sis

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
)

func pRand(seed *fr.Element) *fr.Element {
	var a fr.Element
	return a.Square(seed)
}

func polyRand(seed fr.Element, deg int) []fr.Element {
	res := make([]fr.Element, deg)
	for i := 0; i < deg; i++ {
		res[i].Set(pRand(&seed))
		seed.Set(&res[i])
	}
	return res
}

func TestReference(t *testing.T) {

	size := 16
	logTwoBound := 4
	degree := 4

	var shift fr.Element
	shift.SetString("19540430494807482326159819597004422086093766032135589407132600596362845576832")
	domain := fft.NewDomain(uint64(degree), shift)

	sis, err := NewRSis(5, 2, logTwoBound, size)
	if err != nil {
		t.Fatal(err)
	}
	ssis := sis.(*RSis)

	// generate the key deterministically
	var seed, one fr.Element
	one.SetOne()
	seed.SetUint64(5)
	for i := 0; i < size; i++ {
		ssis.A[i] = polyRand(seed, degree)
		copy(ssis.Ag[i], ssis.A[i])
		domain.FFT(ssis.Ag[i], fft.DIF, fft.WithCoset())
		seed.Add(&seed, &one)
	}

	// message to hash
	var m fr.Element
	m.SetString("21888242871839275222246405745257275088548364400416034343698204186575808495614")
	mb := m.Marshal()
	ssis.Write(mb)
	h := ssis.Sum(nil)
	sh := []byte{0x17, 0xcd, 0xe4, 0x27, 0xaa, 0x1, 0x3e, 0xd1, 0xc5, 0x4d, 0x1, 0xef, 0xa4, 0x6b, 0x6, 0xfc, 0xc4, 0xbe, 0x86, 0x91, 0xfc, 0xd7, 0x4a, 0xcf, 0x33, 0x8d, 0xc0, 0x80, 0xa1, 0x86, 0x7, 0x3b, 0xd, 0x50, 0x3d, 0x4, 0xa9, 0x88, 0xd5, 0xd3, 0x1c, 0x85, 0xe9, 0xea, 0x22, 0x6f, 0xc0, 0xac, 0x8c, 0xa4, 0xc4, 0x5f, 0x3b, 0x65, 0xac, 0xfc, 0xd8, 0x53, 0xf1, 0xf8, 0xf5, 0xe2, 0x6f, 0x9d, 0x23, 0xb9, 0x8b, 0x41, 0xb3, 0xab, 0xbd, 0x38, 0x28, 0xd8, 0xe6, 0x54, 0xee, 0x5f, 0x17, 0x43, 0xf9, 0x9b, 0x51, 0x2d, 0xfb, 0xeb, 0xc8, 0x60, 0x6c, 0x9a, 0x2d, 0xaa, 0x1c, 0xc0, 0x49, 0xa8, 0x12, 0xad, 0xc0, 0x9, 0x27, 0x9a, 0x90, 0xea, 0x95, 0x68, 0x57, 0x3f, 0x3a, 0x3d, 0xc1, 0x19, 0x63, 0xcb, 0xcc, 0x35, 0xd3, 0x18, 0xa5, 0x7c, 0x18, 0x71, 0xf7, 0xec, 0xd1, 0x2, 0xab, 0xa5}

	if len(h) != len(sh) {
		t.Fatal("unexpected length")
	}
	for i := 0; i < len(sh); i++ {
		if h[i] != sh[i] {
			t.Fatal("hash does not match expected result")
		}
	}

	// [ Sage comparison ]
	// m = Fr(21888242871839275222246405745257275088548364400416034343698204186575808495614)
	// mb = toBytes(m)
	// mb = toBytes(m, 32)
	// sis = Sis(5, 16, 4,4)
	// h = sis.sum(mc)
	// res =[]
	// for i in range(4):
	// 		res += toBytes(lift(h.coefficients()[i]), 32)

}

func TestSISParamsZKEVM(t *testing.T) {

	keySize := 65536
	logTwoBound := 3
	logTwoDegree := 1

	sis, _ := NewRSis(5, logTwoDegree, logTwoBound, keySize)

	// 96 = (1 << logTwoDegree) * logTwoBound * keySize / 256
	nbFrElements := ((1 << logTwoDegree) * logTwoBound * keySize) >> 8
	var p fr.Element
	for i := 0; i < nbFrElements; i++ {
		p.SetRandom()
		sis.Write(p.Marshal())
	}

	// sum
	sis.Sum(nil)

}

func BenchmarkSIS(b *testing.B) {

	//keySize := 1 << 20
	keySize := 1 << 20
	logTwoBound := 3
	logTwoDegree := 1

	sis, _ := NewRSis(5, logTwoDegree, logTwoBound, keySize)

	// 96 = (1 << logTwoDegree) * logTwoBound * keySize / 256
	nbFrElements := ((1 << logTwoDegree) * logTwoBound * keySize) >> 8
	var p fr.Element
	for i := 0; i < nbFrElements; i++ {
		p.SetRandom()
		sis.Write(p.Marshal())
	}

	// fmt.Printf("#Field elements %v\n", nbFrElements)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sis.Sum(nil)
	}

}
