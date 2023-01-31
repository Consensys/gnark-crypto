package main

import (
	"encoding/json"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"os"
)

type numericalMiMCTestCase struct {
	In  string `json:"in"`
	Out string `json:"out"`
}

func assertNoError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

//go:generate go run main.go
func main() {
	var tests []numericalMiMCTestCase

	bytes, err := os.ReadFile("./vectors.json")
	assertNoError(err)
	assertNoError(json.Unmarshal(bytes, &tests))

	for i := range tests {
		var x fr.Element
		_, err = x.SetString(tests[i].In)
		assertNoError(err)
		b := x.Bytes()
		bytes, err = mimc.Sum(b[:])
		assertNoError(err)
		x.SetBytes(bytes)
		tests[i].Out = "0x" + x.Text(16)
	}

	bytes, err = json.MarshalIndent(tests, "", "\t")
	assertNoError(err)
	err = os.WriteFile("./vectors.json", bytes, 0)
	assertNoError(err)
}
