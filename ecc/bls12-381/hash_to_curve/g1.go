// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package hash_to_curve

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
)

// Note: This only works for simple extensions

var (
	g1IsogenyXNumeratorMap = []fp.Element{
		{5555391298090832668, 1871845530032595596, 4551034694774233518, 2584197799339864836, 15085749040064757844, 654075415717002996},
		{9910598932128054667, 4357765064159749802, 1555960221863322426, 9671461638228026285, 1275132148248838779, 507072521670460589},
		{11908177372061066827, 18190436643933350086, 6603102998733829542, 6581045210674032871, 16099974426311393401, 541581077397919012},
		{5282195870529824577, 12365729195083706401, 2807246122435955773, 332702220601507168, 7339422050895811209, 1050416448951884523},
		{10443415753526299973, 8852419397684277637, 1088333252544296036, 1174353327457337436, 1626144519293139599, 716651429285276662},
		{7916646322956281527, 11909818257232418749, 1455301921509471421, 3317627683558310107, 12693445337245173919, 1798273032850409769},
		{2577731109215284733, 8810166123993386985, 3186592767751348067, 15050850291391518479, 18435652654155870871, 1330813445865859326},
		{8787912969482053798, 9653629252694769025, 1358451377919714320, 16331599695590198629, 13519934665722691825, 628078949001449512},
		{16605411443261943819, 9536014432113026165, 8685402948537367476, 16291074259433785035, 407185289045737198, 713426768049972652},
		{1001421273809907975, 724433776290697394, 16309429154639760781, 10003715605277815375, 307249038158020985, 688008371043525493},
		{16622893420529658311, 18333652517857227637, 2139173376235292830, 16496634502105693419, 5355299366650241487, 382770009771704860},
		{8276255265012938363, 9997870203437298645, 16819210142450232135, 5062450688048499179, 12776432501206859311, 1778476024187613533},
	}
	g1IsogenyXDenominatorMap = []fp.Element{
		{13358415881952098629, 12009257493157516192, 13928884382876484932, 12988314785833227070, 11244145530317148182, 100673949996487007},
		{2533162896381624793, 10578896196504721258, 4263020647280931071, 1255899686249737875, 17097124965295857733, 590960935246623182},
		{10990404485039254780, 5344458621503091696, 1718862119039451458, 11600049052019063549, 18389973225607751698, 1092616849767867362},
		{16377845895484993601, 15314247056264135931, 14543008873173635408, 4875476272346940127, 2030129768648768484, 1297689274107773964},
		{6376927397170316667, 1460555178565443615, 18156708192400235081, 14761117739963869762, 8361091377443400626, 1421233557303902229},
		{18127417459170613536, 5353764292720778676, 858818813615405862, 3528937506143354306, 12604964186779349896, 489837025077541867},
		{15285065477075910543, 3650488990300576179, 7274499670465195193, 16100555180954076900, 7580582425312971905, 896074979407586822},
		{7582945168915351799, 2506680954090651888, 10272835934257987876, 9924916350558121763, 13577194922650729507, 1698254565890367778},
		{2009730524583761661, 11053280693947850663, 14409256190409559425, 3658799329773368860, 13529638021208614900, 869243908766415668},
		{11058048790650732295, 7059501760293999296, 6596812464094265283, 14567744481299745071, 1591898617514919697, 1344004358835331304},
	}
	g1IsogenyYNumeratorMap = []fp.Element{
		{3122824077082063463, 2111517899915568999, 14844585557031220083, 14713720721132803039, 9041847780307969683, 950267513573868304},
		{11079511902567680319, 18338468344530008184, 6769016392463638666, 1504264063027988936, 8098359051856762276, 760455062874047829},
		{1430247552210236986, 3854575382974307965, 14917507996414511245, 207936139448560, 9498310774218301406, 1438631746617682181},
		{6654065794071117243, 2928282753802966791, 4144383358731160429, 12673586709493869907, 12918170109018188791, 844088361957958231},
		{6416330705244672319, 3552017270878949117, 7777490944331917312, 7917192495177481567, 7271851377118683537, 253926972271069325},
		{11903306495973637341, 11622313950541285762, 17991208474928993001, 12280964980743791783, 14941570282955772167, 143516344770893715},
		{7324386472845891920, 16310961984705608217, 14050364318273732029, 410622978843904432, 13407944087243235067, 570579643952782879},
		{10655681039374273828, 3913226275392147601, 9613292388335178165, 11852815148890010639, 17652581670569921892, 780578093363976825},
		{10454026283255684948, 15005802245309313587, 4420421943175638630, 18052347756729021570, 12181908985148691767, 1485233717472293779},
		{5056344670784885274, 15896288289018563095, 11120951801157184493, 7250506164525313606, 9295677455526059106, 1757175036496698059},
		{417067620545670182, 113740147118943311, 7666319924200602156, 1469963335415292317, 13482947512490784447, 1353298443678343909},
		{13069093794065563159, 18364685236451803588, 2235996605706292724, 1007629142299662669, 4077244143222018961, 162586537120788900},
		{12976751790971550752, 10256454045927919861, 8968423978443605586, 91636529236982767, 9459527627289574163, 949550897353139410},
		{10595118024452621845, 8010256778549625402, 10333144214150401956, 17682229685967587631, 8235697699445463546, 317883997785997129},
		{16894283457285346118, 10513943172407809423, 4685513162956315481, 11558261883362075118, 574375951146893083, 1159440548124233311},
		{9739780494108151959, 17207219630538774058, 553911396609642498, 6085929320386029624, 14175410874026216616, 1183751611824804793},
	}
	g1IsogenyYDenominatorMap = []fp.Element{
		{16963992846030154524, 1796759822929186144, 15995221960860457854, 8232142361908220707, 5977498266010213481, 759868220591477233},
		{7019489280640006651, 8025136855967848721, 17464762292772824538, 4490335113250743896, 7652702793653159798, 1129822927746498110},
		{3164260796573156764, 2639884922337322818, 1251365706181388855, 13142429936036186189, 359878619957828340, 126848055205862465},
		{17472832885692408710, 9911075278795900735, 2614390623136861791, 14474775734428698630, 6462878218464609418, 1225960780180864957},
		{3586995257703132870, 2143554115308730112, 15207899356205612465, 4372523065560113828, 12811868595146042778, 307251632623424763},
		{14298637377310410728, 10963101290308221781, 8192510423058716701, 1175370967867267532, 1029599188863854120, 678981456155013844},
		{11149806480082726900, 3664985661428410608, 18095361538178773836, 14174906593575241395, 15305104369759711886, 901234928011491053},
		{4727074327869776987, 15736954329525418288, 14642679026711520511, 11429849039208981702, 17333567062758618213, 951235897335772166},
		{9130114290642375589, 14069725355798443159, 6621984191700563591, 270173975669947883, 6218390495944243859, 1077419361593130421},
		{9144875514986933294, 16561351410666797616, 8591333879886582656, 15059370240386191395, 7834396448114781869, 946553772269403391},
		{17809450171377747225, 15896956440537434491, 8451524482089653422, 1694507265233574136, 18224201536921880842, 317503425606567070},
		{13940503876759740187, 8772047862193200131, 6080360161890657205, 7935486160089058373, 9407473295146243021, 1255078947940629503},
		{1160821217138360586, 13542760608074182996, 11595911004531652098, 18158686636947034451, 13330657138280564947, 1773960737279760188},
		{9132548444917292754, 16464415422105000789, 6319313500251671073, 12727658548847517900, 10985275115076354035, 1431541893474124246},
		{662485641082390837, 260809847827618849, 6177381409359357075, 18231947741742261351, 18128540110746580014, 1079107229429227022},
	}
)

func G1IsogenyMap() [4][]fp.Element {
	return [4][]fp.Element{
		g1IsogenyXNumeratorMap,
		g1IsogenyXDenominatorMap,
		g1IsogenyYNumeratorMap,
		g1IsogenyYDenominatorMap,
	}
}

func g1IsogenyXNumerator(dst *fp.Element, x *fp.Element) {
	g1EvalPolynomial(dst, false, g1IsogenyXNumeratorMap, x)
}

func g1IsogenyXDenominator(dst *fp.Element, x *fp.Element) {
	g1EvalPolynomial(dst, true, g1IsogenyXDenominatorMap, x)
}

func g1IsogenyYNumerator(dst *fp.Element, x *fp.Element, y *fp.Element) {
	var _dst fp.Element
	g1EvalPolynomial(&_dst, false, g1IsogenyYNumeratorMap, x)
	dst.Mul(&_dst, y)
}

func g1IsogenyYDenominator(dst *fp.Element, x *fp.Element) {
	g1EvalPolynomial(dst, true, g1IsogenyYDenominatorMap, x)
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
	// https://www.ietf.org/archive/id/draft-irtf-cfrg-hash-to-curve-16.html#name-optimized-sqrt_ratio-for-q- (3 mod 4)
	var tv1 fp.Element
	tv1.Square(v) // 1. tv1 = v²
	var tv2 fp.Element
	tv2.Mul(u, v)       // 2. tv2 = u * v
	tv1.Mul(&tv1, &tv2) // 3. tv1 = tv1 * tv2

	var y1 fp.Element
	{
		var c1 big.Int
		// c1 = 1000602388805416848354447456433976039139220704984751971333014534031007912622709466110671907282253916009473568139946
		c1.SetBytes([]byte{6, 128, 68, 122, 142, 95, 249, 166, 146, 198, 233, 237, 144, 210, 235, 53, 217, 29, 210, 225, 60, 225, 68, 175, 217, 204, 52, 168, 61, 172, 61, 137, 7, 170, 255, 255, 172, 84, 255, 255, 238, 127, 191, 255, 255, 255, 234, 170}) // c1 = (q - 3) / 4     # Integer arithmetic

		y1.Exp(tv1, &c1) // 4. y1 = tv1ᶜ¹
	}

	y1.Mul(&y1, &tv2) // 5. y1 = y1 * tv2

	var y2 fp.Element
	// c2 = sqrt(-Z)
	tv3 := fp.Element{17544630987809824292, 17306709551153317753, 8299808889594647786, 5930295261504720397, 675038575008112577, 167386374569371918}
	y2.Mul(&y1, &tv3)              // 6. y2 = y1 * c2
	tv3.Square(&y1)                // 7. tv3 = y1²
	tv3.Mul(&tv3, v)               // 8. tv3 = tv3 * v
	isQNr := tv3.NotEqual(u)       // 9. isQR = tv3 == u
	z.Select(int(isQNr), &y1, &y2) // 10. y = CMOV(y2, y1, isQR)
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

	return x[0] | x[1] | x[2] | x[3] | x[4] | x[5]

}
