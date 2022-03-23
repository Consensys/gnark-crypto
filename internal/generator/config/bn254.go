package config

var BN254 = Curve{
	Name:         "bn254",
	CurvePackage: "bn254",
	EnumID:       "BN254",
	FrModulus:    "21888242871839275222246405745257275088548364400416034343698204186575808495617",
	FpModulus:    "21888242871839275222246405745257275088696311157297823662689037894645226208583",
	G1: Point{
		CoordType:        "fp.Element",
		CoordExtDegree:   1,
		PointName:        "g1",
		GLV:              true,
		CofactorCleaning: false,
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

var tBN254 = TwistedEdwardsCurve{
	Name:     BN254.Name,
	Package:  "twistededwards",
	EnumID:   BN254.EnumID,
	A:        "-1",
	D:        "12181644023421730124874158521699555681764249180949974110617291017600649128846",
	Cofactor: "8",
	Order:    "2736030358979909402780800718157159386076813972158567259200215660948447373041",
	BaseX:    "9671717474070082183213120605117400219616337014328744928644933853176787189663",
	BaseY:    "16950150798460657717958625567821834550301663161624707787222815936182638968203",
}

func init() {
	addCurve(&BN254)
	addTwistedEdwardCurve(&tBN254)
}
