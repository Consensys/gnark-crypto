package hash_to_field

import (
	"path/filepath"

	"github.com/consensys/gnark-crypto/field/generator/common"
	"github.com/consensys/gnark-crypto/field/generator/config"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/hash_to_field/template"
)

func Generate(conf config.FieldDependency, baseDir string, gen *common.Generator) error {
	entries := []bavard.Entry{
		{File: filepath.Join(baseDir, "doc.go"), Templates: []string{"doc.go.tmpl"}},
		{File: filepath.Join(baseDir, "hash_to_field.go"), Templates: []string{"hash_to_field.go.tmpl"}},
		{File: filepath.Join(baseDir, "hash_to_field_test.go"), Templates: []string{"hash_to_field_test.go.tmpl"}},
	}

	h2fGen := common.NewGenerator(template.FS, "Consensys Software Inc.", 2020, "consensys/gnark-crypto")
	return h2fGen.Generate(conf, "hash_to_field", "", "", entries...)
}
