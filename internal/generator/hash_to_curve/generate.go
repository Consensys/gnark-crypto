package hash_to_curve

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	packageName := strings.ReplaceAll(conf.Name, "-", "")
	htcPackageName := "hash_to_curve"

	// hash To curve
	genHashToCurve := func(point *config.Point, suite config.HashSuite) error {
		if suite == nil { //Nothing to generate. Bypass
			return nil
		}

		err := os.MkdirAll(filepath.Join(baseDir, htcPackageName), 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", filepath.Join(baseDir, "hash_to_curve"), err)
		}

		entries := []bavard.Entry{
			{File: filepath.Join(baseDir, fmt.Sprintf("hash_to_%s.go", point.PointName)), Templates: []string{"root.go.tmpl", "root_sswu.go.tmpl", "root_svdw.go.tmpl"}},
			{File: filepath.Join(baseDir, fmt.Sprintf("hash_to_%s_test.go", point.PointName)), Templates: []string{"tests/hash_to_curve.go.tmpl"}},
		}
		htcEntries := []bavard.Entry{
			{File: filepath.Join(baseDir, htcPackageName, fmt.Sprintf("%s.go", point.PointName)), Templates: []string{"pkg_root.go.tmpl", "pkg_sswu.go.tmpl"}},
		}

		hashConf := suite.GetInfo(conf.Fp, point, conf.Name)

		funcs := make(template.FuncMap)
		funcs["asElement"] = hashConf.Field.Base.WriteElement
		bavardOpts := []func(*bavard.Bavard) error{bavard.Funcs(funcs)}

		return errors.Join(
			bgen.GenerateWithOptions(hashConf, packageName, "./hash_to_curve/template", bavardOpts, entries...),
			bgen.GenerateWithOptions(hashConf, htcPackageName, "./hash_to_curve/template", bavardOpts, htcEntries...),
		)
	}

	return errors.Join(
		genHashToCurve(&conf.G1, conf.HashE1),
		genHashToCurve(&conf.G2, conf.HashE2),
	)
}
