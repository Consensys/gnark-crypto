package sis

import (
	"fmt"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
)

func TestRSis(t *testing.T) {

	sis, err := NewRSis(5, 2, 3, 8)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < len(sis.A); i++ {
		printPoly(sis.A[i])
		fmt.Println("")
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
