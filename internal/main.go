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
	confs[1].CurveName = "bls381"
	confs[2].CurveName = "bls377"

	for i := 0; i < 3; i++ {

		err := generator.GenerateFq12over6over2(confs[i], outputDirs[i])
		if err != nil {
			fmt.Printf("\n%s\n", err.Error())
			os.Exit(-1)
		}

		confs[i].CoordType = "fp.Element"
		confs[i].PointName = "g1"
		err = generator.GeneratePoint(confs[i], outputDirs[i])
		if err != nil {
			fmt.Printf("\n%s\n", err.Error())
			os.Exit(-1)
		}

		confs[i].CoordType = "E2"
		confs[i].PointName = "g2"
		err = generator.GeneratePoint(confs[i], outputDirs[i])
		if err != nil {
			fmt.Printf("\n%s\n", err.Error())
			os.Exit(-1)
		}

	}

}
