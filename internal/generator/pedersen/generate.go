package pedersen

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/pedersen/template"
)

func Generate(conf config.Curve, baseDir string, gen *common.Generator) error {
	// pedersen commitment scheme
	conf.Package = "pedersen"
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "pedersen.go"), Templates: []string{"pedersen.go.tmpl"}},
		{File: filepath.Join(baseDir, "pedersen_test.go"), Templates: []string{"pedersen.test.go.tmpl"}},
		{File: filepath.Join(baseDir, "example_test.go"), Templates: []string{"example_test.go.tmpl"}},
	}
	pedersenGen := common.NewDefaultGenerator(template.FS)
	return pedersenGen.Generate(conf, conf.Package, "", "", entries...)

}
