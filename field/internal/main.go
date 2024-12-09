package main

import (
	"fmt"
	"path/filepath"

	"github.com/consensys/gnark-crypto/field/generator"
	"github.com/consensys/gnark-crypto/field/generator/config"
)

//go:generate go run main.go
func main() {
	// generate the following fields

	type field struct {
		name    string
		modulus string
	}

	fields := []field{
		{"goldilocks", "0xFFFFFFFF00000001"},
		{"koalabear", "0x7f000001"}, // 2^31 - 2^24 + 1 ==> the cube map (x -> x^3) is an automorphism of the multiplicative group
		{"babybear", "0x78000001"},  // 2^31 - 2^27 + 1 ==> 2-adicity 27
	}

	for _, f := range fields {

		// finite fields
		fc, err := config.NewFieldConfig(f.name, "Element", f.modulus, f.name, true)
		if err != nil {
			panic(err)
		}
		if err := generator.GenerateFF(fc, filepath.Join("..", f.name), "", ""); err != nil {
			panic(err)
		}
		fmt.Println("successfully generated", f.name, "field")

	}
}
