package hash

// CompressionFunction is a 2 to 1 function
type CompressionFunction interface {
	Apply([]byte, []byte) ([]byte, error) // TODO @Tabaie @ThomasPiellard better name
	BlockSize() int
}
type merkleDamgardHasher struct {
	state []byte
	iv    []byte
	f     CompressionFunction
}

// Write implements hash.Write
func (h *merkleDamgardHasher) Write(p []byte) (n int, err error) {
	blockSize := h.f.BlockSize()
	for len(p) != 0 {
		if len(p) < blockSize {
			p = append(make([]byte, blockSize-len(p), blockSize), p...)
		}
		if h.state, err = h.f.Apply(h.state, p[:blockSize]); err != nil {
			return
		}
		n += blockSize
		p = p[blockSize:]
	}
	return
}

func (h *merkleDamgardHasher) Sum(b []byte) []byte {
	if _, err := h.Write(b); err != nil {
		panic(err)
	}
	return h.state
}

func (h *merkleDamgardHasher) Reset() {
	h.state = h.iv
}

func (h *merkleDamgardHasher) Size() int {
	return h.f.BlockSize()
}

func (h *merkleDamgardHasher) BlockSize() int {
	return h.f.BlockSize()
}

func (h *merkleDamgardHasher) State() []byte {
	return h.state
}

func (h *merkleDamgardHasher) SetState(state []byte) error {
	h.state = state
	return nil
}

// NewMerkleDamgardHasher transforms a 2-1 one-way function into a hash
// initialState is a value whose preimage is not known
// WARNING: The padding performed by the resulting hasher is trivial.
// It simply left zero-pads the last block of input
// THIS IS NOT COLLISION RESISTANT FOR GENERIC DATA
func NewMerkleDamgardHasher(f CompressionFunction, initialState []byte) StateStorer {
	return &merkleDamgardHasher{
		state: initialState,
		iv:    initialState,
		f:     f,
	}
}
