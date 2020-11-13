package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/consensys/goff/field"

	"github.com/consensys/bavard"
	"github.com/consensys/goff/generator"
	"github.com/consensys/gurvy/internal/asm/amd64"
	"github.com/consensys/gurvy/internal/templates/fq12over6over2"
	"github.com/consensys/gurvy/internal/templates/pairing"
	"github.com/consensys/gurvy/internal/templates/point"
)

// Curve describes parameters of the curve useful for the templates
type Curve struct {
	Name string
	Fp   *field.Field
	Fr   *field.Field

	GLV              bool  // scalar mulitplication using GLV
	CofactorCleaning bool  // flag telling if the Cofactor cleaning is available
	CRange           []int // multiexp bucket method: generate inner methods (with const arrays) for each c

	outputDir string
}

const fpTower = "fptower"

// NewCurveConfig returns a struct initialized with the parameters needed for template generation
// (internal use in gurvy)
func NewCurveConfig(name, rTorsion, fpModulus string, glv bool, cc bool) Curve {
	name = strings.ToLower(name)
	conf := Curve{
		Name:             name,
		GLV:              glv,
		CofactorCleaning: cc,
	}

	var err error
	if conf.Fp, err = field.NewField("fp", "Element", fpModulus, false, "unset"); err != nil {
		panic(err)
	}
	if conf.Fr, err = field.NewField("fr", "Element", rTorsion, false, "unset"); err != nil {
		panic(err)
	}

	conf.outputDir = fmt.Sprintf("../%s/", name)

	// default range for C values in the multiExp
	conf.CRange = []int{4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 20, 21, 22}
	return conf
}

// GenerateBaseFields generates the base field fr and fp
func GenerateBaseFields(conf Curve) error {
	if err := generator.GenerateFF("fr", "Element", conf.Fr.Modulus, filepath.Join(conf.outputDir, "fr"), false); err != nil {
		return err
	}
	if err := generator.GenerateFF("fp", "Element", conf.Fp.Modulus, filepath.Join(conf.outputDir, "fp"), false); err != nil {
		return err
	}
	return nil
}

// GenerateFq12over6over2 generates a tower 2->6->12 over fp
func GenerateFq12over6over2(conf Curve) error {
	conf.outputDir = filepath.Join(conf.outputDir, "internal/", fpTower)

	type e2Config struct {
		Curve
		NbWords            int
		NbWordsIndexesFull []int
		Q, QInverse        []uint64
	}

	e2conf := e2Config{
		Curve:              conf,
		NbWords:            conf.Fp.NbWords,
		NbWordsIndexesFull: conf.Fp.NbWordsIndexesFull,
		Q:                  conf.Fp.Q,
		QInverse:           conf.Fp.QInverse,
	}

	bavardOpts := []func(*bavard.Bavard) error{
		bavard.Apache2("ConsenSys Software Inc.", 2020),
		bavard.Package(fpTower),
		bavard.GeneratedBy("gurvy"),
	}

	{
		// fq2 base
		src := []string{
			fq12over6over2.Fq2Common,
		}

		pathSrc := filepath.Join(conf.outputDir, "e2.go")

		if err := bavard.Generate(pathSrc, src, e2conf, bavardOpts...); err != nil {
			return err
		}
	}

	{

		// fq2 amd64
		src := []string{
			fq12over6over2.Fq2Amd64,
		}

		pathSrc := filepath.Join(conf.outputDir, "e2_amd64.go")

		if err := bavard.Generate(pathSrc, src, e2conf, bavardOpts...); err != nil {
			return err
		}

		// fq2 assembly
		pathAmd64 := filepath.Join(conf.outputDir, "e2_amd64.s")
		f, err := os.Create(pathAmd64)
		if err != nil {
			return err
		}
		defer f.Close()

		fp2, _ := field.NewField(conf.Name, "E2", conf.Fp.Modulus, false, "unset")
		fq2Amd64 := amd64.NewFq2Amd64(f, fp2, conf.Name)
		if err := fq2Amd64.Generate(); err != nil {
			return err
		}

	}

	{
		// fq2 fallback
		src := []string{
			fq12over6over2.Fq2FallBack,
		}

		bavardOpts := []func(*bavard.Bavard) error{
			bavard.Apache2("ConsenSys Software Inc.", 2020),
			bavard.Package(fpTower),
			bavard.GeneratedBy("gurvy"),
			bavard.BuildTag("!amd64"),
		}
		pathSrc := filepath.Join(e2conf.outputDir, "e2_fallback.go")

		if err := bavard.Generate(pathSrc, src, e2conf, bavardOpts...); err != nil {
			return err
		}

	}

	// fq2 tests
	{
		src := []string{
			fq12over6over2.Fq2Tests,
		}
		pathSrc := filepath.Join(conf.outputDir, "e2_test.go")
		if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
			return err
		}
	}

	{
		// fq6 base
		src := []string{
			fq12over6over2.Fq6,
		}

		bavardOpts := []func(*bavard.Bavard) error{
			bavard.Apache2("ConsenSys Software Inc.", 2020),
			bavard.GeneratedBy("gurvy"),
			bavard.Package(fpTower),
		}

		pathSrc := filepath.Join(conf.outputDir, "e6.go")

		if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
			return err
		}
	}

	{
		// fq6 tests
		src := []string{
			fq12over6over2.Fq6Tests,
		}
		pathSrc := filepath.Join(conf.outputDir, "e6_test.go")
		if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
			return err
		}
	}
	{
		// fq12 base
		src := []string{
			fq12over6over2.Fq12,
		}

		bavardOpts := []func(*bavard.Bavard) error{
			bavard.Apache2("ConsenSys Software Inc.", 2020),
			bavard.GeneratedBy("gurvy"),
			bavard.Package(fpTower),
		}

		pathSrc := filepath.Join(conf.outputDir, "e12.go")

		if err := bavard.Generate(pathSrc, src, e2conf, bavardOpts...); err != nil {
			return err
		}
	}

	{
		// fq12 tests
		src := []string{
			fq12over6over2.Fq12Tests,
		}
		pathSrc := filepath.Join(conf.outputDir, "e12_test.go")
		if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
			return err
		}
	}
	return nil
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

// GenerateMultiExpHelpers generates multi exp helpers functions
func GenerateMultiExpHelpers(conf Curve) error {

	bavardOpts := []func(*bavard.Bavard) error{
		bavard.Apache2("ConsenSys Software Inc.", 2020),
		bavard.Package(conf.Name),
		bavard.GeneratedBy("gurvy"),
		bavard.Funcs(helpers()),
	}

	// point code
	src := []string{
		point.MultiExpHelpers,
	}

	pathSrc := filepath.Join(conf.outputDir, "multiexp_helpers.go")
	if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
		return err
	}

	return nil
}

// GeneratePoint generates elliptic curve arithmetic
func GeneratePoint(_conf Curve, coordType, pointName string) error {
	type pointConfig struct {
		Curve
		CoordType  string
		PointName  string
		UnusedBits int
	}
	conf := pointConfig{
		Curve:     _conf,
		CoordType: coordType,
		PointName: pointName,
	}

	conf.UnusedBits = 64 - (conf.Fp.NbBits % 64)

	bavardOpts := []func(*bavard.Bavard) error{
		bavard.Apache2("ConsenSys Software Inc.", 2020),
		bavard.Package(conf.Name),
		bavard.GeneratedBy("gurvy"),
		bavard.Funcs(helpers()),
	}

	// point code (without multi exp)
	src := []string{
		point.Point,
	}

	pathSrc := filepath.Join(conf.outputDir, conf.PointName+".go")
	if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
		return err
	}

	// multi exp core code
	src = []string{
		point.MultiExpCore,
	}
	pathSrc = filepath.Join(conf.outputDir, conf.PointName+"_multiexp.go")
	if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
		return err
	}

	// point test
	src = []string{
		point.PointTests,
	}

	pathSrc = filepath.Join(conf.outputDir, conf.PointName+"_test.go")
	if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
		return err
	}

	return nil
}

// GenerateMarshal generates elliptic curve encoder and serialization code
func GenerateMarshal(_conf Curve) error {
	type pointConfig struct {
		Curve
		UnusedBits int
	}
	conf := pointConfig{
		Curve: _conf,
	}

	conf.UnusedBits = 64 - (conf.Fp.NbBits % 64)

	bavardOpts := []func(*bavard.Bavard) error{
		bavard.Apache2("ConsenSys Software Inc.", 2020),
		bavard.Package(conf.Name),
		bavard.GeneratedBy("gurvy"),
		bavard.Funcs(helpers()),
	}

	// encoder / decoder
	src := []string{
		point.Marshal,
	}

	pathSrc := filepath.Join(conf.outputDir, "marshal.go")
	if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
		return err
	}

	// tests
	src = []string{
		point.MarshalTests,
	}

	pathSrc = filepath.Join(conf.outputDir, "marshal_test.go")

	if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
		return err
	}

	return nil
}

// GeneratePairingTests generates elliptic curve arithmetic
func GeneratePairingTests(conf Curve) error {

	src := []string{
		pairing.PairingTests,
	}

	bavardOpts := []func(*bavard.Bavard) error{
		bavard.Apache2("ConsenSys Software Inc.", 2020),
		bavard.Package(conf.Name),
		bavard.GeneratedBy("gurvy"),
	}

	pathSrc := filepath.Join(conf.outputDir, "pairing_test.go")

	if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
		return err
	}

	return nil
}

// GenerateDoc generates package level doc
func GenerateDoc(conf Curve) error {

	bavardOpts := []func(*bavard.Bavard) error{
		bavard.Apache2("ConsenSys Software Inc.", 2020),
		bavard.Package(conf.Name, "provides efficient elliptic curve and pairing implementation for "+conf.Name),
		bavard.GeneratedBy("gurvy"),
	}

	pathSrc := filepath.Join(conf.outputDir, "doc.go")

	if err := bavard.Generate(pathSrc, []string{""}, conf, bavardOpts...); err != nil {
		return err
	}

	return nil
}
