package gpoint

import (
	"path/filepath"
	"strings"

	"github.com/consensys/bavard"
)

// Data data used to generate the templates
type Data struct {
	Fpackage  string
	PName     string
	CoordType string
	GroupType string

	// data useful for the "endomorphism trick" to speed up scalar multiplication
	Lambda       string
	ThirdRootOne string
	Size1        string
	Size2        string
}

// Generate generates pairing
func Generate(d Data, outputDir string) error {

	rootPath := filepath.Join(outputDir, d.Fpackage)
	src := []string{
		Base,
		Add,
		AddMixed,
		Double,
		// EndoMul,
		ScalarMul,
		WindowedMultiExp,
		MultiExp,
	}
	if err := bavard.Generate(filepath.Join(rootPath, strings.ToLower(d.PName)+".go"), src, d,
		bavard.Package(d.Fpackage),
		bavard.Apache2("ConsenSys AG", 2020),
		bavard.GeneratedBy("gurvy/internal/generators"),
	); err != nil {
		return err
	}

	return nil
}
