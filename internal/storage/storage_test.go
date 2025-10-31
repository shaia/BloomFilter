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

	// Verify array structures are initialized
	if s.ArrayOps == nil {
		t.Error("ArrayOps should be initialized in array mode")
	}
	if s.ArrayOpsSet == nil {
		t.Error("ArrayOpsSet should be initialized in array mode")
	}
	if s.ArrayMap == nil {
		t.Error("ArrayMap should be initialized in array mode")
	}

	// Verify map structures are nil
	if s.MapOps != nil {
		t.Error("MapOps should be nil in array mode")
	}
	if s.MapOpsSet != nil {
		t.Error("MapOpsSet should be nil in array mode")
	}
	if s.MapMap != nil {
		t.Error("MapMap should be nil in array mode")
	}
}

// TestNewMapMode verifies map mode initialization
func TestNewMapMode(t *testing.T) {
	// Test map mode (above threshold)
	s := New(15000, 10, 10000)

	if s.UseArrayMode {
		t.Errorf("Expected map mode for 15000 cache lines (threshold: 10000)")
	}

	// Verify map structures are initialized
	if s.MapOps == nil {
		t.Error("MapOps should be initialized in map mode")
	}
	if s.MapOpsSet == nil {
		t.Error("MapOpsSet should be initialized in map mode")
	}
	if s.MapMap == nil {
		t.Error("MapMap should be initialized in map mode")
	}

	// Verify array structures are nil
	if s.ArrayOps != nil {
		t.Error("ArrayOps should be nil in map mode")
	}
	if s.ArrayOpsSet != nil {
		t.Error("ArrayOpsSet should be nil in map mode")
	}
	if s.ArrayMap != nil {
		t.Error("ArrayMap should be nil in map mode")
	}
}

// TestThresholdBoundary verifies behavior at the threshold boundary
func TestThresholdBoundary(t *testing.T) {
	threshold := uint64(10000)

	tests := []struct {
		name          string
		cacheLines    uint64
		expectArray   bool
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

// TestUsedIndicesInitialization verifies used indices tracking is initialized
func TestUsedIndicesInitialization(t *testing.T) {
	hashCount := uint32(10)
	s := New(5000, hashCount, 10000)

	// Verify slices are initialized with appropriate capacity
	if s.UsedIndicesGet == nil {
		t.Error("UsedIndicesGet should be initialized")
	}
	if s.UsedIndicesSet == nil {
		t.Error("UsedIndicesSet should be initialized")
	}
	if s.UsedIndicesHash == nil {
		t.Error("UsedIndicesHash should be initialized")
	}

	// Verify they start empty
	if len(s.UsedIndicesGet) != 0 {
		t.Errorf("UsedIndicesGet should start empty, got length %d", len(s.UsedIndicesGet))
	}
	if len(s.UsedIndicesSet) != 0 {
		t.Errorf("UsedIndicesSet should start empty, got length %d", len(s.UsedIndicesSet))
	}
	if len(s.UsedIndicesHash) != 0 {
		t.Errorf("UsedIndicesHash should start empty, got length %d", len(s.UsedIndicesHash))
	}
}

// TestGetOperations tests getting operations through the API
func TestGetOperations(t *testing.T) {
	modes := []struct {
		name       string
		cacheLines uint64
	}{
		{"Array mode", 5000},
		{"Map mode", 15000},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			s := New(mode.cacheLines, 10, 10000)

			// Add operations
			s.AddGetOperation(42, 1, 5)
			s.AddGetOperation(42, 2, 10)

			// Retrieve operations
			ops := s.GetGetOperations(42)

			if len(ops) != 2 {
				t.Errorf("Expected 2 operations, got %d", len(ops))
			}

			// Verify operation details
			if ops[0].WordIdx != 1 || ops[0].BitOffset != 5 {
				t.Errorf("First operation incorrect: got WordIdx=%d, BitOffset=%d",
					ops[0].WordIdx, ops[0].BitOffset)
			}

			if ops[1].WordIdx != 2 || ops[1].BitOffset != 10 {
				t.Errorf("Second operation incorrect: got WordIdx=%d, BitOffset=%d",
					ops[1].WordIdx, ops[1].BitOffset)
			}
		})
	}
}

// TestClearGetMapArrayMode tests clearing get operations in array mode
func TestClearGetMapArrayMode(t *testing.T) {
	s := New(5000, 10, 10000)

	// Add some data using the proper API
	s.AddGetOperation(10, 1, 2)
	s.AddGetOperation(20, 3, 4)

	// Verify data was added
	ops10 := s.GetGetOperations(10)
	if len(ops10) != 1 {
		t.Errorf("Expected 1 operation at index 10, got %d", len(ops10))
	}

	s.ClearGetMap()

	// Verify cleared
	ops10After := s.GetGetOperations(10)
	if len(ops10After) != 0 {
		t.Error("GetOperations[10] should be cleared")
	}
	if len(s.UsedIndicesGet) != 0 {
		t.Error("UsedIndicesGet should be cleared")
	}
}

// TestClearGetMapMapMode tests clearing get operations in map mode
func TestClearGetMapMapMode(t *testing.T) {
	s := New(15000, 10, 10000)

	// Add some data using the proper API
	s.AddGetOperation(10, 1, 2)
	s.AddGetOperation(20, 3, 4)

	// Verify data was added
	ops10 := s.GetGetOperations(10)
	if len(ops10) != 1 {
		t.Errorf("Expected 1 operation at index 10, got %d", len(ops10))
	}

	s.ClearGetMap()

	// Verify cleared - map should be recreated
	if len(s.MapOps) != 0 {
		t.Errorf("MapOps should be empty after clear, got %d entries", len(s.MapOps))
	}
}

// TestMultipleOperations tests multiple operations in both modes
func TestMultipleOperations(t *testing.T) {
	modes := []struct {
		name       string
		cacheLines uint64
	}{
		{"Array mode", 5000},
		{"Map mode", 15000},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			s := New(mode.cacheLines, 10, 10000)

			// Add multiple operations using the proper API
			for i := uint64(0); i < 100; i++ {
				s.AddGetOperation(i, i, i%64)
			}

			// Verify all operations exist
			for i := uint64(0); i < 100; i++ {
				ops := s.GetGetOperations(i)
				if len(ops) != 1 {
					t.Errorf("Cache line %d: expected 1 op, got %d", i, len(ops))
				}
			}

			// Clear and verify
			s.ClearGetMap()

			for i := uint64(0); i < 100; i++ {
				ops := s.GetGetOperations(i)
				if len(ops) != 0 {
					t.Errorf("After clear, cache line %d should have 0 ops, got %d", i, len(ops))
				}
			}
		})
	}
}

// TestAddHashPosition tests hash position tracking
func TestAddHashPosition(t *testing.T) {
	modes := []struct {
		name       string
		cacheLines uint64
	}{
		{"Array mode", 5000},
		{"Map mode", 15000},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			s := New(mode.cacheLines, 10, 10000)

			// Add hash positions
			s.AddHashPosition(42, 100)
			s.AddHashPosition(42, 200)
			s.AddHashPosition(43, 300)

			// Verify used indices
			usedIndices := s.GetUsedHashIndices()
			if len(usedIndices) == 0 {
				t.Error("Expected used hash indices to be tracked")
			}

			// Clear and verify
			s.ClearHashMap()
			usedIndicesAfter := s.GetUsedHashIndices()
			if len(usedIndicesAfter) != 0 {
				t.Error("Used hash indices should be cleared")
			}
		})
	}
}

// TestSetOperations tests set operation tracking
func TestSetOperations(t *testing.T) {
	modes := []struct {
		name       string
		cacheLines uint64
	}{
		{"Array mode", 5000},
		{"Map mode", 15000},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			s := New(mode.cacheLines, 10, 10000)

			// Add set operations
			s.AddSetOperation(10, 1, 5)
			s.AddSetOperation(10, 2, 10)
			s.AddSetOperation(20, 3, 15)

			// Verify operations were added
			ops10 := s.GetSetOperations(10)
			if len(ops10) != 2 {
				t.Errorf("Expected 2 set operations at index 10, got %d", len(ops10))
			}

			ops20 := s.GetSetOperations(20)
			if len(ops20) != 1 {
				t.Errorf("Expected 1 set operation at index 20, got %d", len(ops20))
			}

			// Verify used indices
			usedIndices := s.GetUsedSetIndices()
			if len(usedIndices) == 0 {
				t.Error("Expected used set indices to be tracked")
			}

			// Clear and verify
			s.ClearSetMap()
			ops10After := s.GetSetOperations(10)
			if len(ops10After) != 0 {
				t.Error("Set operations should be cleared")
			}
		})
	}
}
