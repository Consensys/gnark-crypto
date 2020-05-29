package primefields

import (
	"path/filepath"

	"github.com/consensys/goff/cmd"
)

// Data data used to generate the templates
type Data struct {
	Fpackage  string
	FpModulus string
	FrModulus string
	FpName    string // TODO other templates assume "fp"
	FrName    string // TODO other templates assume "fr"
}

// Generate generates fp, fr prime fields
func Generate(d Data, outputDir string) error {

	// if !strings.HasSuffix(outputDir, "/") {
	// 	d.RootPath += "/"
	// }
	// rootPath := outputDir + d.Fpackage + "/"
	rootPath := filepath.Join(outputDir, d.Fpackage)

	// fp
	fpPath := filepath.Join(rootPath, d.FpName)
	if err := cmd.GenerateFF(d.FpName, "Element", d.FpModulus, fpPath, false, false); err != nil {
		return err
	}

	// fr
	frPath := filepath.Join(rootPath, d.FrName)
	if err := cmd.GenerateFF(d.FrName, "Element", d.FrModulus, frPath, false, false); err != nil {
		return err
	}

	// tower template generator
	// {
	// 	src := []string{
	// 		TwoInv,
	// 	}
	// 	if err := bavard.Generate("../tower/template/twoinv.go", src, d,
	// 		bavard.Package("tower"),
	// 		// bavard.Apache2("ConsenSys AG", 2020),
	// 		bavard.GeneratedBy("gurvy/internal/generators"),
	// 	); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}
