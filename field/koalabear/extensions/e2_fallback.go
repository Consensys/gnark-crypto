//go:build !amd64
// +build !amd64

package extensions

func addE2(z, x, y *E2) {
	z.A0.Add(&x.A0, &y.A0)
	z.A1.Add(&x.A1, &y.A1)
}

func subE2(z, x, y *E2) {
	z.A0.Sub(&x.A0, &y.A0)
	z.A1.Sub(&x.A1, &y.A1)
}

func doubleE2(z, x *E2) {
	z.A0.Double(&x.A0)
	z.A1.Double(&x.A1)
}

func negE2(z, x *E2) {
	z.A0.Neg(&x.A0)
	z.A1.Neg(&x.A1)
}
