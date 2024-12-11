package main

import (
	"github.com/consensys/bavard"
	"github.com/consensys/gnark-crypto/fft/config"
	"github.com/consensys/gnark-crypto/fft/generator"
)

const (
	copyrightHolder = "Consensys Software Inc."
	copyrightYear   = 2020
)

var bgen = bavard.NewBatchGenerator(copyrightHolder, copyrightYear, "consensys/gnark-crypto")

//go:generate go run main.go
func main() {

	for _, conf := range config.Configs {
		err := generator.Generate(conf, conf.OutputDir, bgen)
		if err != nil {
			panic(err)
		}
	}

}
