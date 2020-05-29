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
		FpName:    curve.FpName,
		FrName:    curve.FrName,
	}

	if err := primefields.Generate(d, "../../../../"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}
