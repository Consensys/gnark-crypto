package generator

import (
	"github.com/consensys/gnark-crypto/field/generator/config"
)

type Option func(*generatorConfig)

type generatorConfig struct {
	fftConfig *config.FFT
	asmConfig *config.Assembly
}

func (cfg *generatorConfig) HasFFT() bool {
	return cfg.fftConfig != nil
}

func (cfg *generatorConfig) HasArm64() bool {
	return cfg.asmConfig != nil && cfg.asmConfig.BuildDir != ""
}

func (cfg *generatorConfig) HasAMD64() bool {
	return cfg.asmConfig != nil && cfg.asmConfig.BuildDir != ""
}

func WithFFT(cfg *config.FFT) Option {
	return func(opt *generatorConfig) {
		opt.fftConfig = cfg
	}
}

func WithASM(cfg *config.Assembly) Option {
	return func(opt *generatorConfig) {
		opt.asmConfig = cfg
	}
}

// default options
func generatorOptions(opts ...Option) generatorConfig {
	// apply options
	opt := generatorConfig{
		asmConfig: &config.Assembly{},
	}
	for _, option := range opts {
		option(&opt)
	}
	return opt
}
