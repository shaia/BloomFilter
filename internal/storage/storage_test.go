package storage

import (
	"testing"
)

// TestNewArrayMode verifies array mode initialization
func TestNewArrayMode(t *testing.T) {
	// Test array mode (below threshold)
	s := New(5000, 10, 10000)

	if !s.UseArrayMode {
		t.Errorf("Expected array mode for 5000 cache lines (threshold: 10000)")
	}
}

// TestNewMapMode verifies map mode initialization
func TestNewMapMode(t *testing.T) {
	// Test map mode (above threshold)
	s := New(15000, 10, 10000)

	if s.UseArrayMode {
		t.Errorf("Expected map mode for 15000 cache lines (threshold: 10000)")
	}
}

// TestThresholdBoundary verifies behavior at the threshold boundary
func TestThresholdBoundary(t *testing.T) {
	threshold := uint64(10000)

	tests := []struct {
		name        string
		cacheLines  uint64
		expectArray bool
	}{
		{"Just below threshold", threshold - 1, true},
		{"At threshold", threshold, true},
		{"Just above threshold", threshold + 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(tt.cacheLines, 10, threshold)
			if s.UseArrayMode != tt.expectArray {
				t.Errorf("%s: expected array mode=%v, got=%v",
					tt.name, tt.expectArray, s.UseArrayMode)
			}
		})
	}
}

// TestOperationStoragePool tests the sync.Pool functionality
func TestOperationStoragePool(t *testing.T) {
	// Test array mode pool
	ops1 := GetOperationStorage(true)
	if !ops1.UseArrayMode {
		t.Error("Array mode operation storage should have UseArrayMode=true")
	}
	if ops1.ArrayOps == nil {
		t.Error("Array mode operation storage should have ArrayOps initialized")
	}
	PutOperationStorage(ops1)

	// Test map mode pool
	ops2 := GetOperationStorage(false)
	if ops2.UseArrayMode {
		t.Error("Map mode operation storage should have UseArrayMode=false")
	}
	if ops2.MapOps == nil {
		t.Error("Map mode operation storage should have MapOps initialized")
	}
	PutOperationStorage(ops2)

	// Test that pool reuses objects
	ops3 := GetOperationStorage(true)
	if ops3 == nil {
		t.Error("Pool should return valid operation storage")
	}
	PutOperationStorage(ops3)
}

// TestGetOperations tests getting operations through the API
func TestGetOperations(t *testing.T) {
	modes := []bool{true, false} // array mode, map mode

	for _, useArrayMode := range modes {
		modeName := "Map mode"
		if useArrayMode {
			modeName = "Array mode"
		}

		t.Run(modeName, func(t *testing.T) {
			ops := GetOperationStorage(useArrayMode)
			defer PutOperationStorage(ops)

			// Add operations
			ops.AddGetOperation(42, 1, 5)
			ops.AddGetOperation(42, 2, 10)

			// Retrieve operations
			operations := ops.GetGetOperations(42)

			if len(operations) != 2 {
				t.Errorf("Expected 2 operations, got %d", len(operations))
			}

			// Verify operation details
			if operations[0].WordIdx != 1 || operations[0].BitOffset != 5 {
				t.Errorf("First operation incorrect: got WordIdx=%d, BitOffset=%d",
					operations[0].WordIdx, operations[0].BitOffset)
			}

			if operations[1].WordIdx != 2 || operations[1].BitOffset != 10 {
				t.Errorf("Second operation incorrect: got WordIdx=%d, BitOffset=%d",
					operations[1].WordIdx, operations[1].BitOffset)
			}
		})
	}
}

// TestClearGetMap tests clearing get operations
func TestClearGetMap(t *testing.T) {
	modes := []bool{true, false}

	for _, useArrayMode := range modes {
		modeName := "Map mode"
		if useArrayMode {
			modeName = "Array mode"
		}

		t.Run(modeName, func(t *testing.T) {
			ops := GetOperationStorage(useArrayMode)
			defer PutOperationStorage(ops)

			// Add some data
			ops.AddGetOperation(10, 1, 2)
			ops.AddGetOperation(20, 3, 4)

			// Verify data was added
			ops10 := ops.GetGetOperations(10)
			if len(ops10) != 1 {
				t.Errorf("Expected 1 operation at index 10, got %d", len(ops10))
			}

			ops.ClearGetMap()

			// Verify cleared
			ops10After := ops.GetGetOperations(10)
			if len(ops10After) != 0 {
				t.Error("GetOperations[10] should be cleared")
			}
			if len(ops.UsedIndicesGet) != 0 {
				t.Error("UsedIndicesGet should be cleared")
			}
		})
	}
}

// TestMultipleOperations tests multiple operations in both modes
func TestMultipleOperations(t *testing.T) {
	modes := []bool{true, false}

	for _, useArrayMode := range modes {
		modeName := "Map mode"
		if useArrayMode {
			modeName = "Array mode"
		}

		t.Run(modeName, func(t *testing.T) {
			ops := GetOperationStorage(useArrayMode)
			defer PutOperationStorage(ops)

			// Add multiple operations
			for i := uint64(0); i < 100; i++ {
				ops.AddGetOperation(i, i, i%64)
			}

			// Verify all operations exist
			for i := uint64(0); i < 100; i++ {
				operations := ops.GetGetOperations(i)
				if len(operations) != 1 {
					t.Errorf("Cache line %d: expected 1 op, got %d", i, len(operations))
				}
			}

			// Clear and verify
			ops.ClearGetMap()

			for i := uint64(0); i < 100; i++ {
				operations := ops.GetGetOperations(i)
				if len(operations) != 0 {
					t.Errorf("After clear, cache line %d should have 0 ops, got %d", i, len(operations))
				}
			}
		})
	}
}

// TestAddHashPosition tests hash position tracking
func TestAddHashPosition(t *testing.T) {
	modes := []bool{true, false}

	for _, useArrayMode := range modes {
		modeName := "Map mode"
		if useArrayMode {
			modeName = "Array mode"
		}

		t.Run(modeName, func(t *testing.T) {
			ops := GetOperationStorage(useArrayMode)
			defer PutOperationStorage(ops)

			// Add hash positions
			ops.AddHashPosition(42, 100)
			ops.AddHashPosition(42, 200)
			ops.AddHashPosition(43, 300)

			// Verify used indices
			usedIndices := ops.GetUsedHashIndices()
			if len(usedIndices) == 0 {
				t.Error("Expected used hash indices to be tracked")
			}

			// Clear and verify
			ops.ClearHashMap()
			usedIndicesAfter := ops.GetUsedHashIndices()
			if len(usedIndicesAfter) != 0 {
				t.Error("Used hash indices should be cleared")
			}
		})
	}
}

// TestSetOperations tests set operation tracking
func TestSetOperations(t *testing.T) {
	modes := []bool{true, false}

	for _, useArrayMode := range modes {
		modeName := "Map mode"
		if useArrayMode {
			modeName = "Array mode"
		}

		t.Run(modeName, func(t *testing.T) {
			ops := GetOperationStorage(useArrayMode)
			defer PutOperationStorage(ops)

			// Add set operations
			ops.AddSetOperation(10, 1, 5)
			ops.AddSetOperation(10, 2, 10)
			ops.AddSetOperation(20, 3, 15)

			// Verify operations were added
			ops10 := ops.GetSetOperations(10)
			if len(ops10) != 2 {
				t.Errorf("Expected 2 set operations at index 10, got %d", len(ops10))
			}

			ops20 := ops.GetSetOperations(20)
			if len(ops20) != 1 {
				t.Errorf("Expected 1 set operation at index 20, got %d", len(ops20))
			}

			// Verify used indices
			usedIndices := ops.GetUsedSetIndices()
			if len(usedIndices) == 0 {
				t.Error("Expected used set indices to be tracked")
			}

			// Clear and verify
			ops.ClearSetMap()
			ops10After := ops.GetSetOperations(10)
			if len(ops10After) != 0 {
				t.Error("Set operations should be cleared")
			}
		})
	}
}

// TestConcurrentPoolAccess tests that the pool is safe for concurrent access
func TestConcurrentPoolAccess(t *testing.T) {
	const numGoroutines = 100
	const numOperationsPerGoroutine = 1000

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numOperationsPerGoroutine; j++ {
				// Alternate between array and map mode
				useArrayMode := j%2 == 0
				ops := GetOperationStorage(useArrayMode)

				// Do some operations
				ops.AddGetOperation(uint64(j), uint64(j), uint64(j%64))
				_ = ops.GetUsedGetIndices()

				PutOperationStorage(ops)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
