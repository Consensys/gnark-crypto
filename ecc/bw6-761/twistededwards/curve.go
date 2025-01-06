// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package twistededwards

import (
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
)

// CurveParams curve parameters: ax^2 + y^2 = 1 + d*x^2*y^2
type CurveParams struct {
	A, D     fr.Element
	Cofactor fr.Element
	Order    big.Int
	Base     PointAffine
}

// GetEdwardsCurve returns the twisted Edwards curve on bw6-761/Fr
func GetEdwardsCurve() CurveParams {
	initOnce.Do(initCurveParams)
	// copy to keep Order private
	var res CurveParams

	res.A.Set(&curveParams.A)
	res.D.Set(&curveParams.D)
	res.Cofactor.Set(&curveParams.Cofactor)
	res.Order.Set(&curveParams.Order)
	res.Base.Set(&curveParams.Base)

	return res
}

var (
	initOnce    sync.Once
	curveParams CurveParams
)

func initCurveParams() {
	curveParams.A.SetString("-1")
	curveParams.D.SetString("79743")
	curveParams.Cofactor.SetString("8")
	curveParams.Order.SetString("32333053251621136751331591711861691692049189094364332567435817881934511297123972799646723302813083835942624121493", 10)

	curveParams.Base.X.SetString("109887223397525145051017418760180386187632078445902299543670312117371514695798874370143656894667315818446285582389")
	curveParams.Base.Y.SetString("31146823455109675839494591101665406662142618451815824757336761504421066243585705807124836638254810186490790034654")
}

// mulByA multiplies fr.Element by curveParams.A
func mulByA(x *fr.Element) {
	x.Neg(x)
}
