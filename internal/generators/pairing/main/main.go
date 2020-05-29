package main

import (
	"fmt"
	"os"

	"github.com/consensys/gurvy/internal/generators/curve"
	"github.com/consensys/gurvy/internal/generators/pairing"
)

func main() {

	d := pairing.Data{
		Fpackage:        curve.C.Fpackage,
		Fp6NonResidue:   curve.C.Fp6NonResidue,
		EmbeddingDegree: curve.C.EmbeddingDegree,
		T:               curve.C.T,
		TNeg:            curve.C.TNeg,
		Fp2Name:         curve.C.Fp2Name,
		Fp6Name:         curve.C.Fp6Name,
		Fp12Name:        curve.C.Fp12Name,
	}

	if err := pairing.Generate(d, "../../../../"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}
