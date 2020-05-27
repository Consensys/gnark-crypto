package main

import (
	"fmt"
	"os"

	pairing "github.com/consensys/gurvy/internal/generators/pairing/template"
)

//go:generate go run main.go
func main() {

	// TODO curve data copied in each template generator. Read from a config file instead
	// -------------------------------------------------------------------------------------------------
	// bls377
	bls377 := pairing.GenerateData{
		Fpackage:      "bls377",
		FpModulus:     "258664426012969094010652733694893533536393512754914660539884262666720468348340822774968888139573360124440321458177",
		FrModulus:     "8444461749428370424248824938781546531375899335154063827935233455917409239041",
		Fp2NonResidue: "5",
		Fp6NonResidue: "0,1",
		MakeFp12:      true,
		Fp2Name:       pairing.Fp2Name,
		Fp6Name:       pairing.Fp6Name,
		Fp12Name:      pairing.Fp12Name,
	}

	// -------------------------------------------------------------------------------------------------
	// BW6-781
	bw6_761 := pairing.GenerateData{
		Fpackage:      "bw6_761",
		FpModulus:     "6891450384315732539396789682275657542479668912536150109513790160209623422243491736087683183289411687640864567753786613451161759120554247759349511699125301598951605099378508850372543631423596795951899700429969112842764913119068299",
		FrModulus:     bls377.FpModulus,
		Fp2NonResidue: "-1",
		Fp6NonResidue: "1,1",
		MakeFp12:      false,
		Fp2Name:       pairing.Fp2Name,
		Fp6Name:       pairing.Fp6Name,
		Fp12Name:      pairing.Fp12Name,
	}

	curve := [...]pairing.GenerateData{
		// bls381,
		// bls377,
		// bn256,
		bw6_761,
	}

	for _, d := range curve {
		if err := pairing.GeneratePairing(d); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
	}
}
