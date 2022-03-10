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
	HashE1: &HashSuite{
		A: []string{"0x1ae3a4617c510ea34b3c4687866d1616212919cefb9b37e860f40fde03873fc0a0bf847bffffff8b9857ffffffffff2"},
		B: []string{"0x16"},
		Z: []int{5},
		Isogeny: &Isogeny{
			XMap: RationalPolynomial{
				Num: [][]string{
					{"0x142abb491d3ccb00d65810beba93dbb0a661fd85974d6aa82c4bb2e1a3c84ffdd6ef419b80000000000000000000000"},
					{"0x4d9d782ee8a7b7630cd57be9a2ca555e2f689a3cb86f60022910be6480000004284600000000001"},
					{"0x142abb491d3ccb014ac44505178f6ec539a237640b7ceab573689a3cb86f600114885f32400000063c6900000000001"},
				},
				Den: [][]string{
					{"0x13675e0bba29edd8c3355efa68b295578bda268f2e1bd8008a442f99200000010a11800000000004"},
				},
			},
			YMap: RationalPolynomial{
				Num: [][]string{
					{"0x142abb491d3ccb014ac44505178f6ec539a237640b7ceab573689a3cb86f600114885f32400000063c68fffffffffff"},
					{"0x35c748c2f8a21d6af848e30c1b78229a46644922460e73f6faf06c327b438084815848140000010a11800000000002"},
					{"0xd71d230be288756a6446249c205dced645709767bd81c863eb7f8d8e4f15003f5f407b84000000a64af00000000002"},
					{"0x17872fd54cc6ecd6d73a5085f0d2013b6de7eb4a0d6711d3b14f5e9c2c81f001429f19baa0000007467a80000000001"},
				},
				Den: [][]string{
					{"0x1ae3a4617c510eac63b05c06ca1493b1a22d9f300f5138f1ef3622fba094800170b5d44300000008508bffffffffff9"},
					{"0x746c34465cfb9314934039de742f800d471ce75b14a710033d991d96c00000063c6900000000000c"},
					{"0x3a361a232e7dc98a49a01cef3a17c006a38e73ad8a5388019ecc8ecb600000031e3480000000000c"},
				},
			},
		},
	},
}

func init() {
	addCurve(&BLS12_377)

}
