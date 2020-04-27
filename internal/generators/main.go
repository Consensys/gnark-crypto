package main

import (
	"os"

	"github.com/consensys/gurvy/internal/generators/template/generator"
)

//go:generate go run main.go
func main() {

	// -------------------------------------------------------------------------------------------------
	// bls381
	bls381 := generator.GenerateData{
		Fpackage:      "bls381",
		RootPath:      "../../bls381/",
		FpModulus:     "4002409555221667393417789825735904156556882819939007885332058136124031650490837864442687629129015664037894272559787",
		FpName:        "fp",
		FrModulus:     "52435875175126190479447740508185965837690552500527637822603658699938581184513",
		FrName:        "fr",
		Fp2Name:       "e2",
		Fp2NonResidue: "-1",
		Fp6Name:       "e6",
		Fp6NonResidue: "1,1",
		Fp12Name:      "e12",
		T:             "15132376222941642752",
		TNeg:          true,
		PointName:     "G",
		ThirdRootOne:  "4002409555221667392624310435006688643935503118305586438271171395842971157480381377015405980053539358417135540939436",
		Lambda:        "228988810152649578064853576960394133503",
		Size1:         "128",
		Size2:         "128",
	}

	// -------------------------------------------------------------------------------------------------
	// bls377
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
		ThirdRootOne:  "80949648264912719408558363140637477264845294720710499478137287262712535938301461879813459410945",
		Lambda:        "91893752504881257701523279626832445440",
		Size1:         "129",
		Size2:         "127",
	}

	// -------------------------------------------------------------------------------------------------
	// bn256
	bn256 := generator.GenerateData{
		Fpackage:      "bn256",
		RootPath:      "../../bn256/",
		FpModulus:     "21888242871839275222246405745257275088696311157297823662689037894645226208583",
		FpName:        "fp",
		FrModulus:     "21888242871839275222246405745257275088548364400416034343698204186575808495617",
		FrName:        "fr",
		Fp2Name:       "e2",
		Fp2NonResidue: "-1",
		Fp6Name:       "e6",
		Fp6NonResidue: "9,1",
		Fp12Name:      "e12",
		T:             "4965661367192848881",
		TNeg:          false,
		PointName:     "G",
		ThirdRootOne:  "2203960485148121921418603742825762020974279258880205651966",
		Lambda:        "4407920970296243842393367215006156084916469457145843978461",
		Size1:         "65",
		Size2:         "191",
	}

	curve := [3]generator.GenerateData{
		bls381,
		bls377,
		bn256,
	}

	for _, d := range curve {
		// create folder for the cruve
		if err := os.MkdirAll(d.RootPath, 0700); err != nil {
			panic(err)
		}
		if err := generator.GenerateCurve(d); err != nil {
			panic(err)
		}
	}
}
