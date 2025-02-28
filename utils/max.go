package utils

// Max returns the maximum of two integers
func Max(a, b int) int {
	// Deprecated: but keeping until next major release to avoid breaking some vendored code
	return max(a, b)
}
