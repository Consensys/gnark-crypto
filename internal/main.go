package main

import (
	"fmt"
	"os"

	"github.com/consensys/gurvy/internal/generator"
)

//go:generate go run main.go
func main() {

	var confs [3]generator.CurveConfig
	var outputDirs [3]string

	outputDirs[0] = "../bn256/"
	outputDirs[1] = "../bls381/"
	outputDirs[2] = "../bls377/"

	confs[0].CurveName = "bn256"
	confs[0].RTorsion = "21888242871839275222246405745257275088548364400416034343698204186575808495617"
	confs[0].FpModulus = "21888242871839275222246405745257275088696311157297823662689037894645226208583"

	confs[1].CurveName = "bls381"
	confs[1].RTorsion = "52435875175126190479447740508185965837690552500527637822603658699938581184513"
	confs[1].FpModulus = "4002409555221667393417789825735904156556882819939007885332058136124031650490837864442687629129015664037894272559787"

	confs[2].CurveName = "bls377"
	confs[2].RTorsion = "8444461749428370424248824938781546531375899335154063827935233455917409239041"
	confs[2].FpModulus = "258664426012969094010652733694893533536393512754914660539884262666720468348340822774968888139573360124440321458177"

	for i := 0; i < 3; i++ {

		if err := generator.GenerateBaseFields(confs[i], outputDirs[i]); err != nil {
			fmt.Printf("\n%s\n", err.Error())
			os.Exit(-1)
		}

		if err := generator.GenerateFq12over6over2(confs[i], outputDirs[i]); err != nil {
			fmt.Printf("\n%s\n", err.Error())
			os.Exit(-1)
		}

		confs[i].CoordType = "fp.Element"
		confs[i].PointName = "g1"
		if err := generator.GeneratePoint(confs[i], outputDirs[i]); err != nil {
			fmt.Printf("\n%s\n", err.Error())
			os.Exit(-1)
		}

		confs[i].CoordType = "E2"
		confs[i].PointName = "g2"
		if err := generator.GeneratePoint(confs[i], outputDirs[i]); err != nil {
			fmt.Printf("\n%s\n", err.Error())
			os.Exit(-1)
		}

		err := generator.GeneratePairing(confs[i], outputDirs[i])
		if err != nil {
			fmt.Printf("\n%s\n", err.Error())
			os.Exit(-1)
		}

	}

}
