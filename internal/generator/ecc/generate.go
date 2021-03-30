package ecc

import (
	"path/filepath"
	"strings"

	"github.com/consensys/bavard"
	"github.com/consensys/gurvy/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	doc := "provides efficient elliptic curve and pairing implementation for " + conf.Name
	packageName := strings.ReplaceAll(conf.Name, "-", "")

	g1 := pconf{conf, conf.G1}
	g2 := pconf{conf, conf.G2}

	entriesF := []bavard.EntryF{
		{File: filepath.Join(baseDir, "multiexp.go"), TemplateF: []string{"multiexp.go.tmpl"}},
		{File: filepath.Join(baseDir, "multiexp_test.go"), TemplateF: []string{"tests/multiexp.go.tmpl"}},
		{File: filepath.Join(baseDir, "marshal.go"), TemplateF: []string{"marshal.go.tmpl"}, PackageDoc: doc},
		{File: filepath.Join(baseDir, "marshal_test.go"), TemplateF: []string{"tests/marshal.go.tmpl"}},
	}
	if err := bgen.GenerateF(conf, packageName, "./ecc/template", entriesF...); err != nil {
		return err
	}

	// G1
	entriesF = []bavard.EntryF{
		{File: filepath.Join(baseDir, "g1.go"), TemplateF: []string{"point.go.tmpl"}},
		{File: filepath.Join(baseDir, "g1_test.go"), TemplateF: []string{"tests/point.go.tmpl"}},
	}
	if err := bgen.GenerateF(g1, packageName, "./ecc/template", entriesF...); err != nil {
		return err
	}

	// G2
	entriesF = []bavard.EntryF{
		{File: filepath.Join(baseDir, "g2.go"), TemplateF: []string{"point.go.tmpl"}},
		{File: filepath.Join(baseDir, "g2_test.go"), TemplateF: []string{"tests/point.go.tmpl"}},
	}
	return bgen.GenerateF(g2, packageName, "./ecc/template", entriesF...)

}

type pconf struct {
	config.Curve
	config.Point
}
