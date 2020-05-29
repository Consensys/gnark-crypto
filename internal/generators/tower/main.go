package main

import (
	"fmt"
	"os"

	tower "github.com/consensys/gurvy/internal/generators/tower/template"
)

// go:generate go run main.go
func main() {

	// TODO curve data copied in each template generator. Read from a config file instead
	// -------------------------------------------------------------------------------------------------
	// bls381
	bls381 := tower.GenerateData{
		Fpackage:      "bls381",
		FpModulus:     "4002409555221667393417789825735904156556882819939007885332058136124031650490837864442687629129015664037894272559787",
		FrModulus:     "52435875175126190479447740508185965837690552500527637822603658699938581184513",
		Fp2NonResidue: "-1",
		Fp6NonResidue: "1,1",
		MakeFp12:      true,
		Fp2Name:       tower.Fp2Name,
		Fp6Name:       tower.Fp6Name,
		Fp12Name:      tower.Fp12Name,
	}

	// -------------------------------------------------------------------------------------------------
	// bls377
	bls377 := tower.GenerateData{
		Fpackage:      "bls377",
		FpModulus:     "258664426012969094010652733694893533536393512754914660539884262666720468348340822774968888139573360124440321458177",
		FrModulus:     "8444461749428370424248824938781546531375899335154063827935233455917409239041",
		Fp2NonResidue: "5",
		Fp6NonResidue: "0,1",
		MakeFp12:      true,
		Fp2Name:       tower.Fp2Name,
		Fp6Name:       tower.Fp6Name,
		Fp12Name:      tower.Fp12Name,
	}

	// -------------------------------------------------------------------------------------------------
	// bn256
	bn256 := tower.GenerateData{
		Fpackage:      "bn256",
		FpModulus:     "21888242871839275222246405745257275088696311157297823662689037894645226208583",
		FrModulus:     "21888242871839275222246405745257275088548364400416034343698204186575808495617",
		Fp2NonResidue: "-1",
		Fp6NonResidue: "9,1",
		MakeFp12:      true,
		Fp2Name:       tower.Fp2Name,
		Fp6Name:       tower.Fp6Name,
		Fp12Name:      tower.Fp12Name,
	}

	// -------------------------------------------------------------------------------------------------
	// BW6-781
	bw6_761 := tower.GenerateData{
		Fpackage:      "bw6_761",
		FpModulus:     "6891450384315732539396789682275657542479668912536150109513790160209623422243491736087683183289411687640864567753786613451161759120554247759349511699125301598951605099378508850372543631423596795951899700429969112842764913119068299",
		FrModulus:     bls377.FpModulus,
		Fp2NonResidue: "-1",
		Fp6NonResidue: "1,1",
		MakeFp12:      false,
		Fp2Name:       tower.Fp2Name,
		Fp6Name:       tower.Fp6Name,
		Fp12Name:      tower.Fp12Name,
	}

	curve := [...]tower.GenerateData{
		bls381,
		bls377,
		bn256,
		bw6_761,
	}

	for _, d := range curve {
		if err := tower.GenerateTower(d); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
	}
}
