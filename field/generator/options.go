package generator

import (
	"github.com/consensys/gnark-crypto/field/generator/config"
)

type Option func(*generatorConfig)

type generatorConfig struct {
	fftConfig      *config.FFT
	asmConfig      *config.Assembly
	withSIS        bool
	withPoseidon2  bool
	withExtensions bool
	withIOP        bool
}

func (cfg *generatorConfig) HasExtensions() bool {
	return cfg.withExtensions
}

func (cfg *generatorConfig) HasPoseidon2() bool {
	return cfg.withPoseidon2
}

func (cfg *generatorConfig) HasSIS() bool {
	return cfg.withSIS
}

func (cfg *generatorConfig) HashIOP() bool {
	return cfg.withIOP
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

func WithIOP() Option {
	return func(opt *generatorConfig) {
		opt.withIOP = true
	}
}

func WithSIS() Option {
	return func(opt *generatorConfig) {
		opt.withSIS = true
	}
}

func WithPoseidon2() Option {
	return func(opt *generatorConfig) {
		opt.withPoseidon2 = true
	}
}

func WithExtensions() Option {
	return func(opt *generatorConfig) {
		opt.withExtensions = true
	}
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
