package config

var OCTOBEAR = Curve{
	Name:         "octobear",
	CurvePackage: "octobear",
	EnumID:       "OCTOBEAR",
	FpModulus:    "2130706433",
	FrModulus:    "424804331891979973455971894938199991839487883914575852667663156896715214921",
	NoFieldSuite: true,
	NoECC:        true,
	NoECDSA:      true,
	ExistingFp: ExistingFieldPackage{
		PackagePath: "github.com/consensys/gnark-crypto/field/koalabear",
		PackageName: "koalabear",
	},
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
	addCurve(&OCTOBEAR)
}
