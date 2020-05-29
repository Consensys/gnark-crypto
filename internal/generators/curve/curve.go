package curve

// Data data used to generate the templates
type Data struct {
	Fpackage        string
	FpModulus       string
	FrModulus       string
	Fp2NonResidue   string
	Fp6NonResidue   string
	EmbeddingDegree int

	// pairing
	T    string
	TNeg bool

	// gpoint
	PointName    string // TODO this name cannot change; remove it
	ThirdRootOne string
	Lambda       string
	Size1        string
	Size2        string // TODO this is a function of Size1; remove it

	// data needed in the template, always set to constants
	Fp2Name  string
	Fp6Name  string
	Fp12Name string
}

const (
	FpName    = "fp"
	FrName    = "fr"
	Fp2Name   = "E2"
	Fp6Name   = "E6"
	Fp12Name  = "E12"
	PointName = "G"
)

const CTemplate = `
// C holds data for a specific curve
// Examples: BLS12-381, BLS12-377, BN256, BW6-761
var C Data

func init() {
	C = Data{
		Fpackage:        {{$.Fpackage}},
		FpModulus:       {{$.FpModulus}},
		FrModulus:       {{$.FrModulus}},
		Fp2NonResidue:   {{$.Fp2NonResidue}},
		Fp6NonResidue:   {{$.Fp6NonResidue}},
		EmbeddingDegree: {{$.EmbeddingDegree}},
		T:               {{$.T}},
		TNeg:            {{$.TNeg}},
		ThirdRootOne:    {{$.ThirdRootOne}},
		Lambda:          {{$.Lambda}},
		Size1:           {{$.Size1}},
		Size2:           {{$.Size2}},
		Fp2Name:         Fp2Name,
		Fp6Name:         Fp6Name,
		Fp12Name:        Fp12Name,
		PointName:       PointName,
	}
}
`
