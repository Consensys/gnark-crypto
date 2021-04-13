package edwards

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	conf.Package = "twistededwards"

	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "point.go"), Templates: []string{"pointtwistededwards.go.tmpl"}},
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
	}

	return bgen.Generate(conf, conf.Package, "./edwards/template", entries...)

}
