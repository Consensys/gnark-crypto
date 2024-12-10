package generator

import (
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field/generator/asm/amd64"
	"github.com/consensys/gnark-crypto/field/generator/asm/arm64"
	"github.com/consensys/gnark-crypto/field/generator/config"
	"github.com/consensys/gnark-crypto/field/generator/internal/addchain"
	"github.com/consensys/gnark-crypto/field/generator/internal/templates/element"
	"golang.org/x/sync/errgroup"
)

// GenerateFF will generate go (and .s) files in outputDir for modulus (in base 10)
//
// Example usage
//
//	fp, _ = config.NewField("fp", "Element", fpModulus")
//	generator.GenerateFF(fp, filepath.Join(baseDir, "fp"))
func GenerateFF(F *config.FieldConfig, outputDir, asmDirBuildPath, asmDirIncludePath string) error {
	// source file templates
	sourceFiles := []string{
		element.Base,
		element.Reduce,
		element.Exp,
		element.Conv,
		element.MulDoc,
		element.MulCIOS,
		element.MulNoCarry,
		element.Sqrt,
		element.Inverse,
		element.BigNum,
	}

	// test file templates
	testFiles := []string{
		element.MulCIOS,
		element.MulNoCarry,
		element.Reduce,
		element.Test,
		element.InverseTests,
	}
	funcs := template.FuncMap{}
	if F.UseAddChain {
		for _, f := range addchain.Functions {
			funcs[f.Name] = f.Func
		}
	}

	os.Remove(filepath.Join(outputDir, "vector_arm64.go"))
	os.Remove(filepath.Join(outputDir, "exp.go"))

	funcs["shorten"] = shorten
	funcs["ltu64"] = func(a, b uint64) bool {
		return a < b
	}

	generate := func(suffix string, templates []string, opts ...option) func() error {
		opt := generateOptions(opts...)
		if opt.skip {
			return func() error { return nil }
		}
		return func() error {
			bavardOpts := []func(*bavard.Bavard) error{
				bavard.Apache2("Consensys Software Inc.", 2020),
				bavard.GeneratedBy("consensys/gnark-crypto"),
				bavard.Funcs(funcs),
			}
			if !strings.HasSuffix(suffix, ".s") {
				bavardOpts = append(bavardOpts, bavard.Package(F.PackageName))
			}
			if opt.buildTag != "" {
				bavardOpts = append(bavardOpts, bavard.BuildTag(opt.buildTag))
			}
			if suffix == ".go" {
				suffix = filepath.Join(outputDir, suffix)
			} else {
				suffix = filepath.Join(outputDir, suffix)
			}

			tmplData := any(F)
			if opt.tmplData != nil {
				tmplData = opt.tmplData
			}

			return bavard.GenerateFromString(suffix, templates, tmplData, bavardOpts...)
		}
	}

	// generate asm files;
	// couple of cases;
	// 1. we generate arm64 and amd64
	// 2. we generate only amd64
	// 3. we generate only purego

	// sanity check
	if (F.GenerateOpsARM64 && !F.GenerateOpsAMD64) ||
		(F.GenerateVectorOpsAMD64 && !F.GenerateOpsAMD64) ||
		(F.GenerateVectorOpsARM64 && !F.GenerateOpsARM64) {
		panic("not implemented.")
	}

	// get hash of the common asm files to force compiler to recompile in case of changes.
	var amd64d, arm64d ASMWrapperData
	var err error

	if F.GenerateOpsAMD64 {
		amd64d, err = hashAndInclude(asmDirBuildPath, asmDirIncludePath, amd64.ElementASMFileName(F.NbWords, F.NbBits))
		if err != nil {
			return err
		}
	}

	if F.GenerateOpsARM64 {
		arm64d, err = hashAndInclude(asmDirBuildPath, asmDirIncludePath, arm64.ElementASMFileName(F.NbWords, F.NbBits))
		if err != nil {
			return err
		}
	}

	// purego files have no build tags if we don't generate asm
	pureGoBuildTag := "purego || (!amd64 && !arm64)"
	if !F.GenerateOpsAMD64 && !F.GenerateOpsARM64 {
		pureGoBuildTag = ""
	} else if !F.GenerateOpsARM64 {
		pureGoBuildTag = "purego || (!amd64)"
	}

	pureGoVectorBuildTag := "purego || (!amd64 && !arm64)"
	if !F.GenerateVectorOpsAMD64 && !F.GenerateVectorOpsARM64 {
		pureGoVectorBuildTag = ""
	} else if !F.GenerateVectorOpsARM64 {
		pureGoVectorBuildTag = "purego || (!amd64)"
	}

	if F.F31 {
		pureGoBuildTag = "" // always generate pure go for F31
		pureGoVectorBuildTag = "purego || (!amd64)"
	}

	var g errgroup.Group

	g.Go(generate("element.go", sourceFiles))
	g.Go(generate("doc.go", []string{element.Doc}))
	g.Go(generate("vector.go", []string{element.Vector}))
	g.Go(generate("arith.go", []string{element.Arith}, Only(!F.F31)))
	g.Go(generate("element_test.go", testFiles))
	g.Go(generate("vector_test.go", []string{element.TestVector}))

	g.Go(generate("element_amd64.s", []string{element.IncludeASM}, Only(F.GenerateOpsAMD64), WithBuildTag("!purego"), WithData(amd64d)))
	g.Go(generate("element_arm64.s", []string{element.IncludeASM}, Only(F.GenerateOpsARM64), WithBuildTag("!purego"), WithData(arm64d)))

	g.Go(generate("element_amd64.go", []string{element.OpsAMD64, element.MulDoc}, Only(F.GenerateOpsAMD64 && !F.F31), WithBuildTag("!purego")))
	g.Go(generate("element_arm64.go", []string{element.OpsARM64, element.MulNoCarry, element.Reduce}, Only(F.GenerateOpsARM64), WithBuildTag("!purego")))

	g.Go(generate("element_purego.go", []string{element.OpsNoAsm, element.MulCIOS, element.MulNoCarry, element.Reduce, element.MulDoc}, WithBuildTag(pureGoBuildTag)))

	g.Go(generate("vector_amd64.go", []string{element.VectorOpsAmd64}, Only(F.GenerateVectorOpsAMD64 && !F.F31), WithBuildTag("!purego")))
	g.Go(generate("vector_amd64.go", []string{element.VectorOpsAmd64F31}, Only(F.GenerateVectorOpsAMD64 && F.F31), WithBuildTag("!purego")))
	g.Go(generate("vector_arm64.go", []string{element.VectorOpsArm64}, Only(F.GenerateVectorOpsARM64), WithBuildTag("!purego")))

	g.Go(generate("vector_purego.go", []string{element.VectorOpsPureGo}, WithBuildTag(pureGoVectorBuildTag)))

	g.Go(generate("asm_adx.go", []string{element.Asm}, Only(F.GenerateOpsAMD64 && !F.F31), WithBuildTag("!noadx")))
	g.Go(generate("asm_noadx.go", []string{element.AsmNoAdx}, Only(F.GenerateOpsAMD64 && !F.F31), WithBuildTag("noadx")))
	g.Go(generate("asm_avx.go", []string{element.Avx}, Only(F.GenerateVectorOpsAMD64), WithBuildTag("!noavx")))
	g.Go(generate("asm_noavx.go", []string{element.NoAvx}, Only(F.GenerateVectorOpsAMD64), WithBuildTag("noavx")))

	if F.UseAddChain {
		g.Go(generate("element_exp.go", []string{element.FixedExp}))
	}

	if err := g.Wait(); err != nil {
		return err
	}

	{
		// run go fmt on whole directory
		cmd := exec.Command("gofmt", "-s", "-w", outputDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	{
		// run asmfmt on whole directory
		cmd := exec.Command("asmfmt", "-w", outputDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

type ASMWrapperData struct {
	IncludePath string
	Hash        string
}

func hashAndInclude(asmDirBuildPath, asmDirIncludePath, fileName string) (data ASMWrapperData, err error) {
	// we hash the file content and include the hash in comment of the generated file
	// to force the Go compiler to recompile the file if the content has changed
	fData, err := os.ReadFile(filepath.Join(asmDirBuildPath, fileName))
	if err != nil {
		return ASMWrapperData{}, err
	}
	// hash the file using FNV
	hasher := fnv.New64()
	hasher.Write(fData)
	hash64 := hasher.Sum64()

	hash := fmt.Sprintf("%d", hash64)
	includePath := filepath.Join(asmDirIncludePath, fileName)
	// on windows, we replace the "\" by "/"
	if filepath.Separator == '\\' {
		includePath = strings.ReplaceAll(includePath, "\\", "/")
	}

	return ASMWrapperData{
		IncludePath: includePath,
		Hash:        hash,
	}, nil

}

func shorten(input string) string {
	const maxLen = 15
	if len(input) > maxLen {
		return input[:6] + "..." + input[len(input)-6:]
	}
	return input
}

func GenerateARM64(nbWords, nbBits int, asmDir string, hasVector bool) error {
	os.MkdirAll(asmDir, 0755)
	pathSrc := filepath.Join(asmDir, arm64.ElementASMFileName(nbWords, nbBits))

	fmt.Println("generating", pathSrc)
	f, err := os.Create(pathSrc)
	if err != nil {
		return err
	}

	if err := arm64.GenerateCommonASM(f, nbWords, nbBits, hasVector); err != nil {
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

	return nil
}

func GenerateAMD64(nbWords, nbBits int, asmDir string, hasVector bool) error {
	os.MkdirAll(asmDir, 0755)
	pathSrc := filepath.Join(asmDir, amd64.ElementASMFileName(nbWords, nbBits))

	fmt.Println("generating", pathSrc)
	f, err := os.Create(pathSrc)
	if err != nil {
		return err
	}

	if err := amd64.GenerateCommonASM(f, nbWords, nbBits, hasVector); err != nil {
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

	return nil
}

type option func(*generateConfig)
type generateConfig struct {
	buildTag string
	skip     bool
	tmplData any
}

func WithBuildTag(buildTag string) option {
	return func(opt *generateConfig) {
		opt.buildTag = buildTag
	}
}

func Only(condition bool) option {
	return func(opt *generateConfig) {
		opt.skip = !condition
	}
}

func WithData(data any) option {
	return func(opt *generateConfig) {
		opt.tmplData = data
	}
}

// default options
func generateOptions(opts ...option) generateConfig {
	// apply options
	opt := generateConfig{}
	for _, option := range opts {
		option(&opt)
	}
	return opt
}
