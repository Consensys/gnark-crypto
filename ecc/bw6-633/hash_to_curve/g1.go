// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package hash_to_curve

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bw6-633/fp"
)

// Note: This only works for simple extensions

func g1IsogenyXNumerator(dst *fp.Element, x *fp.Element) {
	g1EvalPolynomial(dst,
		false,
		[]fp.Element{
			{7272862494598035149, 17938690902425062231, 17767077323866067514, 4042726381719152827, 18265955203653031758, 7989828268732772972, 18236438619995713786, 9293367820166308923, 579132017742513255, 61135481787611126},
			{14103265042103244800, 8689209911719872644, 4354557994639147275, 10015585990482518164, 7421596411839167820, 4090839137155710457, 11608543132371831098, 13726191584359776921, 5686323290725750987, 81414521630725786},
			{18168148533207762438, 11427550061515796123, 18288052689977291309, 17549058602874803907, 1811835605999111241, 14482385559564730468, 18371679523324113956, 7570413432649696247, 5685594971732218569, 10775296935876024},
			{17728445431987990123, 8760786791117372111, 2573782286514545042, 1334017444068088326, 13467298694204265433, 6851979702063838450, 988626161393389398, 16562654212673250360, 12282014862309670411, 80101036366709613},
			{2326830949067553368, 3529405818358416595, 146860478980129391, 1676900901388419777, 1082703688237676999, 7387945513834349563, 9650183203768502904, 11145522862419582316, 13264718987205403361, 13256495642676566},
			{12664678546015998202, 10045674430418935451, 7730403065839041631, 8219534329550967867, 1124289072196216451, 14392018765708614948, 16565553149614556003, 15197943132749823589, 12019150165746003035, 24988392243039541},
			{6694676015950629658, 7127563144408641267, 7082407350490358984, 17809866550737839973, 8771033794417365609, 10016426335785311067, 12427530939245540241, 3052758539102218458, 11641228273074855077, 25565165552182336},
			{9724060856090837937, 991577158671892474, 8081435887940851138, 17002015655933157445, 16512917147008832308, 4279771572102941926, 2249177714740675361, 7317077314484777626, 9914066020430477973, 6118378914363629},
		},
		x)
}

func g1IsogenyXDenominator(dst *fp.Element, x *fp.Element) {
	g1EvalPolynomial(dst,
		true,
		[]fp.Element{
			{7689367989221506360, 8466465815454855662, 5983600713003646820, 2581385575943478091, 9916432639299762387, 10803845534690581303, 13766744687398880084, 15045951258274284650, 10804811560468246090, 29872680157072640},
			{17931533655233022896, 2616834597954890161, 15297975164914689963, 12590465759698493931, 10233486349520996474, 735196127439996935, 3560747164676852194, 10735851124747749259, 17674946501015660508, 3306627583693497},
			{17716592541637989415, 2308856258992909515, 12321274951566722095, 15518009541869506179, 17585856967648697577, 13662172177461470988, 11485518780630966140, 8578584312450957544, 6227931909735105992, 33357481697750278},
			{16987009127955058020, 9531687309899711333, 8408043309689180103, 607401699215863109, 18292924216570439182, 13998767425014867069, 5338398210358867446, 14267801732372066290, 6813736547055367223, 61325738952091138},
			{13421066935948937976, 12318222198047014607, 14834312436015053872, 16700627760283291334, 4891844673141485395, 6252497771692828833, 11433160297376574630, 9870447387776384328, 11354566386586086837, 63511274412395176},
			{3111070609570438967, 6386169732085448708, 3204145000601381420, 17010231417177763314, 8617715899066662393, 536344151661697031, 7755071729129494035, 3825968914207782957, 2784415075519314680, 9751775967500675},
		},
		x)
}

func g1IsogenyYNumerator(dst *fp.Element, x *fp.Element, y *fp.Element) {
	var _dst fp.Element
	g1EvalPolynomial(&_dst,
		false,
		[]fp.Element{
			{9878447280690208476, 18035489718126671479, 16185036754982099195, 7328840215175603832, 3327555148021634143, 4731603924914666172, 14264314873550791027, 4538983491650033364, 1173609413709250348, 10349586991384491},
			{6013373483158194119, 10820801813182078326, 14598847571665760046, 11500540534236838924, 7782052717321334748, 14140166529189609834, 2964884042771384841, 16207414686346775778, 4629732333568574423, 4878794148730819},
			{12584528423975290756, 5798690717935596347, 6840751821254968468, 4803032538791716164, 4985526036624269917, 1479700245295912748, 13313798593102419181, 4969411887124120857, 3015920959041820281, 44921547284500425},
			{10461467081920196626, 16934070357422525213, 9517834684613575055, 2313584312088805570, 9031354913035368886, 12671644525466129236, 15886041369872958427, 16719536892038789270, 3829106597473744278, 1931735510244152},
			{14746212271987435353, 16614700088184332548, 12732527948961695321, 8527366302977827323, 17516184175149545265, 16732051276961112757, 13204423237463486744, 5000382421388638456, 7180920511183622623, 11555120281504263},
			{8663947747845995257, 16792483225373063374, 18268521190375255935, 4876326292498032031, 736768499122237671, 18244916601541093591, 4363883771233911261, 8989384718461158638, 12859624057357757038, 80977983165161056},
			{1888392448147101547, 278604480652219071, 18128419606300115944, 9201243398782896594, 6846159496103160821, 18248972704629781971, 8579402342792036543, 10976917330220264136, 11266589230198511881, 51029070295934750},
			{14247728228070040725, 11277441241513723962, 13278074988262372791, 5925040732263282901, 4830744489723491604, 13429219773244868874, 9574076864314497862, 15801710222777129423, 8208683544227273702, 13968877296974686},
			{10004935117279285226, 17549535430748601812, 12561470574481243390, 17245589729476447997, 6611491938405591356, 8408236954437324051, 8684479128757933086, 15579257205718648740, 13276819376284075144, 6359286773003154},
			{12326780164586609392, 5644226856352746503, 15218547976647793114, 1423996738351226436, 17696389932683702646, 16585810622359183722, 11617060654723736714, 10287699425090206986, 6288819172487953077, 58313628540302044},
		},
		x)

	dst.Mul(&_dst, y)
}

func g1IsogenyYDenominator(dst *fp.Element, x *fp.Element) {
	g1EvalPolynomial(dst,
		true,
		[]fp.Element{
			{14197616310547505477, 13248972589415769338, 5655590928954867770, 14499450495178324765, 10477012947100229518, 18368542199165419613, 2341453125094301785, 13165751376957884540, 2845217790110152808, 67501159738732262},
			{10329124799877583249, 8756876257370047475, 2027304531048772658, 14136513784278359146, 10612935083520178707, 6207950952294049860, 5989834420685599066, 16200344836864102349, 17989931768761686572, 36605130150858281},
			{15712638640355571572, 6911732328807606292, 10176318194353482793, 3635851365566106398, 9892567339900120038, 14461612377214612955, 16389615241393274057, 12651838945236028812, 13536083907098569269, 18515749350944921},
			{5076827480350464506, 18249575656742683023, 7202276925283013655, 1675474444887786899, 14564549657361301580, 13651349150863069141, 5233748578191063137, 10057693774177509688, 2263095285762743999, 44987482788755588},
			{2313434157949960381, 15493925599360292661, 12462877951536944165, 3484089169656765135, 16631132581765223246, 9606226240817426378, 9081464327451896874, 7882188674703192698, 14480525291200472943, 32841521205681759},
			{14971360958626235018, 6354257353819984162, 16708799287338381112, 2926594902781714529, 10489049316241960411, 18330102542973983952, 12044258370667954725, 2704306777095480810, 11580584760279813236, 24375418591899070},
			{6032389327291386739, 3861083195237077814, 3011703309625364553, 6313989484955116904, 10381012665260180778, 3855632300083161044, 11508306303084133188, 7774069649892635363, 13752428322375589071, 7164666883386632},
			{18309283861373580155, 15569853377733268658, 16820330372723914620, 7813464936948997015, 12356527369467981372, 7637946307109476197, 15283013241197161945, 1140793393437432174, 149111516692432071, 15556285163230785},
			{12423083707804413657, 2561324669216553892, 3354623821752926293, 11610276741867222767, 10363268849840872219, 5463863190870447699, 14455444601362058804, 2604004256097566119, 13259551450805497295, 56059041820898806},
		},
		x)
}

// G1 computes the isogeny map of the curve element, given by its coordinates pX and pY.
// It mutates the coordinates pX and pY to the new coordinates of the isogeny map.
func G1Isogeny(pX, pY *fp.Element) {

	den := make([]fp.Element, 2)

	g1IsogenyYDenominator(&den[1], pX)
	g1IsogenyXDenominator(&den[0], pX)

	g1IsogenyYNumerator(pY, pX, pY)
	g1IsogenyXNumerator(pX, pX)

	den = fp.BatchInvert(den)

	pX.Mul(pX, &den[0])
	pY.Mul(pY, &den[1])
}

// G1SqrtRatio computes the square root of u/v and returns 0 iff u/v was indeed a quadratic residue.
// If not, we get sqrt(Z * u / v). Recall that Z is non-residue.
// If v = 0, u/v is meaningless and the output is unspecified, without raising an error.
// The main idea is that since the computation of the square root involves taking large powers of u/v, the inversion of v can be avoided
func G1SqrtRatio(z *fp.Element, u *fp.Element, v *fp.Element) uint64 {

	// https://www.ietf.org/archive/id/draft-irtf-cfrg-hash-to-curve-16.html#name-optimized-sqrt_ratio-for-q-5 (mod 8)

	var tv1, tv2 fp.Element
	tv1.Square(v)       // 1. tv1 = v²
	tv2.Mul(&tv1, v)    // 2. tv2 = tv1 * v
	tv1.Square(&tv1)    // 3. tv1 = tv1²
	tv2.Mul(&tv2, u)    // 4. tv2 = tv2 * u
	tv1.Mul(&tv1, &tv2) // 5. tv1 = tv1 * tv2

	var c1 big.Int
	// c1 = (q - 5) / 8 = 2561809830520971834851673423317370187208698865113597259441094318876502093964723972342941881295000173427447071280550850682906970629308555147846507041252715387332745928667496467633282424373249
	c1.SetBytes([]byte{36, 204, 103, 152, 30, 107, 236, 127, 131, 66, 233, 224, 58, 229, 86, 181, 31, 155, 24, 235, 175, 58, 88, 233, 203, 46, 211, 91, 55, 123, 69, 240, 42, 84, 216, 31, 91, 212, 146, 23, 27, 83, 235, 208, 126, 175, 137, 47, 193, 209, 10, 29, 183, 180, 128, 250, 246, 185, 207, 87, 7, 56, 68, 167, 166, 211, 122, 98, 40, 254, 231, 154, 233, 34, 221, 72, 174, 0, 1})
	var y1 fp.Element
	y1.Exp(tv1, &c1)  // 6. y1 = tv1ᶜ¹
	y1.Mul(&y1, &tv2) // 7. y1 = y1 * tv2
	// c2 = sqrt(-1)
	c2 := fp.Element{7899625277197386435, 5217716493391639390, 7472932469883704682, 7632350077606897049, 9296070723299766388, 14353472371414671016, 14644604696869838127, 11421353192299464576, 237964513547175570, 46667570639865841}
	tv1.Mul(&y1, &c2) // 8. tv1 = y1 * c2
	tv2.Square(&tv1)  // 9. tv2 = tv1²
	tv2.Mul(&tv2, v)  // 10. tv2 = tv2 * v
	// 11. e1 = tv2 == u
	y1.Select(int(tv2.NotEqual(u)), &tv1, &y1) // 12. y1 = CMOV(y1, tv1, e1)
	tv2.Square(&y1)                            // 13. tv2 = y1²
	tv2.Mul(&tv2, v)                           // 14. tv2 = tv2 * v
	isQNr := tv2.NotEqual(u)                   // 15. isQR = tv2 == u
	// c3 = sqrt(Z / c2)
	y2 := fp.Element{17779787117422825675, 4939796438941267298, 17917495165866445585, 10745325761140663972, 1801762923695319826, 15143865745144854318, 15159144602769454149, 1697730798702809869, 18119907780441139825, 48351383131100009}
	y2.Mul(&y1, &y2)  // 16. y2 = y1 * c3
	tv1.Mul(&y2, &c2) // 17. tv1 = y2 * c2
	tv2.Square(&tv1)  // 18. tv2 = tv1²
	tv2.Mul(&tv2, v)  // 19. tv2 = tv2 * v
	var tv3 fp.Element
	// Z = [11]
	G1MulByZ(&tv3, u) // 20. tv3 = Z * u
	// 21. e2 = tv2 == tv3
	y2.Select(int(tv2.NotEqual(&tv3)), &tv1, &y2) // 22. y2 = CMOV(y2, tv1, e2)
	z.Select(int(isQNr), &y1, &y2)                // 23. y = CMOV(y2, y1, isQR)
	return isQNr
}

// G1MulByZ multiplies x by [11] and stores the result in z
func G1MulByZ(z *fp.Element, x *fp.Element) {

	res := *x

	res.Double(&res)
	res.Double(&res)
	res.Add(&res, x)
	res.Double(&res)
	res.Add(&res, x)

	*z = res
}

func g1EvalPolynomial(z *fp.Element, monic bool, coefficients []fp.Element, x *fp.Element) {
	dst := coefficients[len(coefficients)-1]

	if monic {
		dst.Add(&dst, x)
	}

	for i := len(coefficients) - 2; i >= 0; i-- {
		dst.Mul(&dst, x)
		dst.Add(&dst, &coefficients[i])
	}

	z.Set(&dst)
}

// G1Sgn0 is an algebraic substitute for the notion of sign in ordered fields.
// Namely, every non-zero quadratic residue in a finite field of characteristic =/= 2 has exactly two square roots, one of each sign.
//
// See: https://www.ietf.org/archive/id/draft-irtf-cfrg-hash-to-curve-16.html#name-the-sgn0-function
//
// The sign of an element is not obviously related to that of its Montgomery form
func G1Sgn0(z *fp.Element) uint64 {

	nonMont := z.Bits()

	// m == 1
	return nonMont[0] % 2

}

func G1NotZero(x *fp.Element) uint64 {

	return x[0] | x[1] | x[2] | x[3] | x[4] | x[5] | x[6] | x[7] | x[8] | x[9]

}
