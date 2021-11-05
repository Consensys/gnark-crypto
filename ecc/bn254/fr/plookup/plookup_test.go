package plookup

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/kzg"
)

func TestH(t *testing.T) {

	lookup := make(Table, 8)
	f := make(Table, 7)
	for i := 0; i < 8; i++ {
		lookup[i].SetUint64(uint64(2 * i))
	}
	for i := 0; i < 7; i++ {
		f[i].Set(&lookup[(4*i+1)%8])
	}

	srs, err := kzg.NewSRS(64, big.NewInt(13))
	if err != nil {
		t.Fatal(err)
	}

	_, err = Prove(srs, f, lookup)
	if err != nil {
		t.Fatal(err)
	}

}

// func TestSortByT(t *testing.T) {

// 	a := make(Table, 8)
// 	b := make(Table, 7)
// 	for i := 0; i < 8; i++ {
// 		a[i].SetRandom()
// 	}
// 	for i := 0; i < 7; i++ {
// 		b[i].Set(&a[(4*i)%8])
// 	}
// 	sort.Sort(a)
// 	sort.Sort(b)

// 	for i := 0; i < 8; i++ {
// 		fmt.Printf("%s\n", a[i].String())
// 	}
// 	fmt.Printf("--\n")
// 	for i := 0; i < 7; i++ {
// 		fmt.Printf("%s\n", b[i].String())
// 	}

// 	v := sortByt(b, a)
// 	fmt.Printf("--\n")
// 	fmt.Printf("len(v): %d\n", len(v))
// 	for i := 0; i < len(v); i++ {
// 		fmt.Printf("%s\n", v[i].String())
// 	}

// }
