package hash

import (
	"crypto/rand"
	"io"
)

// HashFunction represents a hash function that can be used in MPC setup ceremonies
type HashFunction interface {
	// Hash expands a message to a slice of bytes using a domain separation tag
	Hash(msg, dst []byte, lenInBytes int) ([]byte, error)
}

// RandomnessSource represents a source of randomness that can be used in MPC setup ceremonies
type RandomnessSource interface {
	// GetRandomness fills the provided byte slice with random bytes
	GetRandomness(b []byte) (int, error)
}

// DefaultHashFunction is the default implementation of HashFunction
type DefaultHashFunction struct{}

// Hash implements the HashFunction interface using the default ExpandMsgXmd function
func (d DefaultHashFunction) Hash(msg, dst []byte, lenInBytes int) ([]byte, error) {
	return ExpandMsgXmd(msg, dst, lenInBytes)
}

// DefaultRandomnessSource is the default implementation of RandomnessSource
type DefaultRandomnessSource struct{}

// GetRandomness implements the RandomnessSource interface using crypto/rand
func (d DefaultRandomnessSource) GetRandomness(b []byte) (int, error) {
	return io.ReadFull(rand.Reader, b)
}

// Global instances that can be replaced by users
var (
	// GlobalHashFunction is the global hash function used by default
	GlobalHashFunction HashFunction = DefaultHashFunction{}

	// GlobalRandomnessSource is the global randomness source used by default
	GlobalRandomnessSource RandomnessSource = DefaultRandomnessSource{}
)
