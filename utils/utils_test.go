package utils

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMaxIndexFunc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		n        int
		gt       func(int, int) bool
		expected int
	}{
		{
			name: "max value",
			n:    10,
			gt: func(i, j int) bool {
				return i > j
			},
			expected: 9,
		},
		{
			name: "min value",
			n:    10,
			gt: func(i, j int) bool {
				return i < j
			},
			expected: 0,
		},
		{
			name: "even indices",
			n:    10,
			gt: func(i, j int) bool {
				return i%2 == 0 && (j%2 != 0 || i > j)
			},
			expected: 8,
		},
		{
			name: "single element",
			n:    1,
			gt: func(i, j int) bool {
				return i > j
			},
			expected: 0,
		},
		{
			name: "empty range",
			n:    0,
			gt: func(i, j int) bool {
				return i > j
			},
			expected: 0,
		},
	}

	for _, test := range tests {
		test := test // capture range variable
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result := MaxIndexFunc(test.n, test.gt)
			if result != test.expected {
				t.Errorf("MaxIndexFunc(%d, func): expected %d, got %d", test.n, test.expected, result)
			}
		})
	}
}

func TestPartialSums(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "empty slice",
			input:    []int{},
			expected: nil,
		},
		{
			name:     "singleton",
			input:    []int{5},
			expected: []int{5},
		},
		{
			name:     "consecutive integers",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 3, 6, 10, 15},
		},
		{
			name:     "with negative numbers",
			input:    []int{-1, 2, -3, 4},
			expected: []int{-1, 1, -2, 2},
		},
		{
			name:     "large values",
			input:    []int{1000000, 2000000, 3000000},
			expected: []int{1000000, 3000000, 6000000},
		},
	}

	for _, test := range tests {
		test := test // capture range variable
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result := PartialSums(test.input...)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("PartialSums(%v): expected %v, got %v", test.input, test.expected, result)
			}
		})
	}
}

func TestPartialSumsF(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		n        int
		f        func(int) int
		expected []int
	}{
		{
			name:     "empty range",
			n:        0,
			f:        func(i int) int { return i },
			expected: nil,
		},
		{
			name:     "identity function",
			n:        5,
			f:        func(i int) int { return i },
			expected: []int{0, 1, 3, 6, 10},
		},
		{
			name:     "square function",
			n:        5,
			f:        func(i int) int { return i * i },
			expected: []int{0, 1, 5, 14, 30},
		},
		{
			name:     "constant function",
			n:        4,
			f:        func(i int) int { return 2 },
			expected: []int{2, 4, 6, 8},
		},
	}

	for _, test := range tests {
		test := test // capture range variable
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result := PartialSumsF(test.n, test.f)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("PartialSumsF(%d, func): expected %v, got %v", test.n, test.expected, result)
			}
		})
	}
}

func TestMax(t *testing.T) {
	t.Parallel()

	tests := []struct {
		a, b     int
		expected int
	}{
		{1, 2, 2},
		{3, 2, 3},
		{-1, -2, -1},
		{0, 0, 0},
		{-10, 5, 5},
		{math.MaxInt32, math.MaxInt32 - 1, math.MaxInt32},
	}

	for i, test := range tests {
		i, test := i, test // capture range variable
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			t.Parallel()
			result := Max(test.a, test.b)
			if result != test.expected {
				t.Errorf("Max(%d, %d): expected %d, got %d", test.a, test.b, test.expected, result)
			}
		})
	}
}

func TestWorkerPool(t *testing.T) {
	t.Parallel()

	t.Run("basic functionality", func(t *testing.T) {
		t.Parallel()

		pool := NewWorkerPool()
		defer pool.Stop()

		n := 1000
		result := make([]int, n)
		expected := make([]int, n)

		for i := 0; i < n; i++ {
			expected[i] = i * 2
		}

		wg := pool.Submit(n, func(start, end int) {
			for i := start; i < end; i++ {
				result[i] = i * 2
			}
		}, 100)
		wg.Wait()

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("WorkerPool computation results don't match expected values")
		}
	})

	t.Run("worker count", func(t *testing.T) {
		t.Parallel()

		pool := NewWorkerPool()
		defer pool.Stop()

		// Ensure NbWorkers returns a reasonable number (at least 1)
		if pool.NbWorkers() < 1 {
			t.Errorf("NbWorkers(): expected at least 1, got %d", pool.NbWorkers())
		}

		// Should be related to runtime.NumCPU()
		if pool.NbWorkers() < runtime.NumCPU() {
			t.Logf("NbWorkers() returned %d which is less than NumCPU() = %d, this is just informational",
				pool.NbWorkers(), runtime.NumCPU())
		}
	})

	t.Run("concurrent tasks", func(t *testing.T) {
		t.Parallel()

		pool := NewWorkerPool()
		defer pool.Stop()

		var counter int64
		var wg sync.WaitGroup

		// Create several concurrent tasks that increment the counter
		numTasks := 5
		numItemsPerTask := 100

		for i := 0; i < numTasks; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				localWg := pool.Submit(numItemsPerTask, func(start, end int) {
					for i := start; i < end; i++ {
						atomic.AddInt64(&counter, 1)
					}
				}, 10)
				localWg.Wait()
			}()
		}

		wg.Wait()
		expected := int64(numTasks * numItemsPerTask)
		if counter != expected {
			t.Errorf("Counter: expected %d, got %d", expected, counter)
		}
	})

	t.Run("stress test", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping stress test in short mode")
		}
		t.Parallel()

		pool := NewWorkerPool()
		defer pool.Stop()

		var counter int64
		var wg sync.WaitGroup
		numGoroutines := 20
		tasksPerGoroutine := 10
		itemsPerTask := 1000

		// Start multiple goroutines that submit tasks to the worker pool
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < tasksPerGoroutine; j++ {
					localWg := pool.Submit(itemsPerTask, func(start, end int) {
						// Simulate some work
						time.Sleep(time.Microsecond)
						atomic.AddInt64(&counter, int64(end-start))
					}, itemsPerTask/10)
					localWg.Wait()
				}
			}(i)
		}

		wg.Wait()
		expected := int64(numGoroutines * tasksPerGoroutine * itemsPerTask)
		if counter != expected {
			t.Errorf("Stress test counter: expected %d, got %d", expected, counter)
		}
	})
}

func TestToSuperscript(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"123", "¹²³"},
		{"abc", "ᵃᵇᶜ"},
		{"xyz", "ˣʸᶻ"},
		{"H2O", "ᴴ²ᴼ"},
		{"m2/s2", "ᵐ²/ˢ²"},
		{"e=mc2", "ᵉ⁼ᵐᶜ²"},
		{"ABC", "ᴬᴮC"},
		{"_!?", "_!?"}, // Characters without superscript versions should remain unchanged
		{"", ""},       // Empty string
		{" ", " "},     // Space
	}

	for _, test := range tests {
		test := test // capture range variable
		t.Run(fmt.Sprintf("string_%s", test.input), func(t *testing.T) {
			t.Parallel()
			result := ToSuperscript(test.input)
			if result != test.expected {
				t.Errorf("ToSuperscript(%q): expected %q, got %q", test.input, test.expected, result)
			}
		})
	}
}

func TestToSubscript(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"123", "₁₂₃"},
		{"0123456789", "₀₁₂₃₄₅₆₇₈₉"},
		{"H2O", "H₂O"},
		{"CO2", "CO₂"},
		{"a1b2c3", "ₐ₁b₂c₃"},
		{"x+y=z", "ₓ₊y₌z"},
		{"(n+1)", "₍ₙ₊₁₎"},
		{"_!?", "_!?"}, // Characters without subscript versions should remain unchanged
		{"", ""},       // Empty string
		{" ", " "},     // Space
	}

	for _, test := range tests {
		test := test // capture range variable
		t.Run(fmt.Sprintf("string_%s", test.input), func(t *testing.T) {
			t.Parallel()
			result := ToSubscript(test.input)
			if result != test.expected {
				t.Errorf("ToSubscript(%q): expected %q, got %q", test.input, test.expected, result)
			}
		})
	}
}

func TestSupAndSubFunctions(t *testing.T) {
	t.Parallel()

	t.Run("sup function", func(t *testing.T) {
		t.Parallel()
		// Test characters that have superscript versions
		for input, expected := range map[rune]rune{
			'0': '⁰', '1': '¹', '2': '²', '3': '³', '4': '⁴',
			'5': '⁵', '6': '⁶', '7': '⁷', '8': '⁸', '9': '⁹',
			'a': 'ᵃ', 'b': 'ᵇ',
			'A': 'ᴬ', 'B': 'ᴮ',
		} {
			result, err := sup(input)
			if err != nil {
				t.Errorf("sup(%q) returned unexpected error: %v", input, err)
			}
			if result != expected {
				t.Errorf("sup(%q): expected %q, got %q", input, expected, result)
			}
		}

		// Test characters without superscript versions
		for _, input := range []rune{'$', '@', '#', '£', '€', '¥', 'Ω', 'C', '-'} {
			result, err := sup(input)
			if err == nil {
				t.Errorf("sup(%q) should return an error", input)
			}
			if result != input {
				t.Errorf("sup(%q): expected %q (unchanged), got %q", input, input, result)
			}
			if !strings.Contains(err.Error(), "no corresponding superscript") {
				t.Errorf("Unexpected error message for sup(%q): %v", input, err)
			}
		}
	})

	t.Run("sub function", func(t *testing.T) {
		t.Parallel()
		// Test characters that have subscript versions
		for input, expected := range map[rune]rune{
			'0': '₀', '1': '₁', '2': '₂', '3': '₃', '4': '₄',
			'5': '₅', '6': '₆', '7': '₇', '8': '₈', '9': '₉',
			'a': 'ₐ', 'e': 'ₑ', 'o': 'ₒ',
			'+': '₊', '=': '₌', '(': '₍', ')': '₎',
		} {
			result, err := sub(input)
			if err != nil {
				t.Errorf("sub(%q) returned unexpected error: %v", input, err)
			}
			if result != expected {
				t.Errorf("sub(%q): expected %q, got %q", input, expected, result)
			}
		}

		// Test characters without subscript versions
		for _, input := range []rune{'$', '@', '#', '£', '€', '¥', 'Ω', 'A', 'B', 'D', '-'} {
			result, err := sub(input)
			if err == nil {
				t.Errorf("sub(%q) should return an error", input)
			}
			if result != input {
				t.Errorf("sub(%q): expected %q (unchanged), got %q", input, input, result)
			}
			if !strings.Contains(err.Error(), "no corresponding subscript") {
				t.Errorf("Unexpected error message for sub(%q): %v", input, err)
			}
		}
	})

	// Test for non-ASCII characters
	t.Run("unicode characters", func(t *testing.T) {
		t.Parallel()
		// Unicode characters from different scripts
		unicodeChars := []rune{'α', 'β', 'γ', 'δ', 'א', 'ב', 'ג', 'ד', '你', '我', '他', '她'}

		for _, char := range unicodeChars {
			// For superscript
			result, err := sup(char)
			// Either it has a mapping or it returns the original with an error
			if err == nil {
				t.Logf("Character %q has a superscript mapping to %q", char, result)
			} else {
				if result != char {
					t.Errorf("When sup(%q) returns an error, it should return the original character", char)
				}
			}

			// For subscript
			result, err = sub(char)
			// Either it has a mapping or it returns the original with an error
			if err == nil {
				t.Logf("Character %q has a subscript mapping to %q", char, result)
			} else {
				if result != char {
					t.Errorf("When sub(%q) returns an error, it should return the original character", char)
				}
			}
		}
	})
}

// BenchmarkWorkerPool measures the performance of the worker pool implementation
func BenchmarkWorkerPool(b *testing.B) {
	pool := NewWorkerPool()
	defer pool.Stop()

	b.Run("small tasks", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			wg := pool.Submit(1000, func(start, end int) {
				for j := start; j < end; j++ {
					_ = j * j // Simple computation
				}
			}, 100)
			wg.Wait()
		}
	})

	b.Run("medium tasks", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			wg := pool.Submit(10000, func(start, end int) {
				for j := start; j < end; j++ {
					_ = j * j // Simple computation
				}
			}, 1000)
			wg.Wait()
		}
	})
}

// BenchmarkToSuperscript measures the performance of the ToSuperscript function
func BenchmarkToSuperscript(b *testing.B) {
	testStrings := []string{
		"0123456789",
		"abcdefghijklmnopqrstuvwxyz",
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"This is a longer string with mixed characters 123!",
	}

	for _, str := range testStrings {
		b.Run(fmt.Sprintf("string_len_%d", len(str)), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = ToSuperscript(str)
			}
		})
	}
}

// BenchmarkToSubscript measures the performance of the ToSubscript function
func BenchmarkToSubscript(b *testing.B) {
	testStrings := []string{
		"0123456789",
		"abcdefghijklmnopqrstuvwxyz",
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"This is a longer string with mixed characters 123!",
	}

	for _, str := range testStrings {
		b.Run(fmt.Sprintf("string_len_%d", len(str)), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = ToSubscript(str)
			}
		})
	}
}
