package pedersen

import (
	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"path/filepath"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	if conf.Equal(config.SECP256K1) || conf.Equal(config.SECQ256K1) {
		return nil
	}

	// pedersen commitment scheme
	conf.Package = "pedersen"
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "pedersen.go"), Templates: []string{"pedersen.go.tmpl"}},
		{File: filepath.Join(baseDir, "pedersen_test.go"), Templates: []string{"pedersen.test.go.tmpl"}},
	}
	return bgen.Generate(conf, conf.Package, "./pedersen/template/", entries...)

}
