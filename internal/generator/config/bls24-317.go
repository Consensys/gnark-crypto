package config

var BLS24_317 = Curve{
	Name:         "bls24-317",
	CurvePackage: "bls24317",
	EnumID:       "BLS24_317",
	FrModulus:    "30869589236456844204538189757527902584594726589286811523515204428962673459201",
	FpModulus:    "136393071104295911515099765908274057061945112121419593977210139303905973197232025618026156731051",
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

var tBLS24_317 = TwistedEdwardsCurve{
	Name:     BLS24_317.Name,
	Package:  "twistededwards",
	EnumID:   BLS24_317.EnumID,
	A:        "-1",
	D:        "20748505950524021841644589704740731932416084248011369709738936344973878925081",
	Cofactor: "8",
	Order:    "3858698654557105525567273719690987823069521430163883173133245580997415449969",
	BaseX:    "4348505656527095883506785370890963704100065639426869666063106978260788240233",
	BaseY:    "1929349327278552762783636859845493911537170411830425720219700276810167091201",
}

func init() {
	addCurve(&BLS24_317)
	addTwistedEdwardCurve(&tBLS24_317)
}
