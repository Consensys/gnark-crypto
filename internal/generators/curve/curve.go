package curve

const (
	fpName    = "fp"
	frName    = "fr"
	fp2Name   = "E2"
	fp6Name   = "E6"
	fp12Name  = "E12"
	pointName = "G"
)

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

	// data needed in the template, initialized to constants by Init() method
	FpName   string
	FrName   string
	Fp2Name  string
	Fp6Name  string
	Fp12Name string
}

// Init initialize string constants such as z.Fp2Name, etc; return z
func (z *Data) Init() *Data {
	z.FpName = fpName
	z.FrName = frName
	z.Fp2Name = fp2Name
	z.Fp6Name = fp6Name
	z.Fp12Name = fp12Name
	z.PointName = pointName
	return z
}

// CTemplate is the template to generate code to initialize C to a specific curve
const CTemplate = `
// C holds data for a specific curve
// Examples: BLS12-381, BLS12-377, BN256, BW6-761
var C Data

func init() {
	C = Data{
		Fpackage:        "{{$.Fpackage}}",
		FpModulus:       "{{$.FpModulus}}",
		FrModulus:       "{{$.FrModulus}}",
		Fp2NonResidue:   "{{$.Fp2NonResidue}}",
		Fp6NonResidue:   "{{$.Fp6NonResidue}}",
		EmbeddingDegree: {{$.EmbeddingDegree}},
		T:               "{{$.T}}",
		TNeg:            {{$.TNeg}},
		ThirdRootOne:    "{{$.ThirdRootOne}}",
		Lambda:          "{{$.Lambda}}",
		Size1:           "{{$.Size1}}",
		Size2:           "{{$.Size2}}",
	}
	C.Init()
}
`
