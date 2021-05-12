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

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
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

	p1.X.SetString("19913616433154696376327749871164055537288139465412275235880410113495361295094")
	p1.Y.SetString("19901468478985561483042531927595422521340692017806372336929063658330472937985")

	p2.X.SetString("2139491712597164764962668757454237126013140305826310350110400765900261000673")
	p2.Y.SetString("10486983842662731872207101904853058650818320076768283933077529934573055225572")

	var expectedX, expectedY fr.Element

	expectedX.SetString("10076403870840175294373195580515037340181970119938669337222735105343507498472")
	expectedY.SetString("17322669000658609068927118537392865214260971921316864664109543721402666310461")

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

	p1.X.SetString("19913616433154696376327749871164055537288139465412275235880410113495361295094")
	p1.Y.SetString("19901468478985561483042531927595422521340692017806372336929063658330472937985")

	p2.X.SetString("2139491712597164764962668757454237126013140305826310350110400765900261000673")
	p2.Y.SetString("10486983842662731872207101904853058650818320076768283933077529934573055225572")

	p1proj.FromAffine(&p1)
	p2proj.FromAffine(&p2)

	var expectedX, expectedY fr.Element

	expectedX.SetString("10076403870840175294373195580515037340181970119938669337222735105343507498472")
	expectedY.SetString("17322669000658609068927118537392865214260971921316864664109543721402666310461")

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

	p.X.SetString("13895729259634002747227836666242409516408412631624015559476573210321865119376")
	p.Y.SetString("11709735077858872717129179662842229518834255722201269436041280986409129993414")

	p.Double(&p)

	var expectedX, expectedY fr.Element

	expectedX.SetString("6025722326359295734279334333210838918527972409625130218257369808064782275772")
	expectedY.SetString("5955245965638822723061875949502429753912238367078909922082472610892337073840")

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

	p.X.SetString("13895729259634002747227836666242409516408412631624015559476573210321865119376")
	p.Y.SetString("11709735077858872717129179662842229518834255722201269436041280986409129993414")

	pproj.FromAffine(&p).Double(&pproj)

	p.FromProj(&pproj)

	var expectedX, expectedY fr.Element

	expectedX.SetString("6025722326359295734279334333210838918527972409625130218257369808064782275772")
	expectedY.SetString("5955245965638822723061875949502429753912238367078909922082472610892337073840")

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

	expectedX.SetString("168832981142221655341708526283999562680740818212602108643953704367987598747")
	expectedY.SetString("12956808000482532416873382696451950668786244907047953547021024966691314258300")

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
