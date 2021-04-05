package mimc

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/internal/generator/config"
)

func Generate(conf config.Curve, baseDir string, bgen *bavard.BatchGenerator) error {
	doc := "provides MiMC hash function using Miyaguchi–Preneel construction."
	entriesF := []bavard.EntryF{
		{File: filepath.Join(baseDir, "mimc.go"), TemplateF: []string{"mimc.go.tmpl"}, PackageDoc: doc},
	}
	return bgen.GenerateF(conf, "mimc", "./crypto/hash/mimc/template", entriesF...)

}
