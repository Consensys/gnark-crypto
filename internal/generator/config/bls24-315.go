package config

var BLS24_315 = Curve{
	Name:         "bls24-315",
	CurvePackage: "bls24315",
	EnumID:       "BLS24_315",
	FrModulus:    "11502027791375260645628074404575422495959608200132055716665986169834464870401",
	FpModulus:    "39705142709513438335025689890408969744933502416914749335064285505637884093126342347073617133569",
	G1: Point{
		CoordType:        "fp.Element",
		CoordExtDegree:   1,
		PointName:        "g1",
		GLV:              true,
		CofactorCleaning: true,
		CRange:           defaultCRange(),
	},
	G2: Point{
		CoordType:        "fptower.E4",
		CoordExtDegree:   4,
		PointName:        "g2",
		GLV:              true,
		CofactorCleaning: true,
		CRange:           defaultCRange(),
		Projective:       true,
	},
}

var tBLS24_315 = TwistedEdwardsCurve{
	Name:     BLS24_315.Name,
	Package:  "twistededwards",
	EnumID:   BLS24_315.EnumID,
	A:        "-1",
	D:        "8771873785799030510227956919069912715983412030268481769609515223557738569779",
	Cofactor: "8",
	Order:    "1437753473921907580703509300571927811987591765799164617677716990775193563777",
	BaseX:    "750878639751052675245442739791837325424717022593512121860796337974109802674",
	BaseY:    "1210739767513185331118744674165833946943116652645479549122735386298364723201",
}

func init() {
	addCurve(&BLS24_315)
	addTwistedEdwardCurve(&tBLS24_315)
}
