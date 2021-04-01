// Copyright 2020 ConsenSys AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bw6761

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// ------------------------------------------------------------
// tests

func TestPairing(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	genA := GenE6()
	genR1 := GenFr()
	genR2 := GenFr()

	properties.Property("Having the receiver as operand (final expo) should output the same result", prop.ForAll(
		func(a *GT) bool {
			var b GT
			b.Set(a)
			b = FinalExponentiation(a)
			*a = FinalExponentiation(a)
			return a.Equal(&b)
		},
		genA,
	))

	properties.Property("Exponentiating FinalExpo(a) to r should output 1", prop.ForAll(
		func(a *GT) bool {
			var one GT
			var e big.Int
			e.SetString("258664426012969094010652733694893533536393512754914660539884262666720468348340822774968888139573360124440321458177", 10)
			one.SetOne()
			*a = FinalExponentiation(a)
			a.Exp(a, e)
			return a.Equal(&one)
		},
		genA,
	))

	properties.Property("FinalExp(a) should be equal to a^e", prop.ForAll(
		func(a *GT) bool {

			// e = 3(x^3-x^2+1)(q^6-1)/r
			var e big.Int
			e.SetString("1094403144802124617867780851426463544285069002716170704713418137000569118907851781115354339502103335729147809950641041505849492484126431876995151997381667958797779224941233215197441412437092725980391961640633014639822556404310183763816507751012376266254656863203671517087551882146388206524580199414399443371324115369257042473844122621031826081186858688047456408969500399247447267566338193281611799750747236046477714706394747930231752399392518822000670698993951366756208602565719025811329981375290498512149657425723584702550273771019186765565774567960098223347191858076768150159376144907860066768272548518578200075354576101938598164045970330220588604765063584831816343017603697369571320548088911048017738876167717501631779747212959920834561272933455782425735618542607253734264933415951278960038466951732825689182433431428912581865642063612703052527492278655498677228493387352648650210619560512868697329152911744344490254295794110402115720360153038096427777557450451732448549078802615647945298953901742254125682654958657435345416634369348855681760165633863304980585478332758842522554853682532253742502676018826471581925487625355127309376612062579625277485424869295671099592527555536703800681983319656004857113530938799305920556587793659058141024439784508499092558526614308067585741575912955418753000527605847958738062200", 10)

			var res1, res2 GT

			// naive
			res2.Exp(a, e)

			// opt

			res1 = FinalExponentiation(a)
			return res1.Equal(&res2)
		},
		genA,
	))

	properties.Property("easy part of FinalExp(a) should be equal to a^e", prop.ForAll(
		func(elt *GT) bool {

			// e = (q^3-1)(q+1)
			var e big.Int
			e.SetString("2255498460544541184709993096103282161623896302272660120614110311283831536626044882052953658628759088484386953324377806098048574694616358372054835890340693912765817636966992255171623244927411299606788929385695947716821964265427467546585940187866261876873645222993304801845386635088547227670976835176618784247129024158888311049605792091872259655114443809165945067685856751666761982906786886518698640929460216780737033442323749979556030071527619031504713115743517416025483791492517789113237965076285300901518956644855961457773470660608684917229153958479338135373256552625552483282267587361576745595331136530677821280003274439470493533651040569432739093255018010099891451631105607307429737939497189452683665986283721066624119991858515786931259859421616935859914821729137979905737676340678076553959903797339100817016141564103062702557217036918678637503717692276013888429884973994533255741243416055347284026477040095533400", 10)

			var buf, res1, res2 GT

			// naive
			res1.Exp(elt, e)

			// opt
			buf.FrobeniusCube(elt)
			res2.Inverse(elt)
			buf.Mul(&buf, &res2)
			res2.Frobenius(&buf).
				MulAssign(&buf)

			return res1.Equal(&res2)
		},
		genA,
	))

	properties.Property("bilinearity", prop.ForAll(
		func(a, b fr.Element) bool {

			var res, resa, resb, resab, zero GT

			var aG1 G1Jac
			var bG2 G2Jac

			var g1affine, ag1 G1Affine
			var g2affine, bg2 G2Affine

			var abigint, bbigint, ab big.Int

			a.ToBigIntRegular(&abigint)
			b.ToBigIntRegular(&bbigint)
			ab.Mul(&abigint, &bbigint)

			g1affine.FromJacobian(&g1Gen)
			g2affine.FromJacobian(&g2Gen)

			aG1.ScalarMultiplication(&g1Gen, &abigint)
			bG2.ScalarMultiplication(&g2Gen, &bbigint)
			ag1.FromJacobian(&aG1)
			bg2.FromJacobian(&bG2)

			res, err := Pair([]G1Affine{g1affine}, []G2Affine{g2affine})
			if err != nil {
				t.Fatal(err)
			}
			resa, err = Pair([]G1Affine{ag1}, []G2Affine{g2affine})
			if err != nil {
				t.Fatal(err)
			}
			resb, err = Pair([]G1Affine{g1affine}, []G2Affine{bg2})
			if err != nil {
				t.Fatal(err)
			}
			resab.Exp(&res, ab)
			resa.Exp(&resa, bbigint)
			resb.Exp(&resb, abigint)

			return resab.Equal(&resa) && resab.Equal(&resb) && !res.Equal(&zero)

		},
		genR1,
		genR2,
	))

	properties.Property("MillerLoop of pairs should be equal to the product of MillerLoops", prop.ForAll(
		func(a, b fr.Element) bool {

			var simpleProd, factorizedProd GT

			var ag1 G1Affine
			var bg2 G2Affine

			var abigint, bbigint big.Int

			a.ToBigIntRegular(&abigint)
			b.ToBigIntRegular(&bbigint)

			ag1.ScalarMultiplication(&g1GenAff, &abigint)
			bg2.ScalarMultiplication(&g2GenAff, &bbigint)

			P0 := []G1Affine{g1GenAff}
			P1 := []G1Affine{ag1}
			Q0 := []G2Affine{g2GenAff}
			Q1 := []G2Affine{bg2}

			// FE( ML(a,b) * ML(c,d) * ML(e,f) * ML(g,h) )
			M1, _ := MillerLoop(P0, Q0)
			M2, _ := MillerLoop(P1, Q0)
			M3, _ := MillerLoop(P0, Q1)
			M4, _ := MillerLoop(P1, Q1)
			simpleProd.Mul(&M1, &M2).Mul(&simpleProd, &M3).Mul(&simpleProd, &M4)
			simpleProd = FinalExponentiation(&simpleProd)

			tabP := []G1Affine{g1GenAff, ag1, g1GenAff, ag1}
			tabQ := []G2Affine{g2GenAff, g2GenAff, bg2, bg2}

			// FE( ML([a,c,e,g] ; [b,d,f,h]) ) -> saves 3 squares in Fqk
			factorizedProd, _ = Pair(tabP, tabQ)

			return simpleProd.Equal(&factorizedProd)
		},
		genR1,
		genR2,
	))

	properties.Property("PairingCheck", prop.ForAll(
		func() bool {

			var g1GenAffNeg G1Affine
			g1GenAffNeg.Neg(&g1GenAff)
			tabP := []G1Affine{g1GenAff, g1GenAffNeg}
			tabQ := []G2Affine{g2GenAff, g2GenAff}

			res, _ := PairingCheck(tabP, tabQ)

			return res
		},
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// ------------------------------------------------------------
// benches

func BenchmarkPairing(b *testing.B) {

	var g1GenAff G1Affine
	var g2GenAff G2Affine

	g1GenAff.FromJacobian(&g1Gen)
	g2GenAff.FromJacobian(&g2Gen)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmp, _ := MillerLoop([]G1Affine{g1GenAff}, []G2Affine{g2GenAff})
		FinalExponentiation(&tmp)
	}
}

func BenchmarkMillerLoop(b *testing.B) {

	var g1GenAff G1Affine
	var g2GenAff G2Affine

	g1GenAff.FromJacobian(&g1Gen)
	g2GenAff.FromJacobian(&g2Gen)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MillerLoop([]G1Affine{g1GenAff}, []G2Affine{g2GenAff})
	}
}

func BenchmarkFinalExponentiation(b *testing.B) {

	var a GT
	a.SetRandom()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FinalExponentiation(&a)
	}

}
