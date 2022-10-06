package polynomial

import (
	"github.com/consensys/gnark-crypto/internal/generator/gkr/small_rational"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEval(t *testing.T) {
	p := make(Polynomial, 3)
	p[0].SetInt64(1)
	p[1].SetInt64(-3)
	p[2].SetInt64(2)

	expectedP := make(Polynomial, 3)
	expectedP[0].SetInt64(1)
	expectedP[1].SetInt64(-3)
	expectedP[2].SetInt64(2)

	var x, expectedC small_rational.SmallRational
	x.SetInt64(0)

	expectedC.SetInt64(1)

	c := p.Eval(&x)

	assert.True(t, c.Equal(&expectedC))
	assert.True(t, p[2].Equal(&expectedP[2]), "evaluation shouldn't modify a polynomial. p[2] changed to %s", p[2].Text(10))
}

func TestEval1(t *testing.T) {
	p := make(Polynomial, 2)
	p[0].SetInt64(1)
	p[1].SetInt64(-3)

	expectedP := make(Polynomial, 2)
	expectedP[0].SetInt64(1)
	expectedP[1].SetInt64(-3)

	var x, expectedC small_rational.SmallRational
	x.SetInt64(0)

	expectedC.SetInt64(1)

	c := p.Eval(&x)

	assert.True(t, c.Equal(&expectedC))
	assert.True(t, p[1].Equal(&expectedP[1]), "evaluation shouldn't modify a polynomial. p[1] changed to %s", p[1].Text(10))
}
