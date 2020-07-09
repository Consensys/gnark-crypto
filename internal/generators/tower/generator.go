package tower

import (
	"path/filepath"
	"strings"

	"github.com/consensys/bavard"

	"github.com/consensys/gurvy/internal/generators/tower/templates/fp12"
	"github.com/consensys/gurvy/internal/generators/tower/templates/fp2"
	"github.com/consensys/gurvy/internal/generators/tower/templates/fp6"
)

// Data data used to generate the templates
type Data struct {
	Fpackage        string
	Fp2NonResidue   string
	Fp6NonResidue   string
	EmbeddingDegree int

	// data needed in the template, always set to constants
	Fp2Name  string
	Fp6Name  string
	Fp12Name string
}

// Generate generates pairing
func Generate(d Data, outputDir string) error {

	rootPath := filepath.Join(outputDir, d.Fpackage)

	// fp2
	if d.EmbeddingDegree >= 2 {
		src := []string{
			fp2.Base,
			fp2.Inline,
			fp2.Mul,
		}
		if err := bavard.Generate(filepath.Join(rootPath, strings.ToLower(d.Fp2Name)+".go"), src, d,
			bavard.Package(d.Fpackage),
			bavard.Apache2("ConsenSys AG", 2020),
			bavard.GeneratedBy("gurvy/internal/generators"),
		); err != nil {
			return err
		}
	}

	// fp6
	if d.EmbeddingDegree >= 6 {
		src := []string{
			fp6.Base,
			fp2.Inline,
			fp6.Inline,
			fp6.Mul,
		}
		if err := bavard.Generate(filepath.Join(rootPath, strings.ToLower(d.Fp6Name)+".go"), src, d,
			bavard.Package(d.Fpackage),
			bavard.Apache2("ConsenSys AG", 2020),
			bavard.GeneratedBy("gurvy/internal/generators"),
		); err != nil {
			return err
		}
	}

	// fp12
	if d.EmbeddingDegree >= 12 {
		src := []string{
			fp12.Base,
			fp2.Inline,
			fp6.Inline,
			fp12.Inline,
			fp12.Mul,
		}
		if err := bavard.Generate(filepath.Join(rootPath, strings.ToLower(d.Fp12Name)+".go"), src, d,
			bavard.Package(d.Fpackage),
			bavard.Apache2("ConsenSys AG", 2020),
			bavard.GeneratedBy("gurvy/internal/generators"),
		); err != nil {
			return err
		}
	}

	return nil
}
