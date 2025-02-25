package config

type FFT struct {
	// TODO this should be in the finite field package API
	GeneratorFullMultiplicativeGroup uint64 // generator of \mathbb{F}_r^{*}

	// TODO should be generated by goff
	GeneratorMaxTwoAdicSubgroup string // generator of the maximum subgroup of size 2^<something>

	// TODO should be generated by goff
	LogTwoOrderMaxTwoAdicSubgroup string // log_2 of the max order of the max two adic subgroup
}

// NewConfig returns a data structure with needed information to generate apis for the FFT
func NewConfig(
	genFullMultiplicativeGroup uint64,
	generatorMaxTwoAdicSubgroup,
	logTwoOrderMaxTwoAdicSubgroup string) FFT {
	return FFT{
		GeneratorFullMultiplicativeGroup: genFullMultiplicativeGroup,
		GeneratorMaxTwoAdicSubgroup:      generatorMaxTwoAdicSubgroup,
		LogTwoOrderMaxTwoAdicSubgroup:    logTwoOrderMaxTwoAdicSubgroup,
	}
}
