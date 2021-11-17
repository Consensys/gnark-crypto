package permutation

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/kzg"
)

func TestProof(t *testing.T) {

	srs, err := kzg.NewSRS(64, big.NewInt(13))
	if err != nil {
		t.Fatal(err)
	}

	a := make([]fr.Element, 8)
	b := make([]fr.Element, 8)
	for i := 0; i < 8; i++ {
		a[i].SetUint64(uint64(4*i + 1))
	}
	for i := 0; i < 8; i++ {
		b[i].Set(&a[(5*i)%8])
	}
	fmt.Print("t1 = [")
	for i := 0; i < 8; i++ {
		fmt.Printf("%s,", a[i].String())
	}
	fmt.Print("]\n")
	fmt.Print("t2 = [")
	for i := 0; i < 8; i++ {
		fmt.Printf("%s,", b[i].String())
	}
	fmt.Print("]\n")
	proof, err := Prove(srs, a, b)
	if err != nil {
		t.Fatal(err)
	}

	err = Verify(srs, proof)
	if err != nil {
		t.Fatal(err)
	}

}
