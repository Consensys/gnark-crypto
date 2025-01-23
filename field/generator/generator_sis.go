package generator

import (
	"fmt"
	"math/big"
	"path/filepath"
	"strings"

	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/field/generator/config"
)

func generateSIS(F *config.Field, outputDir string) error {

	fieldImportPath, err := getImportPath(outputDir)
	if err != nil {
		return err
	}

	outputDir = filepath.Join(outputDir, "sis")

	entries := []bavard.Entry{
		{File: filepath.Join(outputDir, "sis.go"), Templates: []string{"sis.go.tmpl"}},
		{File: filepath.Join(outputDir, "sis_test.go"), Templates: []string{"sis.test.go.tmpl"}},
	}
	// only on field byte size == 32, we unroll a 64-wide FFT (used in linea for bls12-377)
	if F.NbBytes == 32 {
		entries = append(entries, bavard.Entry{File: filepath.Join(outputDir, "sis_fft.go"), Templates: []string{"fft.go.tmpl"}})
	}

	type sisTemplateData struct {
		FF               string
		FieldPackagePath string
		HasUnrolledFFT   bool
	}

	data := &sisTemplateData{
		FF:               F.PackageName,
		FieldPackagePath: fieldImportPath,
		HasUnrolledFFT:   F.NbBytes == 32,
	}

	funcs := make(map[string]interface{})
	funcs["bitReverse"] = bitReverse
	funcs["pow"] = pow

	bavardOpts := []func(*bavard.Bavard) error{bavard.Funcs(funcs)}
	if data.HasUnrolledFFT {
		funcs["partialFFT"] = partialFFT
	}

	bgen := bavard.NewBatchGenerator("Consensys Software Inc.", 2020, "consensys/gnark-crypto")

	sisTemplatesRootDir, err := findTemplatesRootDir()
	if err != nil {
		return err
	}
	sisTemplatesRootDir = filepath.Join(sisTemplatesRootDir, "sis")

	if err := bgen.GenerateWithOptions(data, "sis", sisTemplatesRootDir, bavardOpts, entries...); err != nil {
		return err
	}

	return runFormatters(outputDir)
}

// From linea-monorepo/prover/crypto/ringsis/templates/partial_fft.go at 6e15740

func partialFFT(domainSize, numField int, mask int64) string {

	gen := initializePartialFFTCodeGen(int64(domainSize), int64(numField), mask)

	gen.header()
	gen.indent()

	var (
		numStages int = log2Ceil(int(domainSize))
		numSplits int = 1
		splitSize int = int(domainSize)
	)

	for level := 0; level < numStages; level++ {
		for s := 0; s < numSplits; s++ {
			for k := 0; k < splitSize/2; k++ {
				gen.twiddleMulLine(s*splitSize+splitSize/2+k, numSplits-1+s)
			}
		}

		for s := 0; s < numSplits; s++ {
			for k := 0; k < splitSize/2; k++ {
				gen.butterFlyLine(s*splitSize+k, s*splitSize+splitSize/2+k)
			}
		}

		splitSize /= 2
		numSplits *= 2
	}

	gen.desindent()
	gen.tail()
	return gen.Builder.String()
}

func initializePartialFFTCodeGen(domainSize, numField, mask int64) PartialFFTCodeGen {
	res := PartialFFTCodeGen{
		DomainSize: int(domainSize),
		NumField:   int(numField),
		Mask:       int(mask),
		IsZero:     make([]bool, domainSize),
		Builder:    &strings.Builder{},
		NumIndent:  0,
	}

	for i := range res.IsZero {
		var (
			fieldSize = domainSize / numField
			bit       = i / int(fieldSize)
			isZero    = ((mask >> bit) & 1) == 0
		)

		res.IsZero[i] = isZero
	}

	return res
}

type PartialFFTCodeGen struct {
	DomainSize int
	NumField   int
	Mask       int
	Builder    *strings.Builder
	NumIndent  int
	IsZero     []bool
}

func (p *PartialFFTCodeGen) header() {
	writeIndent(p.Builder, p.NumIndent)
	line := fmt.Sprintf("func partialFFT_%v(a, twiddles fr.Vector) {\n", p.Mask)
	p.Builder.WriteString(line)
}

func (p *PartialFFTCodeGen) tail() {
	writeIndent(p.Builder, p.NumIndent)
	p.Builder.WriteString("}\n")
}

func (p *PartialFFTCodeGen) butterFlyLine(i, j int) {
	allZeroes := p.IsZero[i] && p.IsZero[j]
	if allZeroes {
		return
	}

	p.IsZero[i] = false
	p.IsZero[j] = false

	writeIndent(p.Builder, p.NumIndent)

	line := fmt.Sprintf("fr.Butterfly(&a[%v], &a[%v])\n", i, j)
	if _, err := p.Builder.WriteString(line); err != nil {
		panic(err)
	}
}

func (p *PartialFFTCodeGen) twiddleMulLine(i, twidPos int) {
	if p.IsZero[i] {
		return
	}

	writeIndent(p.Builder, p.NumIndent)

	line := fmt.Sprintf("a[%v].Mul(&a[%v], &twiddles[%v])\n", i, i, twidPos)
	if _, err := p.Builder.WriteString(line); err != nil {
		panic(err)
	}
}

func (p *PartialFFTCodeGen) desindent() {
	p.NumIndent--
}

func (p *PartialFFTCodeGen) indent() {
	p.NumIndent++
}

func writeIndent(w *strings.Builder, n int) {
	for i := 0; i < n; i++ {
		w.WriteString("\t")
	}
}

func log2Floor(a int) int {
	res := 0
	for i := a; i > 1; i = i >> 1 {
		res++
	}
	return res
}

func log2Ceil(a int) int {
	floor := log2Floor(a)
	if a != 1<<floor {
		floor++
	}
	return floor
}

func pow(base int64, pow int) int64 {
	var (
		b = new(big.Int).SetInt64(base)
		p = new(big.Int).SetInt64(int64(pow))
	)
	b.Exp(b, p, nil)

	if !b.IsInt64() {
		panic("could not cast big.Int to int64 as it overflows")
	}

	return b.Int64()
}
