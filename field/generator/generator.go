package generator

import (
	"fmt"
	"io"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field"
	"github.com/consensys/gnark-crypto/field/asm/amd64"
	"github.com/consensys/gnark-crypto/field/internal/templates/element"
)

// TODO @gbotrel --> pattern for code generation is different than gnark-crypto/internal because a binary like goff can generate
// base field. in Go 1.16, can embed the template in the binary, and use same pattern than gnark-crypto/internal

// GenerateFF will generate go (and .s) files in outputDir for modulus (in base 10)
func GenerateFF(F *field.Field, outputDir string) error {
	// source file templates
	src := []string{
		element.Base,
		element.Reduce,
		element.Exp,
		element.Conv,
		element.MulCIOS,
		element.MulNoCarry,
		element.Sqrt,
		element.Inverse,
	}

	// test file templates
	tst := []string{
		element.MulCIOS,
		element.MulNoCarry,
		element.Reduce,
		element.Test,
	}

	// output files
	eName := strings.ToLower(F.ElementName)

	pathSrc := filepath.Join(outputDir, eName+".go")
	pathSrcArith := filepath.Join(outputDir, "arith.go")
	pathTest := filepath.Join(outputDir, eName+"_test.go")

	// remove old format generated files
	oldFiles := []string{"_mul.go", "_mul_amd64.go",
		"_square.go", "_square_amd64.go", "_ops_decl.go", "_square_amd64.s", "_ops_amd64.go"}
	for _, of := range oldFiles {
		_ = os.Remove(filepath.Join(outputDir, eName+of))
	}

	bavardOpts := []func(*bavard.Bavard) error{
		bavard.Apache2("ConsenSys Software Inc.", 2020),
		bavard.Package(F.PackageName),
		bavard.GeneratedBy("consensys/gnark-crypto"),
		bavard.Funcs(template.FuncMap{"toTitle": strings.Title}),
	}
	bModulus, _ := new(big.Int).SetString(F.Modulus, 10)

	packageDoc := fmt.Sprintf(`contains field arithmetic operations for modulus = 0x%s. 

The API is similar to math/big (big.Int), but the operations are significantly faster (up to 20x for the modular multiplication on amd64, see also https://hackmd.io/@zkteam/modular_multiplication)
 
The modulus is hardcoded in all the operations.

Field elements are represented as an array, and assumed to be in Montgomery form in all methods:
	type %s [%d]uint64

Example API signature
	// Mul z = x * y mod q
	func (z *Element) Mul(x, y *Element) *Element 

and can be used like so:
	var a, b Element
	a.SetUint64(2)
	b.SetString("984896738")

	a.Mul(a, b)

	a.Sub(a, a)
	.Add(a, b)
	.Inv(a)
	
	b.Exp(b, new(big.Int).SetUint64(42))
	b.Neg(b)

Modulus
	0x%s // base 16
	%s // base 10
	`, shorten(bModulus.Text(16)), F.ElementName, F.NbWords, bModulus.Text(16), F.Modulus)

	optsWithPackageDoc := append(bavardOpts, bavard.Package(F.PackageName, packageDoc))

	// generate source file
	if err := bavard.Generate(pathSrc, src, F, optsWithPackageDoc...); err != nil {
		return err
	}
	// generate arithmetics source file
	if err := bavard.Generate(pathSrcArith, []string{element.Arith}, F, bavardOpts...); err != nil {
		return err
	}

	// generate test file
	if err := bavard.Generate(pathTest, tst, F, bavardOpts...); err != nil {
		return err
	}

	// if we generate assembly code
	if F.ASM {
		// generate ops.s
		{
			pathSrc := filepath.Join(outputDir, eName+"_ops_amd64.s")
			fmt.Println("generating", pathSrc)
			f, err := os.Create(pathSrc)
			if err != nil {
				return err
			}

			if err := amd64.Generate(f, F); err != nil {
				_ = f.Close()
				return err
			}
			_ = f.Close()

			// run asmfmt
			// run go fmt on whole directory
			cmd := exec.Command("asmfmt", "-w", pathSrc)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
		}

		{
			pathSrc := filepath.Join(outputDir, eName+"_mul_amd64.s")
			fmt.Println("generating", pathSrc)
			f, err := os.Create(pathSrc)
			if err != nil {
				return err
			}

			_, _ = io.WriteString(f, "// +build !amd64_adx\n")

			if err := amd64.GenerateMul(f, F); err != nil {
				_ = f.Close()
				return err
			}
			_ = f.Close()

			// run asmfmt
			// run go fmt on whole directory
			cmd := exec.Command("asmfmt", "-w", pathSrc)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
		}

		{
			pathSrc := filepath.Join(outputDir, eName+"_mul_adx_amd64.s")
			fmt.Println("generating", pathSrc)
			f, err := os.Create(pathSrc)
			if err != nil {
				return err
			}

			_, _ = io.WriteString(f, "// +build amd64_adx\n")

			if err := amd64.GenerateMulADX(f, F); err != nil {
				_ = f.Close()
				return err
			}
			_ = f.Close()

			// run asmfmt
			// run go fmt on whole directory
			cmd := exec.Command("asmfmt", "-w", pathSrc)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
		}

	}

	{
		// generate ops_amd64.go
		src := []string{
			element.OpsAMD64,
		}
		pathSrc := filepath.Join(outputDir, eName+"_ops_amd64.go")
		if err := bavard.Generate(pathSrc, src, F, bavardOpts...); err != nil {
			return err
		}
	}

	{
		// generate ops.go
		src := []string{
			element.OpsNoAsm,
			element.MulCIOS,
			element.MulNoCarry,
			element.Reduce,
		}
		pathSrc := filepath.Join(outputDir, eName+"_ops_noasm.go")
		bavardOptsCpy := make([]func(*bavard.Bavard) error, len(bavardOpts))
		copy(bavardOptsCpy, bavardOpts)
		if F.ASM {
			bavardOptsCpy = append(bavardOptsCpy, bavard.BuildTag("!amd64"))
		}
		if err := bavard.Generate(pathSrc, src, F, bavardOptsCpy...); err != nil {
			return err
		}
	}

	{
		// generate asm.go and asm_noadx.go
		src := []string{
			element.Asm,
		}
		pathSrc := filepath.Join(outputDir, "asm.go")
		bavardOptsCpy := make([]func(*bavard.Bavard) error, len(bavardOpts))
		copy(bavardOptsCpy, bavardOpts)
		bavardOptsCpy = append(bavardOptsCpy, bavard.BuildTag("!noadx"))
		if err := bavard.Generate(pathSrc, src, F, bavardOptsCpy...); err != nil {
			return err
		}
	}
	{
		// generate asm.go and asm_noadx.go
		src := []string{
			element.AsmNoAdx,
		}
		pathSrc := filepath.Join(outputDir, "asm_noadx.go")
		bavardOptsCpy := make([]func(*bavard.Bavard) error, len(bavardOpts))
		copy(bavardOptsCpy, bavardOpts)
		bavardOptsCpy = append(bavardOptsCpy, bavard.BuildTag("noadx"))
		if err := bavard.Generate(pathSrc, src, F, bavardOptsCpy...); err != nil {
			return err
		}
	}

	// run go fmt on whole directory
	cmd := exec.Command("gofmt", "-s", "-w", outputDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func shorten(input string) string {
	const maxLen = 15
	if len(input) > maxLen {
		return input[:6] + "..." + input[len(input)-6:]
	}
	return input
}
