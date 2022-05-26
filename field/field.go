package field

import "math/big"

// Field represents a finite field
type Field interface {
	Modulus() *big.Int
}
