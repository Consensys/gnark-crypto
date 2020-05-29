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
	// bls381
	bls381 := pairing.GenerateData{
		Fpackage:        "bls381",
		FpModulus:       "4002409555221667393417789825735904156556882819939007885332058136124031650490837864442687629129015664037894272559787",
		FrModulus:       "52435875175126190479447740508185965837690552500527637822603658699938581184513",
		Fp2NonResidue:   "-1",
		Fp6NonResidue:   "1,1",
		MakeFp12:        true,
		T:               "15132376222941642752",
		TNeg:            true,
		EmbeddingDegree: 12,
		Fp2Name:         pairing.Fp2Name,
		Fp6Name:         pairing.Fp6Name,
		Fp12Name:        pairing.Fp12Name,
	}

	// -------------------------------------------------------------------------------------------------
	// bls377
	bls377 := pairing.GenerateData{
		Fpackage:        "bls377",
		FpModulus:       "258664426012969094010652733694893533536393512754914660539884262666720468348340822774968888139573360124440321458177",
		FrModulus:       "8444461749428370424248824938781546531375899335154063827935233455917409239041",
		Fp2NonResidue:   "5",
		Fp6NonResidue:   "0,1",
		MakeFp12:        true,
		T:               "9586122913090633729",
		TNeg:            false,
		EmbeddingDegree: 12,
		Fp2Name:         pairing.Fp2Name,
		Fp6Name:         pairing.Fp6Name,
		Fp12Name:        pairing.Fp12Name,
	}

	// -------------------------------------------------------------------------------------------------
	// BW6-781
	bw6_761 := pairing.GenerateData{
		Fpackage:        "bw6_761",
		FpModulus:       "6891450384315732539396789682275657542479668912536150109513790160209623422243491736087683183289411687640864567753786613451161759120554247759349511699125301598951605099378508850372543631423596795951899700429969112842764913119068299",
		FrModulus:       bls377.FpModulus,
		Fp2NonResidue:   "-1",
		Fp6NonResidue:   "1,1",
		MakeFp12:        false,
		T:               "4371281006305286848163485150587564217350348383473813272171392617577304305730246023460948554022786", // TODO this is the hard part exponent.  Implement the optimized hard part from Appendix B of https://eprint.iacr.org/2020/351.pdf
		TNeg:            false,
		EmbeddingDegree: 6,
		Fp2Name:         pairing.Fp2Name,
		Fp6Name:         pairing.Fp6Name,
		Fp12Name:        pairing.Fp12Name,
	}

	curve := [...]pairing.GenerateData{
		bls381,
		// bls377,
		// bn256,
		// bw6_761,
	}

	_ = bls381
	_ = bls377
	_ = bw6_761

	for _, d := range curve {
		if err := pairing.GeneratePairing(d); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
	}
}
