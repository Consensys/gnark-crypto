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

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
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

	p1.X.SetString("3315552591512299779303184808712317385227287137589771172094219636545300751065")
	p1.Y.SetString("3118680409475531351463685519263752131825008658604944897513392250405379901249")

	p2.X.SetString("6336723501920618784893835149136777251720250608098057515290887700547266428029")
	p2.Y.SetString("5496619275633907854559759526736419211322377071978307458692817697363815426689")

	var expectedX, expectedY fr.Element

	expectedX.SetString("4500820355051403554473541074419499489195466324706275819121927262358672448961")
	expectedY.SetString("1544163799156604767836846326520012060174626685575121143115994118939810697769")

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

	p1.X.SetString("3315552591512299779303184808712317385227287137589771172094219636545300751065")
	p1.Y.SetString("3118680409475531351463685519263752131825008658604944897513392250405379901249")

	p2.X.SetString("6336723501920618784893835149136777251720250608098057515290887700547266428029")
	p2.Y.SetString("5496619275633907854559759526736419211322377071978307458692817697363815426689")

	p1proj.FromAffine(&p1)
	p2proj.FromAffine(&p2)

	var expectedX, expectedY fr.Element

	expectedX.SetString("4500820355051403554473541074419499489195466324706275819121927262358672448961")
	expectedY.SetString("1544163799156604767836846326520012060174626685575121143115994118939810697769")

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

	p.X.SetString("4054273413031690150214184383897121503625518951056633275177212354976965625455")
	p.Y.SetString("3656215305282548767146189541242947042444006653545844852001327067666733281531")

	p.Double(&p)

	var expectedX, expectedY fr.Element

	expectedX.SetString("5216110673465714862705952302067819114633501501646084598951778107133052123079")
	expectedY.SetString("86503899552556854235590815953058319029599365042229318098224597874208785818")

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

	p.X.SetString("4054273413031690150214184383897121503625518951056633275177212354976965625455")
	p.Y.SetString("3656215305282548767146189541242947042444006653545844852001327067666733281531")

	pproj.FromAffine(&p).Double(&pproj)

	p.FromProj(&pproj)

	var expectedX, expectedY fr.Element

	expectedX.SetString("5216110673465714862705952302067819114633501501646084598951778107133052123079")
	expectedY.SetString("86503899552556854235590815953058319029599365042229318098224597874208785818")

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

	expectedX.SetString("3325318589486882180368597836061279041613850994206039496744772927680069206357")
	expectedY.SetString("1346375924217781592879811475536412101049343472231217752819745489083157050050")

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
