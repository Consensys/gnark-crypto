package config

import "github.com/consensys/gurvy/field"

// Curve describes parameters of the curve useful for the template
type Curve struct {
	Name      string
	FpModulus string
	FrModulus string

	Fp           *field.Field
	Fr           *field.Field
	FpUnusedBits int
	G1           Point
	G2           Point
	dir          string
}

type Point struct {
	CoordType        string
	PointName        string
	GLV              bool  // scalar mulitplication using GLV
	CofactorCleaning bool  // flag telling if the Cofactor cleaning is available
	CRange           []int // multiexp bucket method: generate inner methods (with const arrays) for each c
}

var Curves []Curve

// TODO @gbotrel should be in the multiexp
func defaultCRange() []int {
	// default range for C values in the multiExp
	return []int{4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 20, 21, 22}
}
