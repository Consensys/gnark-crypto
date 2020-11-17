package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"text/template"

	"github.com/consensys/bavard"
	"github.com/consensys/goff/field"
	"github.com/consensys/goff/generator"
	"github.com/consensys/gurvy/internal/asm/amd64"
	"github.com/consensys/gurvy/internal/templates/fq12over6over2"
	"github.com/consensys/gurvy/internal/templates/pairing"
	"github.com/consensys/gurvy/internal/templates/point"
)

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
				CRange:           []int{4, 8, 16},
			},
			G2: pointConfig{
				CoordType:        "fp.Element",
				PointName:        "g2",
				GLV:              true,
				CofactorCleaning: true,
				CRange:           []int{4, 8, 16},
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
			conf.dir = filepath.Join(baseDir, conf.Name)

			// generate base fields
			assertNoError(generator.GenerateFF(conf.Fr, filepath.Join(conf.dir, "fr")))
			assertNoError(generator.GenerateFF(conf.Fp, filepath.Join(conf.dir, "fp")))

			g1 := pconf{conf, conf.G1}
			g2 := pconf{conf, conf.G2}

			toGenerate := []genOpts{
				{data: conf, dir: conf.dir, file: "doc.go", doc: doc},
				{data: conf, dir: conf.dir, file: "multiexp_helpers.go", templates: []string{point.MultiExpHelpers}},
				{data: conf, dir: conf.dir, file: "marshal.go", templates: []string{point.Marshal}},
				{data: conf, dir: conf.dir, file: "marshal_test.go", templates: []string{point.MarshalTests}},

				{data: g1, dir: conf.dir, file: "g1.go", templates: []string{point.Point}},
				{data: g1, dir: conf.dir, file: "g1_test.go", templates: []string{point.PointTests}},
				{data: g1, dir: conf.dir, file: "g1_multiexp.go", templates: []string{point.MultiExpCore}},

				{data: g2, dir: conf.dir, file: "g2.go", templates: []string{point.Point}},
				{data: g2, dir: conf.dir, file: "g2_test.go", templates: []string{point.PointTests}},
				{data: g2, dir: conf.dir, file: "g2_multiexp.go", templates: []string{point.MultiExpCore}},
			}

			if conf.Name != "bw761" {
				assertNoError(GenerateFq12over6over2(conf))
				toGenerate = append(toGenerate, genOpts{
					data: conf, dir: conf.dir, file: "pairing_test.go", templates: []string{pairing.PairingTests},
				})
			}

			for _, g := range toGenerate {
				generate(g)
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
	for _, g := range []genOpts{
		{data: conf, dir: dir, file: "e2.go", templates: []string{fq12over6over2.Fq2Common}},
		{data: conf, dir: dir, file: "e2_amd64.go", templates: []string{fq12over6over2.Fq2Amd64}},
		{data: conf, dir: dir, file: "e2_test.go", templates: []string{fq12over6over2.Fq2Tests}},
		{data: conf, dir: dir, file: "e6.go", templates: []string{fq12over6over2.Fq6}},
		{data: conf, dir: dir, file: "e6_test.go", templates: []string{fq12over6over2.Fq6Tests}},
		{data: conf, dir: dir, file: "e12.go", templates: []string{fq12over6over2.Fq12}},
		{data: conf, dir: dir, file: "e12_test.go", templates: []string{fq12over6over2.Fq12Tests}},
		{data: conf, dir: dir, file: "e2_fallback.go", templates: []string{fq12over6over2.Fq2FallBack}, buildTag: "!amd64"},
	} {
		generate(g)
	}

	{
		// fq2 assembly
		fName := filepath.Join(dir, "e2_amd64.s")
		f, err := os.Create(fName)
		if err != nil {
			return err
		}

		// TODO dirty, we need that so that generated assembly code points to q"E2"
		fp2, _ := field.NewField(conf.Name, "E2", conf.Fp.Modulus)
		Fq2Amd64 := amd64.NewFq2Amd64(f, fp2, conf.Name)
		if err := Fq2Amd64.Generate(); err != nil {
			_ = f.Close()
			return err
		}
		_ = f.Close()

	}

	return nil
}

type genOpts struct {
	opts      []func(*bavard.Bavard) error
	file      string
	templates []string
	buildTag  string
	dir       string
	doc       string
	data      interface{}
}

func generate(g genOpts) {
	opts := []func(*bavard.Bavard) error{
		bavard.Apache2(copyrightHolder, 2020),
		bavard.GeneratedBy("gurvy"),
		bavard.Funcs(helpers()),
		bavard.Format(false),
		bavard.Import(false),
	}
	if g.buildTag != "" {
		opts = append(opts, bavard.BuildTag(g.buildTag))
	}
	file := filepath.Join(g.dir, g.file)

	opts = append(opts, bavard.Package(filepath.Base(filepath.Dir(file)), g.doc))

	if err := bavard.Generate(file, g.templates, g.data, opts...); err != nil {
		panic(err)
	}
}

// Template helpers (txt/template)
func helpers() template.FuncMap {
	// functions used in template
	return template.FuncMap{
		"divides": divides,
	}
}

// return true if c1 divides c2, that is, c2 % c1 == 0
func divides(c1, c2 interface{}) bool {
	switch cc1 := c1.(type) {
	case int:
		switch cc2 := c2.(type) {
		case int:
			return cc2%cc1 == 0
		case string:
			c2Int, err := strconv.Atoi(cc2)
			if err != nil {
				panic(err)
			}
			return c2Int%cc1 == 0
		}
	case string:
		c1Int, err := strconv.Atoi(cc1)
		if err != nil {
			panic(err)
		}
		switch cc2 := c2.(type) {
		case int:
			return cc2%c1Int == 0
		case string:
			c2Int, err := strconv.Atoi(cc2)
			if err != nil {
				panic(err)
			}
			return c2Int%c1Int == 0
		}
	}
	panic("unexpected type")
}
