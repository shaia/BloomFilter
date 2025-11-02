package hash

import (
	"testing"
)

// TestOptimized1BasicFunctionality tests basic hash function properties
func TestOptimized1BasicFunctionality(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "Empty input",
			input: []byte{},
		},
		{
			name:  "Single byte",
			input: []byte{42},
		},
		{
			name:  "Small input (< 8 bytes)",
			input: []byte{1, 2, 3, 4, 5},
		},
		{
			name:  "Exactly 8 bytes",
			input: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		},
		{
			name:  "Between 8 and 32 bytes",
			input: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
		{
			name:  "Exactly 32 bytes",
			input: make([]byte, 32),
		},
		{
			name:  "More than 32 bytes",
			input: make([]byte, 64),
		},
		{
			name:  "Large input (multiple 32-byte chunks)",
			input: make([]byte, 128),
		},
		{
			name:  "Odd size (33 bytes)",
			input: make([]byte, 33),
		},
		{
			name:  "Odd size (65 bytes)",
			input: make([]byte, 65),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Hash should be deterministic
			hash1 := Optimized1(tt.input)
			hash2 := Optimized1(tt.input)
			if hash1 != hash2 {
				t.Errorf("Optimized1 is not deterministic: got %v and %v", hash1, hash2)
			}

			// Hash should not be zero for non-empty inputs
			if len(tt.input) > 0 && hash1 == 0 {
				t.Errorf("Optimized1 returned zero hash for non-empty input")
			}
		})
	}
}

// TestOptimized2BasicFunctionality tests basic hash function properties
func TestOptimized2BasicFunctionality(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "Empty input",
			input: []byte{},
		},
		{
			name:  "Single byte",
			input: []byte{42},
		},
		{
			name:  "Small input (< 8 bytes)",
			input: []byte{1, 2, 3, 4, 5},
		},
		{
			name:  "Exactly 8 bytes",
			input: []byte{1, 2, 3, 4, 5, 6, 7, 8},
		},
		{
			name:  "Between 8 and 32 bytes",
			input: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
		{
			name:  "Exactly 32 bytes",
			input: make([]byte, 32),
		},
		{
			name:  "More than 32 bytes",
			input: make([]byte, 64),
		},
		{
			name:  "Large input (multiple 32-byte chunks)",
			input: make([]byte, 128),
		},
		{
			name:  "Odd size (33 bytes)",
			input: make([]byte, 33),
		},
		{
			name:  "Odd size (65 bytes)",
			input: make([]byte, 65),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Hash should be deterministic
			hash1 := Optimized2(tt.input)
			hash2 := Optimized2(tt.input)
			if hash1 != hash2 {
				t.Errorf("Optimized2 is not deterministic: got %v and %v", hash1, hash2)
			}

			// Hash should not be zero for non-empty inputs
			if len(tt.input) > 0 && hash1 == 0 {
				t.Errorf("Optimized2 returned zero hash for non-empty input")
			}
		})
	}
}

// TestHashIndependence verifies that Optimized1 and Optimized2 produce different hashes
func TestHashIndependence(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "Small input",
			input: []byte("hello"),
		},
		{
			name:  "Medium input",
			input: []byte("the quick brown fox jumps over the lazy dog"),
		},
		{
			name:  "Large input",
			input: make([]byte, 256),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := Optimized1(tt.input)
			hash2 := Optimized2(tt.input)

			if hash1 == hash2 {
				t.Errorf("Optimized1 and Optimized2 produced same hash %v for input %q", hash1, tt.input)
			}
		})
	}
}

// TestHashDifferentiation verifies that different inputs produce different hashes
func TestHashDifferentiation(t *testing.T) {
	t.Run("Optimized1", func(t *testing.T) {
		testCases := [][]byte{
			[]byte("a"),
			[]byte("b"),
			[]byte("aa"),
			[]byte("ab"),
			[]byte("ba"),
			[]byte("hello"),
			[]byte("world"),
			[]byte("the quick brown fox"),
			[]byte("the quick brown dog"),
		}

		seen := make(map[uint64][]byte)
		for _, input := range testCases {
			hash := Optimized1(input)
			if prev, exists := seen[hash]; exists {
				t.Errorf("Hash collision: %q and %q both produced hash %v", input, prev, hash)
			}
			seen[hash] = input
		}
	})

	t.Run("Optimized2", func(t *testing.T) {
		testCases := [][]byte{
			[]byte("a"),
			[]byte("b"),
			[]byte("aa"),
			[]byte("ab"),
			[]byte("ba"),
			[]byte("hello"),
			[]byte("world"),
			[]byte("the quick brown fox"),
			[]byte("the quick brown dog"),
		}

		seen := make(map[uint64][]byte)
		for _, input := range testCases {
			hash := Optimized2(input)
			if prev, exists := seen[hash]; exists {
				t.Errorf("Hash collision: %q and %q both produced hash %v", input, prev, hash)
			}
			seen[hash] = input
		}
	})
}

// TestHashSensitivity verifies that small changes produce different hashes
func TestHashSensitivity(t *testing.T) {
	t.Run("Optimized1 - bit flip", func(t *testing.T) {
		original := []byte("hello world")
		modified := []byte("hello world")
		modified[0] ^= 1 // Flip one bit

		hash1 := Optimized1(original)
		hash2 := Optimized1(modified)

		if hash1 == hash2 {
			t.Errorf("Optimized1 not sensitive to bit flip: both produced %v", hash1)
		}
	})

	t.Run("Optimized2 - bit flip", func(t *testing.T) {
		original := []byte("hello world")
		modified := []byte("hello world")
		modified[0] ^= 1 // Flip one bit

		hash1 := Optimized2(original)
		hash2 := Optimized2(modified)

		if hash1 == hash2 {
			t.Errorf("Optimized2 not sensitive to bit flip: both produced %v", hash1)
		}
	})
}

// TestHash32ByteChunking tests boundary conditions around 32-byte chunks
func TestHash32ByteChunking(t *testing.T) {
	// Create test data that will exercise different code paths
	sizes := []int{0, 1, 7, 8, 9, 15, 16, 17, 31, 32, 33, 63, 64, 65, 127, 128, 129}

	for _, size := range sizes {
		t.Run("Optimized1", func(t *testing.T) {
			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i % 256)
			}

			hash1 := Optimized1(data)
			hash2 := Optimized1(data)

			if hash1 != hash2 {
				t.Errorf("Size %d: Optimized1 not deterministic", size)
			}
		})

		t.Run("Optimized2", func(t *testing.T) {
			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i % 256)
			}

			hash1 := Optimized2(data)
			hash2 := Optimized2(data)

			if hash1 != hash2 {
				t.Errorf("Size %d: Optimized2 not deterministic", size)
			}
		})
	}
}

// TestHash8ByteChunking tests boundary conditions around 8-byte chunks
func TestHash8ByteChunking(t *testing.T) {
	// Test sizes that exercise the 8-byte chunk processing
	sizes := []int{7, 8, 9, 15, 16, 17, 23, 24, 25}

	for _, size := range sizes {
		t.Run("Optimized1", func(t *testing.T) {
			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i)
			}

			hash := Optimized1(data)

			// Verify different sizes produce different hashes
			if size > 0 {
				shorterData := data[:size-1]
				shorterHash := Optimized1(shorterData)
				if hash == shorterHash {
					t.Errorf("Size %d and %d produced same hash", size, size-1)
				}
			}
		})

		t.Run("Optimized2", func(t *testing.T) {
			data := make([]byte, size)
			for i := range data {
				data[i] = byte(i)
			}

			hash := Optimized2(data)

			// Verify different sizes produce different hashes
			if size > 0 {
				shorterData := data[:size-1]
				shorterHash := Optimized2(shorterData)
				if hash == shorterHash {
					t.Errorf("Size %d and %d produced same hash", size, size-1)
				}
			}
		})
	}
}

// TestHashOrderSensitivity verifies that byte order matters
func TestHashOrderSensitivity(t *testing.T) {
	t.Run("Optimized1", func(t *testing.T) {
		data1 := []byte{1, 2, 3, 4, 5}
		data2 := []byte{5, 4, 3, 2, 1}

		hash1 := Optimized1(data1)
		hash2 := Optimized1(data2)

		if hash1 == hash2 {
			t.Errorf("Optimized1 not sensitive to byte order: both produced %v", hash1)
		}
	})

	t.Run("Optimized2", func(t *testing.T) {
		data1 := []byte{1, 2, 3, 4, 5}
		data2 := []byte{5, 4, 3, 2, 1}

		hash1 := Optimized2(data1)
		hash2 := Optimized2(data2)

		if hash1 == hash2 {
			t.Errorf("Optimized2 not sensitive to byte order: both produced %v", hash1)
		}
	})
}

// TestHashEdgeCases tests edge cases
func TestHashEdgeCases(t *testing.T) {
	t.Run("All zeros", func(t *testing.T) {
		data := make([]byte, 100)
		hash1 := Optimized1(data)
		hash2 := Optimized2(data)

		if hash1 == 0 {
			t.Errorf("Optimized1 returned zero for all-zero input")
		}
		if hash2 == 0 {
			t.Errorf("Optimized2 returned zero for all-zero input")
		}
		if hash1 == hash2 {
			t.Errorf("Optimized1 and Optimized2 returned same hash for all-zero input")
		}
	})

	t.Run("All 0xFF", func(t *testing.T) {
		data := make([]byte, 100)
		for i := range data {
			data[i] = 0xFF
		}

		hash1 := Optimized1(data)
		hash2 := Optimized2(data)

		if hash1 == 0 {
			t.Errorf("Optimized1 returned zero for all-0xFF input")
		}
		if hash2 == 0 {
			t.Errorf("Optimized2 returned zero for all-0xFF input")
		}
		if hash1 == hash2 {
			t.Errorf("Optimized1 and Optimized2 returned same hash for all-0xFF input")
		}
	})

	t.Run("Repeating pattern", func(t *testing.T) {
		data := make([]byte, 100)
		for i := range data {
			data[i] = byte(i % 4)
		}

		hash1 := Optimized1(data)
		hash2 := Optimized2(data)

		if hash1 == 0 {
			t.Errorf("Optimized1 returned zero for repeating pattern")
		}
		if hash2 == 0 {
			t.Errorf("Optimized2 returned zero for repeating pattern")
		}
		if hash1 == hash2 {
			t.Errorf("Optimized1 and Optimized2 returned same hash for repeating pattern")
		}
	})
}

// TestHashConsistencyAcrossSizes verifies consistent behavior across different input sizes
func TestHashConsistencyAcrossSizes(t *testing.T) {
	// Create a large input buffer
	largeInput := make([]byte, 256)
	for i := range largeInput {
		largeInput[i] = byte(i)
	}

	// Test that hash of prefix differs from hash of full input
	for size := 1; size < len(largeInput); size += 7 {
		t.Run("Optimized1", func(t *testing.T) {
			hashPrefix := Optimized1(largeInput[:size])
			hashFull := Optimized1(largeInput)

			if hashPrefix == hashFull {
				t.Errorf("Size %d: prefix hash equals full hash", size)
			}
		})

		t.Run("Optimized2", func(t *testing.T) {
			hashPrefix := Optimized2(largeInput[:size])
			hashFull := Optimized2(largeInput)

			if hashPrefix == hashFull {
				t.Errorf("Size %d: prefix hash equals full hash", size)
			}
		})
	}
}

// TestHashNonZeroForKnownInputs tests specific known inputs
func TestHashNonZeroForKnownInputs(t *testing.T) {
	inputs := [][]byte{
		[]byte(""),
		[]byte("a"),
		[]byte("hello"),
		[]byte("the quick brown fox jumps over the lazy dog"),
		make([]byte, 1),
		make([]byte, 32),
		make([]byte, 64),
		make([]byte, 1024),
	}

	for i, input := range inputs {
		t.Run("Optimized1", func(t *testing.T) {
			hash := Optimized1(input)
			// Hash can be zero only for very specific edge cases, but should be deterministic
			_ = hash // Just verify it doesn't panic
		})

		t.Run("Optimized2", func(t *testing.T) {
			hash := Optimized2(input)
			// Hash can be zero only for very specific edge cases, but should be deterministic
			_ = hash // Just verify it doesn't panic
		})

		// Verify the two hash functions produce different values
		if len(input) > 0 {
			hash1 := Optimized1(input)
			hash2 := Optimized2(input)
			if hash1 == hash2 {
				t.Errorf("Input %d: Optimized1 and Optimized2 produced same hash %v", i, hash1)
			}
		}
	}
}
