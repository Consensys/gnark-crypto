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
		{File: filepath.Join(baseDir, "multiexp_jacobian.go"), Templates: []string{"multiexp_jacobian.go.tmpl"}},
		{File: filepath.Join(baseDir, "multiexp_test.go"), Templates: []string{"tests/multiexp.go.tmpl"}},
		{File: filepath.Join(baseDir, "marshal.go"), Templates: []string{"marshal.go.tmpl"}},
		{File: filepath.Join(baseDir, "marshal_test.go"), Templates: []string{"tests/marshal.go.tmpl"}},
	}
	conf.Package = packageName
	funcs := make(template.FuncMap)
	funcs["last"] = func(x int, a interface{}) bool {
		return x == reflect.ValueOf(a).Len()-1
	}
	lastC := func(c int) int {
		n := (conf.Fr.NbBits + 1) // +1 for the potential carry of the NAF decomposition
		if n%c == 0 {
			return c
		}
		return n - (c * (n / c))
	}
	batchSize := func(c int) int {
		// nbBuckets := (1 << (c - 1))
		// if c <= 12 {
		// 	return nbBuckets/10 + 3*c
		// }
		// if c <= 14 {
		// 	return nbBuckets/15
		// }
		// return nbBuckets / 20
		// TODO @gbotrel / @yelhousni this need a better heuristic
		// in theory, larger batch size == less inversions
		// but if nbBuckets is small, then a large batch size will produce lots of collisions
		// and queue ops.
		// there is probably a cache-friendlyness factor at play here too.
		switch c {
		case 10:
			return 80
		case 11:
			return 150
		case 12:
			return 200
		case 13:
			return 350
		case 14:
			return 400
		case 15:
			return 500
		default:
			return 640
		}
	}
	funcs["lastC"] = lastC
	funcs["batchSize"] = batchSize

	funcs["contains"] = func(v int, s []int) bool {
		for _, sv := range s {
			if v == sv {
				return true
			}
		}
		return false
	}
	for i := 0; i < len(conf.G1.CRange); i++ {
		lc := lastC(conf.G1.CRange[i])
		if !contains(conf.G1.CRange, lc) && !contains(conf.G1.LastCRange, lc) {
			conf.G1.LastCRange = append(conf.G1.LastCRange, lc)
		}
	}
	for i := 0; i < len(conf.G2.CRange); i++ {
		lc := lastC(conf.G2.CRange[i])
		if !contains(conf.G2.CRange, lc) && !contains(conf.G2.LastCRange, lc) {
			conf.G2.LastCRange = append(conf.G2.LastCRange, lc)
		}
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

func contains(slice []int, v int) bool {
	for i := 0; i < len(slice); i++ {
		if slice[i] == v {
			return true
		}
	}
	return false
}
