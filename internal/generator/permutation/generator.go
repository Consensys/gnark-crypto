package permutation

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/permutation/template"
)

func Generate(conf config.Curve, baseDir string, gen *common.Generator) error {
	// permutation data
	conf.Package = "permutation"
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "permutation.go"), Templates: []string{"permutation.go.tmpl"}},
		{File: filepath.Join(baseDir, "permutation_test.go"), Templates: []string{"permutation.test.go.tmpl"}},
	}
	permutationGen := common.NewDefaultGenerator(template.FS)
	return permutationGen.Generate(conf, conf.Package, "", "", entries...)

}
