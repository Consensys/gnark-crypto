package main

import (
	"os"

	"github.com/consensys/gurvy/internal/generators/template/generator"
)

//go:generate go run main.go
func main() {

	bls377 := generator.GenerateData{
		Fpackage:      "bls377",
		RootPath:      "../../bls377/",
		FpModulus:     "258664426012969094010652733694893533536393512754914660539884262666720468348340822774968888139573360124440321458177",
		FpName:        "fp",
		FrModulus:     "8444461749428370424248824938781546531375899335154063827935233455917409239041",
		FrName:        "fr",
		Fp2Name:       "e2",
		Fp2NonResidue: "5",
		Fp6Name:       "e6",
		Fp6NonResidue: "0,1",
		Fp12Name:      "e12",
		T:             "9586122913090633729",
		TNeg:          false,
		PointName:     "G",
	}

	// create folder for the cruve
	if err := os.MkdirAll(bls377.RootPath, 0700); err != nil {
		panic(err)
	}
	if err := generator.GenerateCurve(bls377); err != nil {
		panic(err)
	}

	//generator.GenerateCurve(bls377)

}
