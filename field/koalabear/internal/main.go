package main

import (
	"fmt"

	"github.com/consensys/gnark-crypto/field/generator"
	"github.com/consensys/gnark-crypto/field/generator/config"
)

//go:generate go run main.go
func main() {
	const modulus = "0x7f000001" // KoalaBear 2^31 - 2^24 + 1
	koalabear, err := config.NewFieldConfig("koalabear", "Element", modulus, true)
	if err != nil {
		panic(err)
	}
	if err := generator.GenerateFF(koalabear, "..", "", ""); err != nil {
		panic(err)
	}
	fmt.Println("successfully generated koalabear field")
}
