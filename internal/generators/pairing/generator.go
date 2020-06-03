package pairing

import (
	"path/filepath"

	"github.com/consensys/bavard"
	pairing "github.com/consensys/gurvy/internal/generators/pairing/templates"
)

// Data data used to generate the templates
type Data struct {
	Fpackage string
	// FpModulus string
	// FrModulus string
	// Fp2NonResidue string
	Fp6NonResidue   string
	EmbeddingDegree int
	T               string
	TNeg            bool

	// data needed in the template, always set to constants
	Fp2Name  string // TODO this name cannot change; remove it
	Fp6Name  string // TODO this name cannot change; remove it
	Fp12Name string // TODO this name cannot change; remove it

	// these members are computed as needed
	Frobenius [][]fp2Template // constants used Frobenius
}

// Generate generates pairing
func Generate(d Data, outputDir string) error {

	rootPath := filepath.Join(outputDir, d.Fpackage)

	d.InitFrobenius()

	// pairing.go
	{
		src := []string{
			pairing.Pairing,
			pairing.ExtraWork,
			pairing.MulAssign,
			pairing.Expt,
		}
		if err := bavard.Generate(filepath.Join(rootPath, "pairing.go"), src, d,
			bavard.Package(d.Fpackage),
			bavard.Apache2("ConsenSys AG", 2020),
			bavard.GeneratedBy("gurvy/internal/generators"),
		); err != nil {
			return err
		}
	}

	// frobenius.go
	{
		src := []string{
			pairing.Frobenius,
		}
		if err := bavard.Generate(filepath.Join(rootPath, "frobenius.go"), src, d,
			bavard.Package(d.Fpackage),
			bavard.Apache2("ConsenSys AG", 2020),
			bavard.GeneratedBy("gurvy/internal/generators"),
		); err != nil {
			return err
		}
	}

	return nil
}

const ImportsTemplate = `
import (
	"github.com/consensys/gurvy/{{$.Fpackage}}"
	"github.com/consensys/gurvy/{{$.Fpackage}}/fp"
	"math/big"
)

type fp2 = {{$.Fpackage}}.{{$.Fp2Name}}

func primeModulus() *big.Int {
	return fp.ElementModulus()
}
`
