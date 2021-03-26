package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/consensys/bavard"
)

const copyrightHolder = "ConsenSys Software Inc."

var bgen = bavard.NewBatchGenerator(copyrightHolder, "consensys/crypto")

//go:generate go run main.go
func main() {

	bls377 := templateData{
		RootPath: "../../bls377/",
		Curve:    "BLS377",
	}
	bls381 := templateData{
		RootPath: "../../bls381/",
		Curve:    "BLS381",
	}
	bn256 := templateData{
		RootPath: "../../bn256/",
		Curve:    "BN256",
	}

	bw761 := templateData{
		RootPath: "../../bw761/",
		Curve:    "BW761",
	}

	datas := []templateData{bls377, bls381, bn256, bw761}

	const importCurve = "../imports.go.tmpl"

	var wg sync.WaitGroup

	for _, d := range datas {

		wg.Add(1)

		go func(d templateData) {

			defer wg.Done()

			fftDir := filepath.Join(d.RootPath, "fft")

			entries := []bavard.EntryF{
				{File: filepath.Join(fftDir, "domain_test.go"), TemplateF: []string{"tests/domain.go.tmpl", importCurve}},
				{File: filepath.Join(fftDir, "domain.go"), TemplateF: []string{"domain.go.tmpl", importCurve}},
				{File: filepath.Join(fftDir, "fft_test.go"), TemplateF: []string{"tests/fft.go.tmpl", importCurve}},
				{File: filepath.Join(fftDir, "fft.go"), TemplateF: []string{"fft.go.tmpl", importCurve}},
			}
			if err := bgen.GenerateF(d, "fft", "./template/fft/", entries...); err != nil {
				panic(err)
			}

		}(d)

	}

	wg.Wait()

	// run go fmt on whole directory
	cmd := exec.Command("gofmt", "-s", "-w", "../../../")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}

}

type templateData struct {
	RootPath string
	Curve    string // BLS381, BLS377, BN256, BW761
}
