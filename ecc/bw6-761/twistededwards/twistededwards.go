package twistededwards

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
)

// CurveParams curve parameters: ax^2 + y^2 = 1 + d*x^2*y^2
type CurveParams struct {
	A, D     fr.Element // in Montgomery form
	Cofactor fr.Element // not in Montgomery form
	Order    big.Int
	Base     PointAffine
}

var edwards CurveParams

// GetEdwardsCurve returns the twisted Edwards curve on BN254's Fr
func GetEdwardsCurve() CurveParams {
	// copy to keep Order private
	var res CurveParams

	res.A.Set(&edwards.A)
	res.D.Set(&edwards.D)
	res.Cofactor.Set(&edwards.Cofactor)
	res.Order.Set(&edwards.Order)
	res.Base.Set(&edwards.Base)

	return res
}

func init() {

	edwards.A.SetOne().Neg(&edwards.A)
	edwards.D.SetUint64(79743)
	edwards.Cofactor.SetUint64(8).FromMont()
	edwards.Order.SetString("32333053251621136751331591711861691692049189094364332567435817881934511297123972799646723302813083835942624121493", 10)

	edwards.Base.X.SetString("109887223397525145051017418760180386187632078445902299543670312117371514695798874370143656894667315818446285582389")
	edwards.Base.Y.SetString("31146823455109675839494591101665406662142618451815824757336761504421066243585705807124836638254810186490790034654")
}
