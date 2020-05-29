package main

import (
	"fmt"
	"os"

	"github.com/consensys/gurvy/internal/generators/curve"
	"github.com/consensys/gurvy/internal/generators/primefields"
)

func main() {

	d := primefields.Data{
		Fpackage:  curve.C.Fpackage,
		FpModulus: curve.C.FpModulus,
		FrModulus: curve.C.FrModulus,
		FpName:    curve.C.FpName,
		FrName:    curve.C.FrName,
	}

	// assume working directory is internal/generators
	// TODO make this path more robust to changes in working directory
	if err := primefields.Generate(d, "../../"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}
