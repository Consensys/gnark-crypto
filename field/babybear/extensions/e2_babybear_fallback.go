package extensions

// Mul sets z to the E2-product of x,y, returns z
func (z *E2) Mul(x, y *E2) *E2 { // E: undeclared name: E2
	mulGenericE2(z, x, y) // E: undeclared name: mulGenericE2
	return z
}

// Square sets z to the E2-product of x,x returns z
func (z *E2) Square(x *E2) *E2 { // E: undeclared name: E2
	squareGenericE2(z, x) // E: undeclared name: squareGenericE2
	return z
}
