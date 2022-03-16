package config

var BLS12_377 = Curve{
	Name:         "bls12-377",
	CurvePackage: "bls12377",
	EnumID:       "BLS12_377",
	FrModulus:    "8444461749428370424248824938781546531375899335154063827935233455917409239041",
	FpModulus:    "258664426012969094010652733694893533536393512754914660539884262666720468348340822774968888139573360124440321458177",
	G1: Point{
		CoordType:        "fp.Element",
		CoordExtDegree:   1,
		PointName:        "g1",
		GLV:              true,
		CofactorCleaning: true,
		CRange:           defaultCRange(),
	},
	G2: Point{
		CoordType:        "fptower.E2",
		CoordExtDegree:   2,
		PointName:        "g2",
		GLV:              true,
		CofactorCleaning: true,
		CRange:           defaultCRange(),
		Projective:       true,
	},
}

var tBLS12_77 = TwistedEdwardsCurve{
	Name:     BLS12_377.Name,
	Package:  "twistededwards",
	A:        "-1",
	D:        "3021",
	Cofactor: "4",
	Order:    "2111115437357092606062206234695386632838870926408408195193685246394721360383",
	BaseX:    "717051916204163000937139483451426116831771857428389560441264442629694842243",
	BaseY:    "882565546457454111605105352482086902132191855952243170543452705048019814192",
}

func init() {
	addCurve(&BLS12_377)
	addTwistedEdwardCurve(&tBLS12_77)
}
