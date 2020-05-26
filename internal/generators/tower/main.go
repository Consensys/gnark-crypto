package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/consensys/bavard"
	"github.com/consensys/goff/cmd"
	"github.com/consensys/gurvy/internal/generators/tower/fp12"
	"github.com/consensys/gurvy/internal/generators/tower/fp2"
	"github.com/consensys/gurvy/internal/generators/tower/fp6"
)

const fpName = "fp"
const frName = "fr"
const fp2Name = "E2"
const fp6Name = "E6"
const fp12Name = "E12"

// GenerateData data used to generate the templates
type GenerateData struct {

	// common
	Fpackage string
	// RootPath string // TODO deduce this from Fpackage; remove it

	// fp, fr moduli
	// FpName    string // TODO this name cannot change; remove it
	FpModulus string
	FrModulus string
	// FrName    string // TODO this name cannot change; remove it

	// fp2
	Fp2NonResidue string

	// fp6
	Fp6NonResidue string

	MakeFp12 bool // TODO need a better way to specify which fields to make
	// fp12
	// Fp12Name string // TODO this name cannot change; remove it

	// data needed in the template, always set to constants
	Fp2Name  string // TODO this name cannot change; remove it
	Fp6Name  string // TODO this name cannot change; remove it
	Fp12Name string // TODO this name cannot change; remove it
}

//go:generate go run main.go
func main() {

	// -------------------------------------------------------------------------------------------------
	// bls381
	// bls381 := GenerateData{
	// 	Fpackage: "bls381",
	// 	// RootPath:      "../../bls381/",
	// 	FpModulus: "4002409555221667393417789825735904156556882819939007885332058136124031650490837864442687629129015664037894272559787",
	// 	// FpName:        "fp",
	// 	FrModulus: "52435875175126190479447740508185965837690552500527637822603658699938581184513",
	// 	// FrName:        "fr",
	// 	// Fp2Name:       "E2",
	// 	Fp2NonResidue: "-1",
	// 	// Fp6Name:       "E6",
	// 	Fp6NonResidue: "1,1",
	// 	// Fp12Name:      "E12",
	// }

	// -------------------------------------------------------------------------------------------------
	// bls377
	bls377 := GenerateData{
		Fpackage:      "bls377",
		FpModulus:     "258664426012969094010652733694893533536393512754914660539884262666720468348340822774968888139573360124440321458177",
		FrModulus:     "8444461749428370424248824938781546531375899335154063827935233455917409239041",
		Fp2NonResidue: "5",
		Fp6NonResidue: "0,1",
		MakeFp12:      true,
		Fp2Name:       fp2Name,
		Fp6Name:       fp6Name,
		Fp12Name:      fp12Name,
	}

	// -------------------------------------------------------------------------------------------------
	// BW6-781
	bw6_761 := GenerateData{
		Fpackage:      "bw6_761",
		FpModulus:     "6891450384315732539396789682275657542479668912536150109513790160209623422243491736087683183289411687640864567753786613451161759120554247759349511699125301598951605099378508850372543631423596795951899700429969112842764913119068299",
		FrModulus:     bls377.FpModulus,
		Fp2NonResidue: "-1",
		Fp6NonResidue: "1,1",
		MakeFp12:      false,
		Fp2Name:       fp2Name,
		Fp6Name:       fp6Name,
		Fp12Name:      fp12Name,
	}

	curve := [...]GenerateData{
		// bls381,
		bls377,
		// bn256,
		bw6_761,
	}

	for _, d := range curve {
		if err := GenerateTower(d); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
	}
}

// GenerateTower generates tower, curve, pairing
func GenerateTower(d GenerateData) error {

	rootPath := "../../../" + d.Fpackage + "/"

	// fp, fr
	{
		fpPath := filepath.Join(rootPath, "fp")
		// if err := os.MkdirAll(fpPath, 0700); err != nil {
		// 	return err
		// }
		if err := cmd.GenerateFF(fpName, "Element", d.FpModulus, fpPath, false, false); err != nil {
			return err
		}
		frPath := filepath.Join(rootPath, "fr")
		// if err := os.MkdirAll(frPath, 0700); err != nil {
		// 	return err
		// }
		if err := cmd.GenerateFF(frName, "Element", d.FrModulus, frPath, false, false); err != nil {
			return err
		}
	}

	// fp2
	{
		src := []string{
			fp2.Base,
			fp2.Inline,
			fp2.Mul,
		}
		if err := bavard.Generate(rootPath+strings.ToLower(fp2Name)+".go", src, d,
			bavard.Package(d.Fpackage),
			bavard.Apache2("ConsenSys AG", 2020),
			bavard.GeneratedBy("gurvy/internal/generators"),
		); err != nil {
			return err
		}
	}

	// fp6
	{
		src := []string{
			fp6.Base,
			fp2.Inline,
			fp6.Inline,
			fp6.Mul,
		}
		if err := bavard.Generate(rootPath+strings.ToLower(fp6Name)+".go", src, d,
			bavard.Package(d.Fpackage),
			bavard.Apache2("ConsenSys AG", 2020),
			bavard.GeneratedBy("gurvy/internal/generators"),
		); err != nil {
			return err
		}
	}

	// fp12
	if d.MakeFp12 {
		src := []string{
			fp12.Base,
			fp2.Inline,
			fp6.Inline,
			fp12.Inline,
			fp12.Mul,
			// fp12.Frobenius,
			// fp12.Expt,
		}
		if err := bavard.Generate(rootPath+strings.ToLower(fp12Name)+".go", src, d,
			bavard.Package(d.Fpackage),
			bavard.Apache2("ConsenSys AG", 2020),
			bavard.GeneratedBy("gurvy/internal/generators"),
		); err != nil {
			return err
		}
	}

	return nil
}
