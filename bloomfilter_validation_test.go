package bloomfilter

import (
	"math"
	"strings"
	"testing"
)

// TestInputValidation_ZeroExpectedElements verifies panic on zero expected elements
func TestInputValidation_ZeroExpectedElements(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Expected panic for zero expected elements, but didn't panic")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("Expected string panic message, got %T: %v", r, r)
		}
		expectedMsg := "bloomfilter: expectedElements must be greater than 0"
		if msg != expectedMsg {
			t.Errorf("Expected panic message %q, got %q", expectedMsg, msg)
		}
		t.Logf("Correctly panicked with message: %s", msg)
	}()

	NewCacheOptimizedBloomFilter(0, 0.01)
	t.Fatal("Should not reach here - expected panic")
}

// TestInputValidation_NegativeFPR verifies panic on negative false positive rate
func TestInputValidation_NegativeFPR(t *testing.T) {
	testCases := []struct {
		name string
		fpr  float64
	}{
		{"Slightly negative", -0.01},
		{"Very negative", -1.0},
		{"Negative infinity", math.Inf(-1)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					t.Fatalf("Expected panic for FPR=%f, but didn't panic", tc.fpr)
				}
				msg, ok := r.(string)
				if !ok {
					t.Fatalf("Expected string panic message, got %T: %v", r, r)
				}
				if !strings.Contains(msg, "must be in range (0, 1)") {
					t.Errorf("Expected range error message, got: %s", msg)
				}
				t.Logf("Correctly panicked with message: %s", msg)
			}()

			NewCacheOptimizedBloomFilter(1000, tc.fpr)
			t.Fatal("Should not reach here - expected panic")
		})
	}
}

// TestInputValidation_ZeroFPR verifies panic on zero false positive rate
func TestInputValidation_ZeroFPR(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Expected panic for zero FPR, but didn't panic")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("Expected string panic message, got %T: %v", r, r)
		}
		if !strings.Contains(msg, "must be in range (0, 1)") {
			t.Errorf("Expected range error message, got: %s", msg)
		}
		t.Logf("Correctly panicked with message: %s", msg)
	}()

	NewCacheOptimizedBloomFilter(1000, 0.0)
	t.Fatal("Should not reach here - expected panic")
}

// TestInputValidation_FPRTooHigh verifies panic on FPR >= 1.0
func TestInputValidation_FPRTooHigh(t *testing.T) {
	testCases := []struct {
		name string
		fpr  float64
	}{
		{"Exactly 1.0", 1.0},
		{"Slightly above 1.0", 1.01},
		{"Much greater than 1.0", 2.0},
		{"Positive infinity", math.Inf(1)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					t.Fatalf("Expected panic for FPR=%f, but didn't panic", tc.fpr)
				}
				msg, ok := r.(string)
				if !ok {
					t.Fatalf("Expected string panic message, got %T: %v", r, r)
				}
				if !strings.Contains(msg, "must be in range (0, 1)") {
					t.Errorf("Expected range error message, got: %s", msg)
				}
				t.Logf("Correctly panicked with message: %s", msg)
			}()

			NewCacheOptimizedBloomFilter(1000, tc.fpr)
			t.Fatal("Should not reach here - expected panic")
		})
	}
}

// TestInputValidation_FPRTooHighForElements verifies panic when FPR is so high it results in zero bits
func TestInputValidation_FPRTooHighForElements(t *testing.T) {
	testCases := []struct {
		name     string
		elements uint64
		fpr      float64
	}{
		{"FPR 0.999999 for 1000 elements", 1000, 0.999999},
		{"FPR 0.99999 for 100 elements", 100, 0.99999},
		{"FPR 0.9999 for 10 elements", 10, 0.9999},
		{"FPR 0.999 for 1 element", 1, 0.999},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					t.Fatalf("Expected panic for elements=%d, FPR=%f, but didn't panic", tc.elements, tc.fpr)
				}
				msg, ok := r.(string)
				if !ok {
					t.Fatalf("Expected string panic message, got %T: %v", r, r)
				}
				if !strings.Contains(msg, "too high") || !strings.Contains(msg, "zero bits") {
					t.Errorf("Expected 'too high' and 'zero bits' in message, got: %s", msg)
				}
				t.Logf("Correctly panicked with message: %s", msg)
			}()

			NewCacheOptimizedBloomFilter(tc.elements, tc.fpr)
			t.Fatal("Should not reach here - expected panic")
		})
	}
}

// TestInputValidation_NaNFPR verifies panic on NaN false positive rate
func TestInputValidation_NaNFPR(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Expected panic for NaN FPR, but didn't panic")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("Expected string panic message, got %T: %v", r, r)
		}
		expectedMsg := "bloomfilter: falsePositiveRate cannot be NaN"
		if msg != expectedMsg {
			t.Errorf("Expected panic message %q, got %q", expectedMsg, msg)
		}
		t.Logf("Correctly panicked with message: %s", msg)
	}()

	NewCacheOptimizedBloomFilter(1000, math.NaN())
	t.Fatal("Should not reach here - expected panic")
}

// TestInputValidation_ValidInputs verifies valid inputs don't panic
func TestInputValidation_ValidInputs(t *testing.T) {
	testCases := []struct {
		name     string
		elements uint64
		fpr      float64
	}{
		{"Typical usage", 1000, 0.01},
		{"Low FPR", 10000, 0.001},
		{"Very low FPR", 1000, 0.0001},
		{"Extremely low FPR", 100, 0.0000001},
		{"High FPR", 1000, 0.1},
		{"Very high FPR", 1000, 0.5},
		{"Small elements", 1, 0.01},
		{"Large elements", 1000000000, 0.01},
		{"Minimum valid FPR", 1000, 0.000001},
		{"Reasonably high FPR", 1000, 0.9},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Valid input (elements=%d, fpr=%f) caused panic: %v",
						tc.elements, tc.fpr, r)
				}
			}()

			bf := NewCacheOptimizedBloomFilter(tc.elements, tc.fpr)
			if bf == nil {
				t.Fatal("NewCacheOptimizedBloomFilter returned nil for valid input")
			}

			// Verify basic functionality works
			bf.AddString("test")
			if !bf.ContainsString("test") {
				t.Error("Basic Add/Contains functionality failed")
			}

			t.Logf("Successfully created filter: elements=%d, fpr=%f, hashCount=%d",
				tc.elements, tc.fpr, bf.hashCount)
		})
	}
}

// TestInputValidation_BoundaryValues tests edge cases at boundaries
func TestInputValidation_BoundaryValues(t *testing.T) {
	t.Run("expectedElements = 1", func(t *testing.T) {
		bf := NewCacheOptimizedBloomFilter(1, 0.01)
		if bf == nil {
			t.Fatal("Failed to create filter with expectedElements=1")
		}
		bf.AddString("single")
		if !bf.ContainsString("single") {
			t.Error("Failed to find single element")
		}
	})

	t.Run("FPR just above 0", func(t *testing.T) {
		bf := NewCacheOptimizedBloomFilter(1000, 0.000001)
		if bf == nil {
			t.Fatal("Failed to create filter with very low FPR")
		}
		stats := bf.GetCacheStats()
		t.Logf("Very low FPR: hashCount=%d, bitCount=%d", stats.HashCount, stats.BitCount)
	})

	t.Run("FPR close to 1 (should panic)", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("Expected panic for FPR close to 1.0 (0.999999)")
			}
			msg, ok := r.(string)
			if !ok {
				t.Fatalf("Expected string panic message, got %T: %v", r, r)
			}
			if !strings.Contains(msg, "too high") || !strings.Contains(msg, "zero bits") {
				t.Errorf("Expected 'too high' and 'zero bits' in message, got: %s", msg)
			}
			t.Logf("Correctly panicked for extremely high FPR: %s", msg)
		}()

		NewCacheOptimizedBloomFilter(1000, 0.999999)
		t.Fatal("Should not reach here - expected panic")
	})

	t.Run("Very large expectedElements", func(t *testing.T) {
		bf := NewCacheOptimizedBloomFilter(1<<32, 0.01) // 4 billion elements
		if bf == nil {
			t.Fatal("Failed to create filter with very large expectedElements")
		}
		stats := bf.GetCacheStats()
		t.Logf("Large filter: cacheLines=%d, memory=%d MB",
			stats.CacheLineCount, stats.MemoryUsage/(1024*1024))
	})
}

// TestInputValidation_PanicRecovery verifies panic recovery and error messages
func TestInputValidation_PanicRecovery(t *testing.T) {
	// Test that we can recover from multiple panics in sequence
	invalidInputs := []struct {
		elements uint64
		fpr      float64
		desc     string
	}{
		{0, 0.01, "zero elements"},
		{1000, 0.0, "zero FPR"},
		{1000, -0.5, "negative FPR"},
		{1000, 1.0, "FPR = 1.0"},
		{1000, 1.5, "FPR > 1.0"},
		{1000, math.NaN(), "NaN FPR"},
	}

	for _, input := range invalidInputs {
		t.Run(input.desc, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Expected panic for %s (elements=%d, fpr=%f)",
						input.desc, input.elements, input.fpr)
				} else {
					t.Logf("Correctly panicked for %s: %v", input.desc, r)
				}
			}()

			NewCacheOptimizedBloomFilter(input.elements, input.fpr)
		})
	}

	// Verify we can still create valid filters after panics
	bf := NewCacheOptimizedBloomFilter(1000, 0.01)
	if bf == nil {
		t.Fatal("Failed to create valid filter after panic tests")
	}
}
