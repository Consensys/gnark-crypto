package config

var BW6_756 = Curve{
	Name:         "bw6-756",
	CurvePackage: "bw6756",
	EnumID:       "BW6_756",
	FrModulus:    "605248206075306171733248481581800960739847691770924913753520744034740935903401304776283802348837311170974282940417",
	FpModulus:    "366325390957376286590726555727219947825377821289246188278797409783441745356050456327989347160777465284190855125642086860525706497928518803244008749360363712553766506755227344593404398783886857865261088226271336335268413437902849",
	G1: Point{
		CoordType:        "fp.Element",
		PointName:        "g1",
		GLV:              true,
		CofactorCleaning: true,
		CRange:           []int{4, 5, 8, 16},
		Projective:       true,
	},
	G2: Point{
		CoordType:        "fp.Element",
		PointName:        "g2",
		GLV:              true,
		CofactorCleaning: true,
		CRange:           []int{4, 5, 8, 16},
	},
}

func init() {
	addCurve(&BW6_756)
}
