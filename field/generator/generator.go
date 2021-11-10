package generator

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field"
	"github.com/consensys/gnark-crypto/field/asm/amd64"
	"github.com/consensys/gnark-crypto/field/internal/templates/element"
	"github.com/mmcloughlin/addchain/acc/ir"
	"github.com/mmcloughlin/addchain/acc/printer"
)

// TODO @gbotrel --> pattern for code generation is different than gnark-crypto/internal because a binary like goff can generate
// base field. in Go 1.16, can embed the template in the binary, and use same pattern than gnark-crypto/internal

// GenerateFF will generate go (and .s) files in outputDir for modulus (in base 10)
//
// Example usage
//
// 	fp, _ = field.NewField("fp", "Element", fpModulus")
// 	generator.GenerateFF(fp, filepath.Join(baseDir, "fp"))
func GenerateFF(F *field.Field, outputDir string) error {
	// source file templates
	sourceFiles := []string{
		element.Base,
		element.Reduce,
		element.Exp,
		element.Conv,
		element.MulCIOS,
		element.MulNoCarry,
		element.Sqrt,
		element.Inverse,
		element.ExpBy,
	}

	// test file templates
	testFiles := []string{
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
	pathFuzz := filepath.Join(outputDir, eName+"_fuzz.go")

	// remove old format generated files
	oldFiles := []string{"_mul.go", "_mul_amd64.go",
		"_square.go", "_square_amd64.go", "_ops_decl.go", "_square_amd64.s", "_ops_amd64.go"}
	for _, of := range oldFiles {
		_ = os.Remove(filepath.Join(outputDir, eName+of))
	}

	funcs := template.FuncMap{}
	for _, f := range Functions {
		funcs[f.Name] = f.Func
	}
	funcs["toTitle"] = strings.Title
	funcs["shorten"] = shorten

	bavardOpts := []func(*bavard.Bavard) error{
		bavard.Apache2("ConsenSys Software Inc.", 2020),
		bavard.Package(F.PackageName),
		bavard.GeneratedBy("consensys/gnark-crypto"),
		bavard.Funcs(funcs),
	}

	// generate source file
	if err := bavard.GenerateFromString(pathSrc, sourceFiles, F, bavardOpts...); err != nil {
		return err
	}
	// generate arithmetics source file
	if err := bavard.GenerateFromString(pathSrcArith, []string{element.Arith}, F, bavardOpts...); err != nil {
		return err
	}

	// generate fuzz file
	bopts := make([]func(*bavard.Bavard) error, len(bavardOpts))
	copy(bopts, bavardOpts)
	bopts = append(bopts, bavard.BuildTag("gofuzz"))
	if err := bavard.GenerateFromString(pathFuzz, []string{element.Fuzz}, F, bopts...); err != nil {
		return err
	}

	// generate test file
	if err := bavard.GenerateFromString(pathTest, testFiles, F, bavardOpts...); err != nil {
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
		if err := bavard.GenerateFromString(pathSrc, src, F, bavardOpts...); err != nil {
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
		if err := bavard.GenerateFromString(pathSrc, src, F, bavardOptsCpy...); err != nil {
			return err
		}
	}

	{
		// generate doc.go
		src := []string{
			element.Doc,
		}
		pathSrc := filepath.Join(outputDir, "doc.go")
		if err := bavard.GenerateFromString(pathSrc, src, F, bavardOpts...); err != nil {
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
		if err := bavard.GenerateFromString(pathSrc, src, F, bavardOptsCpy...); err != nil {
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
		if err := bavard.GenerateFromString(pathSrc, src, F, bavardOptsCpy...); err != nil {
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

// TODO @gbotrel add copyright + licencing

// Function is a function provided to templates.
type Function struct {
	Name        string
	Description string
	Func        interface{}
}

// Signature returns the function signature.
func (f *Function) Signature() string {
	return reflect.ValueOf(f.Func).Type().String()
}

// Functions is the list of functions provided to templates.
var Functions = []*Function{
	{
		Name:        "add_",
		Description: "If the input operation is an `ir.Add` then return it, otherwise return `nil`",
		Func: func(op ir.Op) ir.Op {
			if a, ok := op.(ir.Add); ok {
				return a
			}
			return nil
		},
	},
	{
		Name:        "double_",
		Description: "If the input operation is an `ir.Double` then return it, otherwise return `nil`",
		Func: func(op ir.Op) ir.Op {
			if d, ok := op.(ir.Double); ok {
				return d
			}
			return nil
		},
	},
	{
		Name:        "shift_",
		Description: "If the input operation is an `ir.Shift` then return it, otherwise return `nil`",
		Func: func(op ir.Op) ir.Op {
			if s, ok := op.(ir.Shift); ok {
				return s
			}
			return nil
		},
	},
	{
		Name:        "inc_",
		Description: "Increment an integer",
		Func:        func(n int) int { return n + 1 },
	},
	{
		Name:        "format_",
		Description: "Formats an addition chain script (`*ast.Chain`) as a string",
		Func:        printer.String,
	},
	{
		Name:        "split_",
		Description: "Calls `strings.Split`",
		Func:        strings.Split,
	},
	{
		Name:        "join_",
		Description: "Calls `strings.Join`",
		Func:        strings.Join,
	},
	{
		Name:        "lines_",
		Description: "Split input string into lines",
		Func: func(s string) []string {
			var lines []string
			scanner := bufio.NewScanner(strings.NewReader(s))
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			return lines
		},
	},
	{
		Name:        "ptr_",
		Description: "adds & if it's a value",
		Func: func(s *ir.Operand) string {
			if s.String() == "x" || s.String() == "z" || s.String() == "y" {
				return ""
			}
			return "&"
		},
	},
	{
		Name: "last_",
		Func: func(x int, a interface{}) bool {
			return x == reflect.ValueOf(a).Len()-1
		},
	},
}
