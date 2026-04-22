package config

var KB8 = Curve{
	Name:         "kb8",
	CurvePackage: "kb8",
	EnumID:       "KB8",
	FpModulus:    "2130706433",
	FrModulus:    "424804331891979973455971894938199991839487883914575852667663156896715214921",
	NoFieldSuite: true,
	G1: Point{
		CoordType:        "fptower.E8",
		CoordExtDegree:   8,
		CoordExtRoot:     3,
		PointName:        "g1",
		GLV:              false,
		CofactorCleaning: false,
		CRange:           defaultCRange(),
	},
}

func init() {
	addCurve(&KB8)
}
