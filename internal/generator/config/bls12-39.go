package config

var BLS12_39 = Curve{
	Name:         "bls12-39",
	CurvePackage: "bls1239",
	PackageDoc:   "implements a toy BLS curve for test purposes. Warning: not suitable for cryptographic use.",
	EnumID:       "BLS12_39",
	FrModulus:    "99990001",
	FpModulus:    "326667333367",
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
		CoordExtRoot:     3,
		PointName:        "g2",
		GLV:              true,
		CofactorCleaning: true,
		CRange:           defaultCRange(),
		Projective:       true,
	},
}

func init() {
	addCurve(&BLS12_39)
}
