package main

import (
	"fmt"

	bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
)

func main() {

	_, _, g, _ := bls12381.Generators()
	var gg bls12381.G1Affine
	// m := fr.Modulus()
	// var r fr.Element
	// var rb, u big.Int
	// var rb big.Int
	//r.SetRandom()
	// r.ToBigIntRegular(&rb)
	// rb.SetString("", 10)
	// u.Mul(m, big.NewInt(123456789)).Add(&u, &rb)

	fmt.Printf("g: %s\n", g.String())
	// fmt.Printf("u: %s\n", rb.String())

	// gg.ScalarMultiplication(&g, &u)
	gg.ScalarMultiplication(&g, fr.Modulus())
	fmt.Printf("%s\n", gg.String())
	ggBytes := gg.RawBytes()
	fmt.Printf("size: %d\n", len(ggBytes))

	fmt.Printf("%x\n", ggBytes[:48])
	fmt.Printf("%x\n", ggBytes[48:])

}
