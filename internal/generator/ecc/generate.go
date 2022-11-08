package ecc

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	packageName := strings.ReplaceAll(conf.Name, "-", "")

	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "multiexp.go"), Templates: []string{"multiexp.go.tmpl"}},
		{File: filepath.Join(baseDir, "multiexp_affine.go"), Templates: []string{"multiexp_affine.go.tmpl"}},
		{File: filepath.Join(baseDir, "multiexp_test.go"), Templates: []string{"tests/multiexp.go.tmpl"}},
		{File: filepath.Join(baseDir, "marshal.go"), Templates: []string{"marshal.go.tmpl"}},
		{File: filepath.Join(baseDir, "marshal_test.go"), Templates: []string{"tests/marshal.go.tmpl"}},
	}
	conf.Package = packageName
	funcs := make(template.FuncMap)
	funcs["last"] = func(x int, a interface{}) bool {
		return x == reflect.ValueOf(a).Len()-1
	}
	funcs["lastC"] = func(c int) int {
		// lastC := (fr.Limbs * 64) - (c * (fr.Limbs * 64 / c))
		// if c divides fr.Limbs * 64;
		n := (conf.Fr.NbWords * 64)
		if n%c == 0 {
			return c
		}
		return n - (c * (n / c))
	}
	funcs["contains"] = func(v int, s []int) bool {
		for _, sv := range s {
			if v == sv {
				return true
			}
		}
		return false
	}
	// TODO @gbotrel fix me. need to generate usual C, and missing lastC for bucket size.
	conf.G1.CRange = make([]int, 23)
	conf.G2.CRange = make([]int, 23)
	for i := 0; i < len(conf.G1.CRange); i++ {
		conf.G1.CRange[i] = i + 1
		conf.G2.CRange[i] = i + 1
	}
	bavardOpts := []func(*bavard.Bavard) error{bavard.Funcs(funcs)}
	if err := bgen.GenerateWithOptions(conf, packageName, "./ecc/template", bavardOpts, entries...); err != nil {
		return err
	}

	// hash To curve

	genHashToCurve := func(point *config.Point, suite config.HashSuite) error {
		if suite == nil { //Nothing to generate. Bypass
			return nil
		}

		entries = []bavard.Entry{
			{File: filepath.Join(baseDir, fmt.Sprintf("hash_to_%s.go", point.PointName)), Templates: []string{"hash_to_curve.go.tmpl", "sswu.go.tmpl", "svdw.go.tmpl"}},
			{File: filepath.Join(baseDir, fmt.Sprintf("hash_to_%s_test.go", point.PointName)), Templates: []string{"tests/hash_to_curve.go.tmpl"}}}

		hashConf := suite.GetInfo(conf.Fp, point, conf.Name)

		funcs := make(template.FuncMap)
		funcs["asElement"] = hashConf.Field.Base.WriteElement
		bavardOpts := []func(*bavard.Bavard) error{bavard.Funcs(funcs)}

		return bgen.GenerateWithOptions(hashConf, packageName, "./ecc/template", bavardOpts, entries...)
	}

	if err := genHashToCurve(&conf.G1, conf.HashE1); err != nil {
		return err
	}
	if err := genHashToCurve(&conf.G2, conf.HashE2); err != nil {
		return err
	}

	// G1
	entries = []bavard.Entry{
		{File: filepath.Join(baseDir, "g1.go"), Templates: []string{"point.go.tmpl"}},
		{File: filepath.Join(baseDir, "g1_test.go"), Templates: []string{"tests/point.go.tmpl"}},
	}
	g1 := pconf{conf, conf.G1}
	if err := bgen.Generate(g1, packageName, "./ecc/template", entries...); err != nil {
		return err
	}

	// G2
	entries = []bavard.Entry{
		{File: filepath.Join(baseDir, "g2.go"), Templates: []string{"point.go.tmpl"}},
		{File: filepath.Join(baseDir, "g2_test.go"), Templates: []string{"tests/point.go.tmpl"}},
	}
	g2 := pconf{conf, conf.G2}
	return bgen.Generate(g2, packageName, "./ecc/template", entries...)
}

type pconf struct {
	config.Curve
	config.Point
}
