package polynomial

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	doc := "provides polynomial methods and commitment schemes."
	entries := []bavard.EntryF{
		{File: filepath.Join(baseDir, "polynomial.go"), TemplateF: []string{"polynomial.go.tmpl"}, PackageDoc: doc},
	}
	if err := bgen.GenerateF(conf, "polynomial", "./polynomial/template/", entries...); err != nil {
		return err
	}

	// mock commitment scheme
	doc = "provides a mock polynomial commitment scheme, for test and development purposes."
	entries = []bavard.EntryF{
		{File: filepath.Join(baseDir, "mockcommitment", "digest.go"), TemplateF: []string{"commitment_mock/digest.go.tmpl"}},
		{File: filepath.Join(baseDir, "mockcommitment", "proof.go"), TemplateF: []string{"commitment_mock/proof.go.tmpl"}},
		{File: filepath.Join(baseDir, "mockcommitment", "proof_single_point.go"), TemplateF: []string{"commitment_mock/proof.single.point.go.tmpl"}},
		{File: filepath.Join(baseDir, "mockcommitment", "scheme.go"), TemplateF: []string{"commitment_mock/scheme.go.tmpl"}, PackageDoc: doc},
	}

	return bgen.GenerateF(conf, "mockcommitment", "./polynomial/template/", entries...)

}
