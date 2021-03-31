package twistededwards

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"

	"testing"
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

	p1.X.SetString("21793328330329971148710654283888115697962123987759099803244199498744022094670")
	p1.Y.SetString("2101040637884652362150023747029283466236613497763786920682459476507158507058")

	p2.X.SetString("50629843885093813360334764484465489653158679010834922765195739220081842003850")
	p2.Y.SetString("39525475875082628301311747912064089490877815436253076910246067124459956047086")

	var expectedX, expectedY fr.Element

	expectedX.SetString("35199665011228459549784465709909589656817343715952606097903780358611765544262")
	expectedY.SetString("35317228978363680085508213497002527319878195549272460436820924737513178285870")

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

	p1.X.SetString("21793328330329971148710654283888115697962123987759099803244199498744022094670")
	p1.Y.SetString("2101040637884652362150023747029283466236613497763786920682459476507158507058")

	p2.X.SetString("50629843885093813360334764484465489653158679010834922765195739220081842003850")
	p2.Y.SetString("39525475875082628301311747912064089490877815436253076910246067124459956047086")

	p1proj.FromAffine(&p1)
	p2proj.FromAffine(&p2)

	var expectedX, expectedY fr.Element

	expectedX.SetString("35199665011228459549784465709909589656817343715952606097903780358611765544262")
	expectedY.SetString("35317228978363680085508213497002527319878195549272460436820924737513178285870")

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

	p.X.SetString("21793328330329971148710654283888115697962123987759099803244199498744022094670")
	p.Y.SetString("2101040637884652362150023747029283466236613497763786920682459476507158507058")

	p.Double(&p)

	var expectedX, expectedY fr.Element

	expectedX.SetString("4887768767527220265359686405053440846384750454898507249732188959468533044182")
	expectedY.SetString("52332037604151508724685641460923103263088911891587010793017195088380209977878")

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

	p.X.SetString("21793328330329971148710654283888115697962123987759099803244199498744022094670")
	p.Y.SetString("2101040637884652362150023747029283466236613497763786920682459476507158507058")

	pproj.FromAffine(&p).Double(&pproj)

	p.FromProj(&pproj)

	var expectedX, expectedY fr.Element

	expectedX.SetString("4887768767527220265359686405053440846384750454898507249732188959468533044182")
	expectedY.SetString("52332037604151508724685641460923103263088911891587010793017195088380209977878")

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

	expectedX.SetString("46803808651513276177048978152090125758512142729856301157634295837210154385969")
	expectedY.SetString("6051280156044491864815311759850323556790635624820404123991533640491375546590")

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
