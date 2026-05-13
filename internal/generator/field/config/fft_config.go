package config

type FFT struct {
	GeneratorFullMultiplicativeGroup uint64 // generator of \mathbb{F}_r^{*}
	GeneratorMaxTwoAdicSubgroup      string // generator of the maximum subgroup of size 2^<something>
	LogTwoOrderMaxTwoAdicSubgroup    string // log_2 of the max order of the max two adic subgroup
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
