package mimc

import (
	"os"
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	if conf.Equal(config.SECP256K1) {
		return nil
	}

	conf.Package = "mimc"
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "mimc.go"), Templates: []string{"mimc.go.tmpl"}},
	}
	os.Remove(filepath.Join(baseDir, "utils.go"))
	os.Remove(filepath.Join(baseDir, "utils_test.go"))

	return bgen.Generate(conf, conf.Package, "./crypto/hash/mimc/template", entries...)

}
