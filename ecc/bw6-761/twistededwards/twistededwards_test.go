/*
Copyright © 2020 ConsenSys

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package twistededwards

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
)

func TestMarshal(t *testing.T) {

	var point, unmarshalPoint PointAffine
	point.Set(&edwards.Base)
	for i := 0; i < 20; i++ {
		b := point.Marshal()
		unmarshalPoint.Unmarshal(b)
		if !point.Equal(&unmarshalPoint) {
			t.Fatal("error unmarshal(marshal(point))")
		}
		point.Add(&point, &edwards.Base)
	}
}

func TestAdd(t *testing.T) {

	var p1, p2 PointAffine

	p1.X.SetString("4660805967172645089332027186447592360624693072714941456967820085096911052783894439404581260403897471106649733701")
	p1.Y.SetString("79342197333265570482692723254078325538912461654698423568563311515262945474856990363451062603548870121784984328098")

	p2.X.SetString("202442289343722786926786056656912288699177418429572083060310922487795404713283317871094743588900666122004015698690")
	p2.Y.SetString("247974769815137517184505154194568378626561395274790041560253961956563064663465145089967669342879476230498518507723")

	var expectedX, expectedY fr.Element

	expectedX.SetString("87851088022013567519299190403299135803828044203039432649229763792595086402136081587920660418748136523195620878595")
	expectedY.SetString("234173824031311388296193366929495516049946083371350027293666792226170518872371897012196046409485243031080305414077")

	p1.Add(&p1, &p2)

	if !p1.X.Equal(&expectedX) {
		t.Fatal("wrong x coordinate")
	}
	if !p1.Y.Equal(&expectedY) {
		t.Fatal("wrong y coordinate")
	}

}

func TestAddProj(t *testing.T) {

	var p1, p2 PointAffine
	var p1proj, p2proj PointProj

	p1.X.SetString("4660805967172645089332027186447592360624693072714941456967820085096911052783894439404581260403897471106649733701")
	p1.Y.SetString("79342197333265570482692723254078325538912461654698423568563311515262945474856990363451062603548870121784984328098")

	p2.X.SetString("202442289343722786926786056656912288699177418429572083060310922487795404713283317871094743588900666122004015698690")
	p2.Y.SetString("247974769815137517184505154194568378626561395274790041560253961956563064663465145089967669342879476230498518507723")

	p1proj.FromAffine(&p1)
	p2proj.FromAffine(&p2)

	var expectedX, expectedY fr.Element

	expectedX.SetString("87851088022013567519299190403299135803828044203039432649229763792595086402136081587920660418748136523195620878595")
	expectedY.SetString("234173824031311388296193366929495516049946083371350027293666792226170518872371897012196046409485243031080305414077")

	p1proj.Add(&p1proj, &p2proj)
	p1.FromProj(&p1proj)

	if !p1.X.Equal(&expectedX) {
		t.Fatal("wrong x coordinate")
	}
	if !p1.Y.Equal(&expectedY) {
		t.Fatal("wrong y coordinate")
	}

}

func TestDouble(t *testing.T) {

	var p PointAffine

	p.X.SetString("246748149935170006442122218590173395404179842918432028026503324708836243766162535400260211821848008636603589465819")
	p.Y.SetString("134422508542478271447341601263363314500854362589728108085433776240031286906803332386842498538810110264566838724397")

	p.Double(&p)

	var expectedX, expectedY fr.Element

	expectedX.SetString("229726374993824553210032647391680584407382633358123814205683186702063364578694113503631060548413333032166600059605")
	expectedY.SetString("7425633504134541917244736520302104892693206433544110384743886061187746044241073241043222108271067490316569505683")

	if !p.X.Equal(&expectedX) {
		t.Fatal("wrong x coordinate")
	}
	if !p.Y.Equal(&expectedY) {
		t.Fatal("wrong y coordinate")
	}
}

func TestDoubleProj(t *testing.T) {

	var p PointAffine
	var pproj PointProj

	p.X.SetString("246748149935170006442122218590173395404179842918432028026503324708836243766162535400260211821848008636603589465819")
	p.Y.SetString("134422508542478271447341601263363314500854362589728108085433776240031286906803332386842498538810110264566838724397")

	pproj.FromAffine(&p).Double(&pproj)

	p.FromProj(&pproj)

	var expectedX, expectedY fr.Element

	expectedX.SetString("229726374993824553210032647391680584407382633358123814205683186702063364578694113503631060548413333032166600059605")
	expectedY.SetString("7425633504134541917244736520302104892693206433544110384743886061187746044241073241043222108271067490316569505683")

	if !p.X.Equal(&expectedX) {
		t.Fatal("wrong x coordinate")
	}
	if !p.Y.Equal(&expectedY) {
		t.Fatal("wrong y coordinate")
	}
}

func TestScalarMul(t *testing.T) {

	// set curve parameters
	ed := GetEdwardsCurve()

	var scalar big.Int
	scalar.SetUint64(23902374)

	var p PointAffine
	p.ScalarMul(&ed.Base, &scalar)

	var expectedX, expectedY fr.Element

	expectedX.SetString("94508579205267879768236134677909410198642507835693837897321281457412423574974429773274172574457717235740673903198")
	expectedY.SetString("250269583191400352752955110332313615254848222347909213264040100110951145973009094833969596763932583696678722250546")

	if !expectedX.Equal(&p.X) {
		t.Fatal("wrong x coordinate")
	}
	if !expectedY.Equal(&p.Y) {
		t.Fatal("wrong y coordinate")
	}

	// test consistancy with negation
	var expected, base PointAffine
	expected.Set(&ed.Base).Neg(&expected)
	scalar.Set(&ed.Order).Lsh(&scalar, 3) // multiply by cofactor=8
	scalar.Sub(&scalar, big.NewInt(1))
	base.Set(&ed.Base)
	base.ScalarMul(&base, &scalar)
	if !base.Equal(&expected) {
		t.Fatal("Mul by order-1 not consistant with neg")
	}
}
