package gkr

// The following are small test cases in the sense that

// JSON Utils

type WireCoordinates struct {
	layer int
	pos   int
}

type WireJson struct {
	gate   string
	inputs []WireCoordinates
}

type CircuitJson [][]WireJson

type TestCase struct {
	hash    map[string]string // `json:"hash"`
	circuit CircuitJson
}
