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

func TestRSis(t *testing.T) {

	keySize := 8

	sis, err := NewRSis(5, 1, 4, keySize)
	if err != nil {
		t.Fatal(err)
	}

	m := make([]byte, 8)
	m[0] = 0xa1
	m[1] = 0x90
	m[2] = 0xff
	m[3] = 0x0a
	m[4] = 0x13
	m[5] = 0x59
	m[6] = 0x79
	m[7] = 0xcc

	sis.Write(m)

	res := sis.Sum(nil)

	resPol := make([]fr.Element, sis.Degree)
	for i := 0; i < sis.Degree; i++ {
		resPol[i].SetBytes(res[i*32 : (i+1)*32])
	}

	expectedRes := make([]fr.Element, sis.Degree)
	expectedRes[0].SetString("13271020168286836418355708644485735593608516629558571827355518635690915176270")
	expectedRes[1].SetString("9885652947755511462638910175213772082420069489359143817296501612386750845004")

	for i := 0; i < sis.Degree; i++ {
		if !expectedRes[i].Equal(&resPol[i]) {
			t.Fatal("error sis hash")
		}
	}

}

func TestMulMod(t *testing.T) {

	sis, err := NewRSis(5, 2, 3, 8)
	if err != nil {
		t.Fatal(err)
	}

	p := make([]fr.Element, 4)
	p[0].SetString("2389")
	p[1].SetString("987192")
	p[2].SetString("623")
	p[3].SetString("91")

	q := make([]fr.Element, 4)
	q[0].SetString("76755")
	q[1].SetString("232893720")
	q[2].SetString("989273")
	q[3].SetString("675273")

	sis.Domain.FFT(p, fft.DIF, true)
	sis.Domain.FFT(q, fft.DIF, true)
	r := sis.MulMod(p, q)

	expectedr := make([]fr.Element, 4)
	expectedr[0].SetString("21888242871839275222246405745257275088548364400416034343698204185887558114297")
	expectedr[1].SetString("631644300118")
	expectedr[2].SetString("229913166975959")
	expectedr[3].SetString("1123315390878")

	for i := 0; i < 4; i++ {
		if !expectedr[i].Equal(&r[i]) {
			t.Fatal("product failed")
		}
	}

}

func BenchmarkSIS(b *testing.B) {

	keySize := 64
	logTwoBound := 6
	logTwoDegree := 6

	sis, _ := NewRSis(5, logTwoDegree, logTwoBound, keySize)

	// 96 = (1 << logTwoDegree) * logTwoBound * keySize / 256
	var p fr.Element
	for i := 0; i < 96; i++ {
		p.SetRandom()
		sis.Write(p.Marshal())
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sis.Sum(nil)
		sis.Reset()
	}

}
