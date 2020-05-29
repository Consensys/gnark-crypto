package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/consensys/bavard"
	"github.com/consensys/gurvy/internal/generators/curve"
	"github.com/consensys/gurvy/internal/generators/pairing"
	"github.com/consensys/gurvy/internal/generators/tower"
)

// TODO move all this curve data to config file(s)?
// -------------------------------------------------------------------------------------------------
// bls381
var bls381 curve.Data = curve.Data{
	Fpackage:        "bls381",
	FpModulus:       "4002409555221667393417789825735904156556882819939007885332058136124031650490837864442687629129015664037894272559787",
	FrModulus:       "52435875175126190479447740508185965837690552500527637822603658699938581184513",
	Fp2NonResidue:   "-1",
	Fp6NonResidue:   "1,1",
	EmbeddingDegree: 12,
	T:               "15132376222941642752",
	TNeg:            true,
	ThirdRootOne:    "4002409555221667392624310435006688643935503118305586438271171395842971157480381377015405980053539358417135540939436",
	Lambda:          "228988810152649578064853576960394133503",
	Size1:           "128",
	Size2:           "128",
}

// -------------------------------------------------------------------------------------------------
// bls377
var bls377 curve.Data = curve.Data{
	Fpackage:        "bls377",
	FpModulus:       "258664426012969094010652733694893533536393512754914660539884262666720468348340822774968888139573360124440321458177",
	FrModulus:       "8444461749428370424248824938781546531375899335154063827935233455917409239041",
	Fp2NonResidue:   "5",
	Fp6NonResidue:   "0,1",
	EmbeddingDegree: 12,
	T:               "9586122913090633729",
	TNeg:            false,
	ThirdRootOne:    "80949648264912719408558363140637477264845294720710499478137287262712535938301461879813459410945",
	Lambda:          "91893752504881257701523279626832445440",
	Size1:           "129",
	Size2:           "127",
}

// -------------------------------------------------------------------------------------------------
// bn256
var bn256 curve.Data = curve.Data{
	Fpackage:        "bn256",
	FpModulus:       "21888242871839275222246405745257275088696311157297823662689037894645226208583",
	FrModulus:       "21888242871839275222246405745257275088548364400416034343698204186575808495617",
	Fp2NonResidue:   "-1",
	Fp6NonResidue:   "9,1",
	EmbeddingDegree: 12,
	T:               "4965661367192848881",
	TNeg:            false,
	ThirdRootOne:    "2203960485148121921418603742825762020974279258880205651966",
	Lambda:          "4407920970296243842393367215006156084916469457145843978461",
	Size1:           "65",
	Size2:           "191",
}

// -------------------------------------------------------------------------------------------------
// BW6-781
var bw6_761 curve.Data = curve.Data{
	Fpackage:        "bw6_761",
	FpModulus:       "6891450384315732539396789682275657542479668912536150109513790160209623422243491736087683183289411687640864567753786613451161759120554247759349511699125301598951605099378508850372543631423596795951899700429969112842764913119068299",
	FrModulus:       bls377.FpModulus,
	Fp2NonResidue:   "-1",
	Fp6NonResidue:   "1,1",
	EmbeddingDegree: 6,
	T:               "4371281006305286848163485150587564217350348383473813272171392617577304305730246023460948554022786", // TODO this is the hard part exponent.  Implement the optimized hard part from Appendix B of https://eprint.iacr.org/2020/351.pdf
	TNeg:            false,
	ThirdRootOne:    "1968985824090209297278610739700577151397666382303825728450741611566800370218827257750865013421937292370006175842381275743914023380727582819905021229583192207421122272650305267822868639090213645505120388400344940985710520836292650",
	Lambda:          "80949648264912719408558363140637477264845294720710499478137287262712535938301461879813459410945",
	Size1:           "65",
	Size2:           "316",
}

//go:generate go run main.go
func main() {

	curves := [...]curve.Data{
		bls381,
		bls377,
		bn256,
		bw6_761,
	}

	for _, d := range curves {

		// TODO refactor calls to bavard.Generate, exec.Command

		d.Init()

		// generate curve.C
		{
			src := []string{
				curve.CTemplate,
			}
			if err := bavard.Generate("curve/c.go", src, d,
				bavard.Package("curve"),
				bavard.Apache2("ConsenSys AG", 2020),
				bavard.GeneratedBy("gurvy/internal/generators"),
			); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(-1)
			}
		}

		// generate primefield (uses curve.C)
		{
			cmd := exec.Command("go", "run", "./primefields/main/main.go")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				// fmt.Fprintln(os.Stderr, err)
				os.Exit(-1)
			}
		}

		// generate tower generator (uses curve.C, primefield)
		{
			src := []string{
				tower.TwoInvTemplate,
			}
			if err := bavard.Generate("tower/twoinv.go", src, d,
				bavard.Package("tower"),
				bavard.Apache2("ConsenSys AG", 2020),
				bavard.GeneratedBy("gurvy/internal/generators"),
			); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(-1)
			}
		}

		// generate tower
		{
			cmd := exec.Command("go", "run", "./tower/main/main.go")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				// fmt.Fprintln(os.Stderr, err)
				os.Exit(-1)
			}
		}

		// generate pairing generator (uses curve.C, tower)
		{
			src := []string{
				pairing.ConstantsTemplate,
			}
			if err := bavard.Generate("pairing/constants.go", src, d,
				bavard.Package("pairing"),
				bavard.Apache2("ConsenSys AG", 2020),
				bavard.GeneratedBy("gurvy/internal/generators"),
			); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(-1)
			}
		}

		// generate pairing
		{
			cmd := exec.Command("go", "run", "./pairing/main/main.go")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				// fmt.Fprintln(os.Stderr, err)
				os.Exit(-1)
			}
		}

		// generate gpoint
	}
}
