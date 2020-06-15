package gpoint

import (
	"path/filepath"
	"strings"

	"github.com/consensys/bavard"
	gpoint "github.com/consensys/gurvy/internal/generators/gpoint/templates"
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
		gpoint.Base,
		gpoint.Add,
		gpoint.AddMixed,
		gpoint.Double,
		// gpoint.EndoMul,
		gpoint.ScalarMul,
		gpoint.WindowedMultiExp,
		gpoint.MultiExp,
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
