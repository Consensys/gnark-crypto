package tower

import (
	"strings"

	"github.com/consensys/bavard"

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
	Fpackage string

	FpModulus string // TODO is this still needed?
	FrModulus string // TODO is this still needed?

	Fp2NonResidue string
	Fp6NonResidue string

	MakeFp12 bool // TODO need a better way to specify which fields to make

	// data needed in the template, always set to constants
	Fp2Name  string // TODO this name cannot change; remove it
	Fp6Name  string // TODO this name cannot change; remove it
	Fp12Name string // TODO this name cannot change; remove it

	// these members are computed as needed
	TwoInv []uint64 // fp.Element, used only when Fp2NonResidue==-1 and Fp6NonResidue==(1,1). TODO there must be a better way to do this.
}

// GenerateTower generates pairing
func GenerateTower(d GenerateData) error {

	rootPath := "../../../" + d.Fpackage + "/"

	// inverse of 2 in fp is used by some curves
	if d.Fp2NonResidue == "-1" && d.Fp6NonResidue == "1,1" {
		d.InitTwoInv()
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
