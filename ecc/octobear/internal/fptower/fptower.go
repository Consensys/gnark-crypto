package fptower

import ext "github.com/consensys/gnark-crypto/field/koalabear/extensions"

type E2 = ext.E2
type E4 = ext.E4
type E8 = ext.E8

func BatchInvertE2(a []E2) []E2 {
	return ext.BatchInvertE2(a)
}

func BatchInvertE4(a []E4) []E4 {
	return ext.BatchInvertE4(a)
}

func BatchInvertE8(a []E8) []E8 {
	return ext.BatchInvertE8(a)
}
