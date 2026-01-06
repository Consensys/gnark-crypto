package poseidon2

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/common"
	"github.com/consensys/gnark-crypto/internal/generator/config"
	"github.com/consensys/gnark-crypto/internal/generator/crypto/hash/poseidon2/template"
)

func Generate(conf config.Curve, baseDir string, gen *common.Generator) error {
	conf.Package = "poseidon2"
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "hash.go"), Templates: []string{"hash.go.tmpl"}},
		{File: filepath.Join(baseDir, "poseidon2.go"), Templates: []string{"poseidon2.go.tmpl"}},
		{File: filepath.Join(baseDir, "poseidon2_test.go"), Templates: []string{"poseidon2.test.go.tmpl"}},
	}

	poseidonGen := common.NewDefaultGenerator(template.FS)
	return poseidonGen.Generate(conf, conf.Package, "", "", entries...)

}
