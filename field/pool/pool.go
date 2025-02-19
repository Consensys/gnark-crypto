package pool

import (
	"math/big"
	"sync"
)

// BigInt is a shared *big.Int memory pool
var BigInt bigIntPool

var _bigIntPool = sync.Pool{
	New: func() interface{} {
		return new(big.Int)
	},
}

type bigIntPool struct{}

func (bigIntPool) Get() *big.Int {
	v, ok := _bigIntPool.Get().(*big.Int)
	if !ok {
		// If somehow we got a wrong type, create a new one
		return new(big.Int)
	}
	return v
}

func (bigIntPool) Put(v *big.Int) {
	if v == nil {
		return // see https://github.com/Consensys/gnark-crypto/issues/316
	}
	_bigIntPool.Put(v)
}
