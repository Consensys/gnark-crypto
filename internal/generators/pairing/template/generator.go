package pairing

import (
	"github.com/consensys/bavard"
)

const FpName = "fp"
const FrName = "fr"
const Fp2Name = "E2"
const Fp6Name = "E6"
const Fp12Name = "E12"

// GenerateData data used to generate the templates
type GenerateData struct {

	// common
	Fpackage string
	// RootPath string // TODO deduce this from Fpackage; remove it

	// fp, fr moduli
	// FpName    string // TODO this name cannot change; remove it
	FpModulus string
	FrModulus string
	// FrName    string // TODO this name cannot change; remove it

	// fp2
	Fp2NonResidue string

	// fp6
	Fp6NonResidue string

	MakeFp12 bool // TODO need a better way to specify which fields to make
	// fp12
	// Fp12Name string // TODO this name cannot change; remove it

	T    string
	TNeg bool

	EmbeddingDegree int

	// data needed in the template, always set to constants
	Fp2Name  string // TODO this name cannot change; remove it
	Fp6Name  string // TODO this name cannot change; remove it
	Fp12Name string // TODO this name cannot change; remove it

	// these members are computed as needed
	Frobenius [][]fp2Template // constants used Frobenius
}

// GeneratePairing generates pairing
func GeneratePairing(d GenerateData) error {

	rootPath := "../../../" + d.Fpackage + "/"

	d.InitFrobenius()

	// pairing.go
	{
		src := []string{
			Pairing,
			ExtraWork,
			MulAssign,
			Expt,
		}
		if err := bavard.Generate(rootPath+"pairing.go", src, d,
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
			Frobenius,
		}
		if err := bavard.Generate(rootPath+"frobenius.go", src, d,
			bavard.Package(d.Fpackage),
			bavard.Apache2("ConsenSys AG", 2020),
			bavard.GeneratedBy("gurvy/internal/generators"),
		); err != nil {
			return err
		}
	}

	return nil
}
