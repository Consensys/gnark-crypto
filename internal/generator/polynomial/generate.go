package polynomial

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	conf.Package = "polynomial"
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "polynomial.go"), Templates: []string{"polynomial.go.tmpl"}},
	}
	if err := bgen.Generate(conf, conf.Package, "./polynomial/template/", entries...); err != nil {
		return err
	}

	// mock commitment scheme
	conf.Package = "mockcommitment"
	entries = []bavard.Entry{
		{File: filepath.Join(baseDir, "mockcommitment", "doc.go"), Templates: []string{"commitment_mock/doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "mockcommitment", "digest.go"), Templates: []string{"commitment_mock/digest.go.tmpl"}},
		{File: filepath.Join(baseDir, "mockcommitment", "proof.go"), Templates: []string{"commitment_mock/proof.go.tmpl"}},
		{File: filepath.Join(baseDir, "mockcommitment", "proof_single_point.go"), Templates: []string{"commitment_mock/proof.single.point.go.tmpl"}},
		{File: filepath.Join(baseDir, "mockcommitment", "scheme.go"), Templates: []string{"commitment_mock/scheme.go.tmpl"}},
	}

	return bgen.Generate(conf, conf.Package, "./polynomial/template/", entries...)

}
