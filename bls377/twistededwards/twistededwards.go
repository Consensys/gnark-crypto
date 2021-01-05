package twistededwards

import (
	"math/big"
	"sync"

	"github.com/consensys/gurvy/bls377/fr"
)

// CurveParams curve parameters: ax^2 + y^2 = 1 + d*x^2*y^2
type CurveParams struct {
	A, D     fr.Element // in Montgomery form
	Cofactor fr.Element // not in Montgomery form
	Order    big.Int
	Base     Point
}

var edwards CurveParams
var initOnce sync.Once

// GetEdwardsCurve returns the twisted Edwards curve on BN256's Fr
func GetEdwardsCurve() CurveParams {
	initOnce.Do(initEdBLS377)
	return edwards
}

func initEdBLS377() {

	edwards.A.SetOne().Neg(&edwards.A)
	edwards.D.SetUint64(3021)
	edwards.Cofactor.SetUint64(4).FromMont()
	edwards.Order.SetString("2111115437357092606062206234695386632838870926408408195193685246394721360383", 10)

	edwards.Base.X.SetString("717051916204163000937139483451426116831771857428389560441264442629694842243")
	edwards.Base.Y.SetString("882565546457454111605105352482086902132191855952243170543452705048019814192")
}
