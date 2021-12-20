package config

var BLS12_378 = Curve{
	Name:         "bls12-378",
	CurvePackage: "bls12378",
	EnumID:       "BLS12_378",
	FrModulus:    "14883435066912132899950318861128167269793560281114003360875131245101026639873",
	FpModulus:    "605248206075306171733248481581800960739847691770924913753520744034740935903401304776283802348837311170974282940417",
	G1: Point{
		CoordType:        "fp.Element",
		PointName:        "g1",
		GLV:              true,
		CofactorCleaning: true,
		CRange:           defaultCRange(),
	},
	G2: Point{
		CoordType:        "fptower.E2",
		PointName:        "g2",
		GLV:              true,
		CofactorCleaning: true,
		CRange:           defaultCRange(),
		Projective:       true,
	},
}

func init() {
	addCurve(&BLS12_378)

}
