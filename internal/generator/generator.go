package generator

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gurvy/internal/templates/fq12over6over2"
	"github.com/consensys/gurvy/internal/templates/pairing"
	"github.com/consensys/gurvy/internal/templates/point"
)

// CurveConfig describes parameters of the curve useful for the templates
type CurveConfig struct {
	CurveName string
	CoordType string
	PointName string
	RTorsion  string
}

// GenerateFq12over6over2 generates a tower 2->6->12 over fp
func GenerateFq12over6over2(conf CurveConfig, outputDir string) error {

	// fq2 base
	src := []string{
		fq12over6over2.Fq2Common,
	}

	bavardOpts := []func(*bavard.Bavard) error{
		bavard.Apache2("ConsenSys AG", 2020),
		bavard.Package(conf.CurveName),
		bavard.GeneratedBy("gurvy"),
	}
	pathSrc := filepath.Join(outputDir, "e2.go")

	if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
		return err
	}

	// fq2 tests
	src = []string{
		fq12over6over2.Fq2Tests,
	}
	pathSrc = filepath.Join(outputDir, "e2_test.go")
	if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
		return err
	}

	// fq6 base
	src = []string{
		fq12over6over2.Fq6,
	}

	bavardOpts = []func(*bavard.Bavard) error{
		bavard.Apache2("ConsenSys AG", 2020),
		bavard.Package(conf.CurveName),
	}

	pathSrc = filepath.Join(outputDir, "e6.go")

	if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
		return err
	}

	// fq6 tests
	src = []string{
		fq12over6over2.Fq6Tests,
	}
	pathSrc = filepath.Join(outputDir, "e6_test.go")
	if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
		return err
	}

	// fq12 base
	src = []string{
		fq12over6over2.Fq12,
	}

	bavardOpts = []func(*bavard.Bavard) error{
		bavard.Apache2("ConsenSys AG", 2020),
		bavard.Package(conf.CurveName),
	}

	pathSrc = filepath.Join(outputDir, "e12.go")

	if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
		return err
	}

	// fq12 tests
	src = []string{
		fq12over6over2.Fq12Tests,
	}
	pathSrc = filepath.Join(outputDir, "e12_test.go")
	if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
		return err
	}

	return nil
}

// GeneratePoint generates elliptic curve arithmetic
func GeneratePoint(conf CurveConfig, outputDir string) error {

	bavardOpts := []func(*bavard.Bavard) error{
		bavard.Apache2("ConsenSys AG", 2020),
		bavard.Package(conf.CurveName),
		bavard.GeneratedBy("gurvy"),
	}

	// point code
	src := []string{
		point.Point,
	}

	pathSrc := filepath.Join(outputDir, conf.PointName+".go")
	if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
		return err
	}
	if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
		return err
	}

	// point test
	src = []string{
		point.PointTests,
	}

	pathSrc = filepath.Join(outputDir, conf.PointName+"_test.go")
	if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
		return err
	}
	if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
		return err
	}

	return nil
}

// GeneratePairing generates elliptic curve arithmetic
func GeneratePairing(conf CurveConfig, outputDir string) error {

	src := []string{
		pairing.PairingTests,
	}

	bavardOpts := []func(*bavard.Bavard) error{
		bavard.Apache2("ConsenSys AG", 2020),
		bavard.Package(conf.CurveName),
		bavard.GeneratedBy("gurvy"),
	}

	pathSrc := filepath.Join(outputDir, "pairing_test.go")

	if err := bavard.Generate(pathSrc, src, conf, bavardOpts...); err != nil {
		return err
	}

	return nil
}
