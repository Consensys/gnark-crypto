package main

import (
	"fmt"

	"github.com/consensys/gnark-crypto/field"
	"github.com/consensys/gnark-crypto/field/generator"
)

//go:generate go run main.go
func main() {
	const modulus = "0xFFFFFFFF00000001"
	goldilocks, err := field.NewField("goldilocks", "Element", modulus, false)
	if err != nil {
		panic(err)
	}
	if err := generator.GenerateFF(goldilocks, "../"); err != nil {
		panic(err)
	}
	fmt.Println("successfully generated goldilocks field")
}
