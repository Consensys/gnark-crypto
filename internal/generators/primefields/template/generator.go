package primefields

import (
	"path/filepath"

	"github.com/consensys/bavard"
	"github.com/consensys/goff/cmd"
)

const FpName = "fp"
const FrName = "fr"

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
}

// GeneratePrimeFields generates fp, fr prime fields
func GeneratePrimeFields(d GenerateData) error {

	rootPath := "../../../" + d.Fpackage + "/"

	// fp, fr
	{
		fpPath := filepath.Join(rootPath, FpName)
		// if err := os.MkdirAll(fpPath, 0700); err != nil {
		// 	return err
		// }
		if err := cmd.GenerateFF(FpName, "Element", d.FpModulus, fpPath, false, false); err != nil {
			return err
		}
		frPath := filepath.Join(rootPath, FrName)
		// if err := os.MkdirAll(frPath, 0700); err != nil {
		// 	return err
		// }
		if err := cmd.GenerateFF(FrName, "Element", d.FrModulus, frPath, false, false); err != nil {
			return err
		}
	}

	// tower template generator
	{
		src := []string{
			TwoInv,
		}
		if err := bavard.Generate("../tower/template/twoinv.go", src, d,
			bavard.Package("tower"),
			// bavard.Apache2("ConsenSys AG", 2020),
			bavard.GeneratedBy("gurvy/internal/generators"),
		); err != nil {
			return err
		}
	}

	return nil
}
