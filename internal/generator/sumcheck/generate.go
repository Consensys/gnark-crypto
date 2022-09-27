package sumcheck

import (
	"github.com/consensys/bavard"
	"path/filepath"
)

// TODO: Put somewhere else, as this is used in the gkr package as well
type FieldInfo struct {
	PackagePath string
	ElementType string
}

func Generate(conf FieldInfo, baseDir string, bgen *bavard.BatchGenerator) error {
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "sumcheck.go"), Templates: []string{"sumcheck.go.tmpl"}},
		{File: filepath.Join(baseDir, "sumcheck_test.go"), Templates: []string{"sumcheck.test.go.tmpl"}},
	}
	return bgen.Generate(conf, "sumcheck", "./sumcheck/template/", entries...)
}
