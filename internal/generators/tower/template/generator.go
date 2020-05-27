package tower

import (
	"path/filepath"
	"strings"

	"github.com/consensys/bavard"
	"github.com/consensys/goff/cmd"

	"github.com/consensys/gurvy/internal/generators/tower/template/fp12"
	"github.com/consensys/gurvy/internal/generators/tower/template/fp2"
	"github.com/consensys/gurvy/internal/generators/tower/template/fp6"
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

	// data needed in the template, always set to constants
	Fp2Name  string // TODO this name cannot change; remove it
	Fp6Name  string // TODO this name cannot change; remove it
	Fp12Name string // TODO this name cannot change; remove it
}

// GenerateTower generates pairing
func GenerateTower(d GenerateData) error {

	rootPath := "../../../" + d.Fpackage + "/"

	// fp, fr
	{
		fpPath := filepath.Join(rootPath, "fp")
		// if err := os.MkdirAll(fpPath, 0700); err != nil {
		// 	return err
		// }
		if err := cmd.GenerateFF(FpName, "Element", d.FpModulus, fpPath, false, false); err != nil {
			return err
		}
		frPath := filepath.Join(rootPath, "fr")
		// if err := os.MkdirAll(frPath, 0700); err != nil {
		// 	return err
		// }
		if err := cmd.GenerateFF(FrName, "Element", d.FrModulus, frPath, false, false); err != nil {
			return err
		}
	}

	// fp2
	{
		src := []string{
			fp2.Base,
			fp2.Inline,
			fp2.Mul,
		}
		if err := bavard.Generate(rootPath+strings.ToLower(Fp2Name)+".go", src, d,
			bavard.Package(d.Fpackage),
			bavard.Apache2("ConsenSys AG", 2020),
			bavard.GeneratedBy("gurvy/internal/generators"),
		); err != nil {
			return err
		}
	}

	// fp6
	{
		src := []string{
			fp6.Base,
			fp2.Inline,
			fp6.Inline,
			fp6.Mul,
		}
		if err := bavard.Generate(rootPath+strings.ToLower(Fp6Name)+".go", src, d,
			bavard.Package(d.Fpackage),
			bavard.Apache2("ConsenSys AG", 2020),
			bavard.GeneratedBy("gurvy/internal/generators"),
		); err != nil {
			return err
		}
	}

	// fp12
	if d.MakeFp12 {
		src := []string{
			fp12.Base,
			fp2.Inline,
			fp6.Inline,
			fp12.Inline,
			fp12.Mul,
			// fp12.Frobenius,
			// fp12.Expt,
		}
		if err := bavard.Generate(rootPath+strings.ToLower(Fp12Name)+".go", src, d,
			bavard.Package(d.Fpackage),
			bavard.Apache2("ConsenSys AG", 2020),
			bavard.GeneratedBy("gurvy/internal/generators"),
		); err != nil {
			return err
		}
	}

	return nil
}
