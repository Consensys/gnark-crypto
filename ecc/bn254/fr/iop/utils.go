// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package iop

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

//----------------------------------------------------
// exp functions until 5

func exp0(x fr.Element) fr.Element {
	var res fr.Element
	res.SetOne()
	return res
}

func exp1(x fr.Element) fr.Element {
	return x
}

func exp2(x fr.Element) fr.Element {
	return *x.Square(&x)
}

func exp3(x fr.Element) fr.Element {
	var res fr.Element
	res.Square(&x).Mul(&res, &x)
	return res
}

func exp4(x fr.Element) fr.Element {
	x.Square(&x).Square(&x)
	return x
}

func exp5(x fr.Element) fr.Element {
	var res fr.Element
	res.Square(&x).Square(&res).Mul(&res, &x)
	return res
}

// doesn't return any errors, it is a private method, that
// is assumed to be called with correct arguments.
func smallExp(x fr.Element, n int) fr.Element {
	if n == 0 {
		return exp0(x)
	}
	if n == 1 {
		return exp1(x)
	}
	if n == 2 {
		return exp2(x)
	}
	if n == 3 {
		return exp3(x)
	}
	if n == 4 {
		return exp4(x)
	}
	if n == 5 {
		return exp5(x)
	}
	return fr.Element{}
}
