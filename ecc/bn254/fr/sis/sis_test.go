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
	r := sis.mulMod(p, q)

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
