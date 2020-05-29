package main

import (
	"fmt"
	"os"

	"github.com/consensys/gurvy/internal/generators/curve"
	"github.com/consensys/gurvy/internal/generators/tower"
)

func main() {

	d := tower.Data{
		Fpackage:        curve.C.Fpackage,
		Fp2NonResidue:   curve.C.Fp2NonResidue,
		Fp6NonResidue:   curve.C.Fp6NonResidue,
		EmbeddingDegree: curve.C.EmbeddingDegree,
		Fp2Name:         curve.C.Fp2Name,
		Fp6Name:         curve.C.Fp6Name,
		Fp12Name:        curve.C.Fp12Name,
	}

	if err := tower.Generate(d, "../../../../"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}
