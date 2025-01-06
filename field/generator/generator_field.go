package generator

import (
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

func generateField(F *config.Field, outputDir, asmDirIncludePath, hashArm64, hashAMD64 string) error {

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

	funcs["shorten"] = func(input string) string {
		if len(input) > 15 {
			return input[:6] + "..." + input[len(input)-6:]
		}
		return input
	}

	funcs["ltu64"] = func(a, b uint64) bool {
		return a < b
	}

	generate := func(suffix string, templates []string, opts ...fieldOption) func() error {
		opt := fieldOptions(opts...)
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
		amd64d, err = newASMWrapperData(hashAMD64, asmDirIncludePath, amd64.ElementASMFileName(F.NbWords, F.NbBits))
		if err != nil {
			return err
		}
	}

	if F.GenerateOpsARM64 {
		arm64d, err = newASMWrapperData(hashArm64, asmDirIncludePath, arm64.ElementASMFileName(F.NbWords, F.NbBits))
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
	}

	var g errgroup.Group

	g.Go(generate("element.go", sourceFiles))
	g.Go(generate("doc.go", []string{element.Doc}))
	g.Go(generate("vector.go", []string{element.Vector}))
	g.Go(generate("arith.go", []string{element.Arith}, only(!F.F31)))
	g.Go(generate("element_test.go", testFiles))
	g.Go(generate("vector_test.go", []string{element.TestVector}))

	g.Go(generate("element_amd64.s", []string{element.IncludeASM}, only(F.GenerateOpsAMD64), withBuildTag("!purego"), withData(amd64d)))
	g.Go(generate("element_arm64.s", []string{element.IncludeASM}, only(F.GenerateOpsARM64), withBuildTag("!purego"), withData(arm64d)))

	g.Go(generate("element_amd64.go", []string{element.OpsAMD64, element.MulDoc}, only(F.GenerateOpsAMD64 && !F.F31), withBuildTag("!purego")))
	g.Go(generate("element_arm64.go", []string{element.OpsARM64, element.MulNoCarry, element.Reduce}, only(F.GenerateOpsARM64 && !F.F31), withBuildTag("!purego")))

	g.Go(generate("element_purego.go", []string{element.OpsNoAsm, element.MulCIOS, element.MulNoCarry, element.Reduce, element.MulDoc}, withBuildTag(pureGoBuildTag)))

	g.Go(generate("vector_amd64.go", []string{element.VectorOpsAmd64}, only(F.GenerateVectorOpsAMD64 && !F.F31), withBuildTag("!purego")))
	g.Go(generate("vector_amd64.go", []string{element.VectorOpsAmd64F31}, only(F.GenerateVectorOpsAMD64 && F.F31), withBuildTag("!purego")))
	g.Go(generate("vector_arm64.go", []string{element.VectorOpsArm64}, only(F.GenerateVectorOpsARM64 && !F.F31), withBuildTag("!purego")))
	g.Go(generate("vector_arm64.go", []string{element.VectorOpsArm64F31}, only(F.GenerateVectorOpsARM64 && F.F31), withBuildTag("!purego")))

	g.Go(generate("vector_purego.go", []string{element.VectorOpsPureGo}, withBuildTag(pureGoVectorBuildTag)))

	g.Go(generate("asm_adx.go", []string{element.Asm}, only(F.GenerateOpsAMD64 && !F.F31), withBuildTag("!noadx")))
	g.Go(generate("asm_noadx.go", []string{element.AsmNoAdx}, only(F.GenerateOpsAMD64 && !F.F31), withBuildTag("noadx")))
	g.Go(generate("asm_avx.go", []string{element.Avx}, only(F.GenerateVectorOpsAMD64), withBuildTag("!noavx")))
	g.Go(generate("asm_noavx.go", []string{element.NoAvx}, only(F.GenerateVectorOpsAMD64), withBuildTag("noavx")))

	if F.UseAddChain {
		g.Go(generate("element_exp.go", []string{element.FixedExp}))
	}

	return g.Wait()
}

type fieldOption func(*fieldConfig)
type fieldConfig struct {
	buildTag string
	skip     bool
	tmplData any
}

func withBuildTag(buildTag string) fieldOption {
	return func(opt *fieldConfig) {
		opt.buildTag = buildTag
	}
}

func only(condition bool) fieldOption {
	return func(opt *fieldConfig) {
		opt.skip = !condition
	}
}

func withData(data any) fieldOption {
	return func(opt *fieldConfig) {
		opt.tmplData = data
	}
}

// default options
func fieldOptions(opts ...fieldOption) fieldConfig {
	// apply options
	opt := fieldConfig{}
	for _, option := range opts {
		option(&opt)
	}
	return opt
}
