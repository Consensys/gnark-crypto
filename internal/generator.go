package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/consensys/bavard"
	"github.com/consensys/goff/field"
	"github.com/consensys/goff/generator"
	"github.com/consensys/gurvy/internal/asm/amd64"
)

var bgen = bavard.NewBatchGenerator(copyrightHolder, "gurvy")

//go:generate go run generator.go
func main() {
	var wg sync.WaitGroup
	for _, conf := range []curveConfig{
		// BN256
		{
			Name: "bn256",
			fr:   "21888242871839275222246405745257275088548364400416034343698204186575808495617",
			fp:   "21888242871839275222246405745257275088696311157297823662689037894645226208583",
			G1: pointConfig{
				CoordType:        "fp.Element",
				PointName:        "g1",
				GLV:              true,
				CofactorCleaning: false,
				CRange:           defaultCRange(),
			},
			G2: pointConfig{
				CoordType:        "fptower.E2",
				PointName:        "g2",
				GLV:              true,
				CofactorCleaning: true,
				CRange:           defaultCRange(),
			},
		},

		// BLS377
		{
			Name: "bls377",
			fr:   "8444461749428370424248824938781546531375899335154063827935233455917409239041",
			fp:   "258664426012969094010652733694893533536393512754914660539884262666720468348340822774968888139573360124440321458177",
			G1: pointConfig{
				CoordType:        "fp.Element",
				PointName:        "g1",
				GLV:              true,
				CofactorCleaning: true,
				CRange:           defaultCRange(),
			},
			G2: pointConfig{
				CoordType:        "fptower.E2",
				PointName:        "g2",
				GLV:              true,
				CofactorCleaning: true,
				CRange:           defaultCRange(),
			},
		},

		// BLS381
		{
			Name: "bls381",
			fr:   "52435875175126190479447740508185965837690552500527637822603658699938581184513",
			fp:   "4002409555221667393417789825735904156556882819939007885332058136124031650490837864442687629129015664037894272559787",
			G1: pointConfig{
				CoordType:        "fp.Element",
				PointName:        "g1",
				GLV:              true,
				CofactorCleaning: true,
				CRange:           defaultCRange(),
			},
			G2: pointConfig{
				CoordType:        "fptower.E2",
				PointName:        "g2",
				GLV:              true,
				CofactorCleaning: true,
				CRange:           defaultCRange(),
			},
		},

		// BW761
		{
			Name: "bw761",
			fr:   "258664426012969094010652733694893533536393512754914660539884262666720468348340822774968888139573360124440321458177",
			fp:   "6891450384315732539396789682275657542479668912536150109513790160209623422243491736087683183289411687640864567753786613451161759120554247759349511699125301598951605099378508850372543631423596795951899700429969112842764913119068299",
			G1: pointConfig{
				CoordType:        "fp.Element",
				PointName:        "g1",
				GLV:              true,
				CofactorCleaning: true,
				CRange:           []int{4, 5, 8, 16},
			},
			G2: pointConfig{
				CoordType:        "fp.Element",
				PointName:        "g2",
				GLV:              true,
				CofactorCleaning: true,
				CRange:           []int{4, 5, 8, 16},
			},
		},
	} {
		wg.Add(1)
		// for each curve, generate the needed files
		go func(conf curveConfig) {
			defer wg.Done()
			doc := "provides efficient elliptic curve and pairing implementation for " + conf.Name
			conf.Fp, _ = field.NewField("fp", "Element", conf.fp)
			conf.Fr, _ = field.NewField("fr", "Element", conf.fr)
			conf.FpUnusedBits = 64 - (conf.Fp.NbBits % 64)
			dir := filepath.Join(baseDir, conf.Name)
			conf.dir = dir

			// generate base fields
			assertNoError(generator.GenerateFF(conf.Fr, filepath.Join(conf.dir, "fr")))
			assertNoError(generator.GenerateFF(conf.Fp, filepath.Join(conf.dir, "fp")))

			g1 := pconf{conf, conf.G1}
			g2 := pconf{conf, conf.G2}

			entriesF := []bavard.EntryF{
				{File: filepath.Join(dir, "multiexp.go"), TemplateF: []string{"multiexp.go.tmpl"}},
				{File: filepath.Join(dir, "multiexp_test.go"), TemplateF: []string{"tests/multiexp.go.tmpl"}},
				{File: filepath.Join(dir, "marshal.go"), TemplateF: []string{"marshal.go.tmpl"}, PackageDoc: doc},
				{File: filepath.Join(dir, "marshal_test.go"), TemplateF: []string{"tests/marshal.go.tmpl"}},
			}
			if err := bgen.GenerateF(conf, conf.Name, "./templates/point", entriesF...); err != nil {
				panic(err)
			}

			// G1
			entriesF = []bavard.EntryF{
				{File: filepath.Join(dir, "g1.go"), TemplateF: []string{"point.go.tmpl"}},
				{File: filepath.Join(dir, "g1_test.go"), TemplateF: []string{"tests/point.go.tmpl"}},
			}
			if err := bgen.GenerateF(g1, conf.Name, "./templates/point", entriesF...); err != nil {
				panic(err)
			}

			// G2
			entriesF = []bavard.EntryF{
				{File: filepath.Join(dir, "g2.go"), TemplateF: []string{"point.go.tmpl"}},
				{File: filepath.Join(dir, "g2_test.go"), TemplateF: []string{"tests/point.go.tmpl"}},
			}
			if err := bgen.GenerateF(g2, conf.Name, "./templates/point", entriesF...); err != nil {
				panic(err)
			}

			if conf.Name != "bw761" {
				assertNoError(GenerateFq12over6over2(conf))
				if err := bgen.GenerateF(conf, conf.Name, "./templates/pairing", bavard.EntryF{
					File: filepath.Join(dir, "pairing_test.go"), TemplateF: []string{"tests/pairing.go.tmpl"},
				}); err != nil {
					panic(err)
				}
			}

			// twisted Edwards
			if err := bgen.GenerateF(conf, "twistededwards", "./templates/point", bavard.EntryF{
				File: filepath.Join(dir, "/twistededwards/point.go"), TemplateF: []string{"pointtwistededwards.go.tmpl"},
			}); err != nil {
				panic(err)
			}

		}(conf)

	}
	wg.Wait()

	// run go fmt on whole directory
	cmd := exec.Command("gofmt", "-s", "-w", "../")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	assertNoError(cmd.Run())
}

func assertNoError(err error) {
	if err != nil {
		fmt.Printf("\n%s\n", err.Error())
		os.Exit(-1)
	}
}

// curveConfig describes parameters of the curve useful for the templates
type curveConfig struct {
	Name string
	fp   string
	fr   string

	Fp           *field.Field
	Fr           *field.Field
	FpUnusedBits int
	G1           pointConfig
	G2           pointConfig
	dir          string
}

type pointConfig struct {
	CoordType        string
	PointName        string
	GLV              bool  // scalar mulitplication using GLV
	CofactorCleaning bool  // flag telling if the Cofactor cleaning is available
	CRange           []int // multiexp bucket method: generate inner methods (with const arrays) for each c
}

type pconf struct {
	curveConfig
	pointConfig
}

const (
	fpTower         = "fptower"
	copyrightHolder = "ConsenSys Software Inc."
	baseDir         = "../"
)

func defaultCRange() []int {
	// default range for C values in the multiExp
	return []int{4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 20, 21, 22}
}

// GenerateFq12over6over2 generates a tower 2->6->12 over fp
func GenerateFq12over6over2(conf curveConfig) error {
	dir := filepath.Join(conf.dir, "internal", fpTower)
	entries := []bavard.EntryF{
		{File: filepath.Join(dir, "e2.go"), TemplateF: []string{"fq2.go.tmpl"}},
		{File: filepath.Join(dir, "e6.go"), TemplateF: []string{"fq6.go.tmpl"}},
		{File: filepath.Join(dir, "e12.go"), TemplateF: []string{"fq12.go.tmpl"}},
		{File: filepath.Join(dir, "e2_amd64.go"), TemplateF: []string{"amd64.fq2.go.tmpl"}},
		{File: filepath.Join(dir, "e2_fallback.go"), TemplateF: []string{"fallback.fq2.go.tmpl"}, BuildTag: "!amd64"},
		{File: filepath.Join(dir, "e2_test.go"), TemplateF: []string{"tests/fq2.go.tmpl"}},
		{File: filepath.Join(dir, "e6_test.go"), TemplateF: []string{"tests/fq6.go.tmpl"}},
		{File: filepath.Join(dir, "e12_test.go"), TemplateF: []string{"tests/fq12.go.tmpl"}},
		{File: filepath.Join(dir, "asm.go"), TemplateF: []string{"asm.go.tmpl"}, BuildTag: "!noadx"},
		{File: filepath.Join(dir, "asm_noadx.go"), TemplateF: []string{"asm_noadx.go.tmpl"}, BuildTag: "noadx"},
	}

	if err := bgen.GenerateF(conf, fpTower, "./templates/fq12over6over2", entries...); err != nil {
		return err
	}

	{
		// fq2 assembly
		fName := filepath.Join(dir, "e2_amd64.s")
		f, err := os.Create(fName)
		if err != nil {
			return err
		}

		Fq2Amd64 := amd64.NewFq2Amd64(f, conf.Fp, conf.Name)
		if err := Fq2Amd64.Generate(); err != nil {
			_ = f.Close()
			return err
		}
		_ = f.Close()

		cmd := exec.Command("asmfmt", "-w", fName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}

	}

	return nil
}
