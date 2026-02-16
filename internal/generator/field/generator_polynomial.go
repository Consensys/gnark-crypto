package field

import (
	"path/filepath"

	"github.com/consensys/gnark-crypto/internal/generator/field/config"
	"github.com/consensys/gnark-crypto/internal/generator/polynomial"
)

func generatePolynomial(F *config.Field, outputDir string) error {

	fieldImportPath, err := getImportPath(outputDir)
	if err != nil {
		return err
	}

	fieldInfo := config.FieldDependency{
		FieldPackagePath: fieldImportPath,
		FieldPackageName: F.PackageName,
		ElementType:      F.PackageName + ".Element",
	}

	return polynomial.Generate(fieldInfo, filepath.Join(outputDir, "polynomial"), true, nil)
}
