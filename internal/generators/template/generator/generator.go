package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/consensys/bavard"

	"github.com/consensys/gurvy/internal/generators/template/gpoint"
	"github.com/consensys/gurvy/internal/generators/template/pairing"
	"github.com/consensys/gurvy/internal/generators/template/tower/fp12"
	"github.com/consensys/gurvy/internal/generators/template/tower/fp2"
	"github.com/consensys/gurvy/internal/generators/template/tower/fp6"

	"github.com/consensys/goff/cmd"
)

// GenerateData data used to generate the templates
type GenerateData struct {

	// common
	Fpackage string
	RootPath string

	// fp, fr moduli
	FpName    string
	FpModulus string
	FrModulus string
	FrName    string

	// fp2
	Fp2Name       string
	Fp2NonResidue string

	// fp6
	Fp6Name       string
	Fp6NonResidue string

	// fp12
	Fp12Name string

	// pairing
	T    string
	TNeg bool

	// gpoint
	PointName    string
	ThirdRootOne string
	Lambda       string
	Size1        string
	Size2        string
}

// PointData to generate g1.go, g2.go
type PointData struct {
	PName     string
	CoordType string
	GroupType string
	Fpackage  string

	// data useful for the "endomorphism trick" to speed up scalar multiplication
	Lambda       string
	ThirdRootOne string
	Size1        string
	Size2        string
}

// GenerateCurve generates tower, curve, pairing
func GenerateCurve(d GenerateData) error {

	if !strings.HasSuffix(d.RootPath, "/") {
		d.RootPath += "/"
	}

	// fp, fr
	{
		if err := cmd.GenerateFF(d.FpName, "Element", d.FpModulus, filepath.Join(d.RootPath, "fp"), false, false); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
		if err := cmd.GenerateFF(d.FrName, "Element", d.FrModulus, filepath.Join(d.RootPath, "fr"), false, false); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
	}

	// fp2
	{
		// generate e2.go
		src := []string{
			fp2.Base,
			fp2.Inline,
			fp2.Mul,
		}
		if err := bavard.Generate(d.RootPath+d.Fp2Name+".go", src, d,
			bavard.Package(d.Fpackage),
			bavard.Apache2("ConsenSys AG", 2020),
			bavard.GeneratedBy("gurvy/internal/generators"),
		); err != nil {
			return err
		}
	}

	// fp6
	{
		// generate e6.go
		src := []string{
			fp6.Base,
			fp2.Inline,
			fp6.Inline,
			fp6.Mul,
		}
		if err := bavard.Generate(d.RootPath+d.Fp6Name+".go", src, d,
			bavard.Package(d.Fpackage),
			bavard.Apache2("ConsenSys AG", 2020),
			bavard.GeneratedBy("gurvy/internal/generators"),
		); err != nil {
			return err
		}
	}

	// fp12
	{
		// generatz e12.go
		src := []string{
			fp12.Base,
			fp2.Inline,
			fp6.Inline,
			fp12.Inline,
			fp12.Mul,
			fp12.Frobenius,
			fp12.Expt,
		}
		if err := bavard.Generate(d.RootPath+d.Fp12Name+".go", src, d,
			bavard.Package(d.Fpackage),
			bavard.Apache2("ConsenSys AG", 2020),
			bavard.GeneratedBy("gurvy/internal/generators"),
		); err != nil {
			return err
		}
	}

	// gpoint
	{
		// g1.go
		point := PointData{
			PName:        d.PointName + "1",
			CoordType:    d.FpName + ".Element",
			GroupType:    d.FrName,
			Fpackage:     d.Fpackage,
			ThirdRootOne: d.ThirdRootOne,
			Lambda:       d.Lambda,
			Size1:        d.Size1,
			Size2:        d.Size2,
		}
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
		if err := bavard.Generate(d.RootPath+strings.ToLower(point.PName)+".go", src, point,
			bavard.Package(d.Fpackage),
			bavard.Apache2("ConsenSys AG", 2020),
			bavard.GeneratedBy("gurvy/internal/generators"),
		); err != nil {
			return err
		}

		// g2.go
		point = PointData{
			PName:     d.PointName + "2",
			CoordType: d.Fp2Name,
			GroupType: d.FrName,
			Fpackage:  d.Fpackage,
		}
		src = []string{
			gpoint.Base,
			gpoint.Add,
			gpoint.AddMixed,
			gpoint.Double,
			gpoint.ScalarMul,
			gpoint.WindowedMultiExp,
			gpoint.MultiExp,
		}
		if err := bavard.Generate(d.RootPath+strings.ToLower(point.PName)+".go", src, point,
			bavard.Package(d.Fpackage),
			bavard.Apache2("ConsenSys AG", 2020),
			bavard.GeneratedBy("gurvy/internal/generators"),
		); err != nil {
			return err
		}

	}

	// pairing
	{
		// generate pairing.go
		src := []string{
			pairing.Pairing,
			pairing.ExtraWork,
			pairing.MulAssign,
		}
		if err := bavard.Generate(d.RootPath+"pairing.go", src, d,
			bavard.Package(d.Fpackage),
			bavard.Apache2("ConsenSys AG", 2020),
			bavard.GeneratedBy("gurvy/internal/generators"),
		); err != nil {
			return err
		}
	}
	return nil
}
