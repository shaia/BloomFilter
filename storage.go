package bloomfilter

// storageMode handles the hybrid array/map storage abstraction.
// This encapsulates the logic for choosing between array mode (small filters)
// and map mode (large filters) without duplicating code.
type storageMode struct {
	useArrayMode bool

	// Array-based storage (for small filters, zero-overhead indexing)
	arrayOps    *[ArrayModeThreshold][]opDetail
	arrayOpsSet *[ArrayModeThreshold][]struct{ wordIdx, bitOffset uint64 }
	arrayMap    *[ArrayModeThreshold][]uint64

	// Map-based storage (for large filters, dynamic scaling)
	mapOps    map[uint64][]opDetail
	mapOpsSet map[uint64][]struct{ wordIdx, bitOffset uint64 }
	mapMap    map[uint64][]uint64

	// Track which indices are in use for fast clearing
	usedIndicesGet  []uint64
	usedIndicesSet  []uint64
	usedIndicesHash []uint64
}

// newStorageMode creates a new storage mode instance based on the cache line count.
func newStorageMode(cacheLineCount uint64, hashCount uint32) *storageMode {
	useArrayMode := cacheLineCount <= ArrayModeThreshold

	s := &storageMode{
		useArrayMode:    useArrayMode,
		usedIndicesGet:  make([]uint64, 0, hashCount/8+1),
		usedIndicesSet:  make([]uint64, 0, hashCount/8+1),
		usedIndicesHash: make([]uint64, 0, hashCount/8+1),
	}

	if useArrayMode {
		// Small filter: use arrays for zero-overhead indexing
		s.arrayOps = &[ArrayModeThreshold][]opDetail{}
		s.arrayOpsSet = &[ArrayModeThreshold][]struct{ wordIdx, bitOffset uint64 }{}
		s.arrayMap = &[ArrayModeThreshold][]uint64{}
	} else {
		// Large filter: use maps for dynamic scaling
		estimatedCapacity := int(hashCount / 4)
		s.mapOps = make(map[uint64][]opDetail, estimatedCapacity)
		s.mapOpsSet = make(map[uint64][]struct{ wordIdx, bitOffset uint64 }, estimatedCapacity)
		s.mapMap = make(map[uint64][]uint64, estimatedCapacity)
	}

	return s
}

// clearHashMap clears the hash position map efficiently.
func (s *storageMode) clearHashMap() {
	if s.useArrayMode {
		// Clear only used indices - O(used) instead of O(capacity)
		for _, idx := range s.usedIndicesHash {
			s.arrayMap[idx] = s.arrayMap[idx][:0]
		}
		s.usedIndicesHash = s.usedIndicesHash[:0]
	} else {
		// Clear the map efficiently with Go 1.21+ built-in
		clear(s.mapMap)
		s.usedIndicesHash = s.usedIndicesHash[:0]
	}
}

// addHashPosition adds a bit position to the hash map for a given cache line.
func (s *storageMode) addHashPosition(cacheLineIdx uint64, bitPos uint64) {
	if s.useArrayMode {
		// Track first use of this cache line index
		if len(s.arrayMap[cacheLineIdx]) == 0 {
			s.usedIndicesHash = append(s.usedIndicesHash, cacheLineIdx)
		}
		s.arrayMap[cacheLineIdx] = append(s.arrayMap[cacheLineIdx], bitPos)
	} else {
		// Track first use of this cache line index
		// Check length to avoid double map lookup (auto-initializes on first append)
		if len(s.mapMap[cacheLineIdx]) == 0 {
			s.usedIndicesHash = append(s.usedIndicesHash, cacheLineIdx)
		}
		s.mapMap[cacheLineIdx] = append(s.mapMap[cacheLineIdx], bitPos)
	}
}

// getUsedHashIndices returns the list of cache line indices that have hash positions.
func (s *storageMode) getUsedHashIndices() []uint64 {
	return s.usedIndicesHash
}

// clearSetMap clears the set operation map efficiently.
func (s *storageMode) clearSetMap() {
	if s.useArrayMode {
		// Clear only used indices - O(used) instead of O(capacity)
		for _, idx := range s.usedIndicesSet {
			s.arrayOpsSet[idx] = s.arrayOpsSet[idx][:0]
		}
		s.usedIndicesSet = s.usedIndicesSet[:0]
	} else {
		// Clear the map efficiently with Go 1.21+ built-in
		clear(s.mapOpsSet)
		s.usedIndicesSet = s.usedIndicesSet[:0]
	}
}

// addSetOperation adds a set operation for a given cache line.
func (s *storageMode) addSetOperation(cacheLineIdx, wordIdx, bitOffset uint64) {
	if s.useArrayMode {
		// Track first use of this cache line index
		if len(s.arrayOpsSet[cacheLineIdx]) == 0 {
			s.usedIndicesSet = append(s.usedIndicesSet, cacheLineIdx)
		}
		s.arrayOpsSet[cacheLineIdx] = append(s.arrayOpsSet[cacheLineIdx], struct{ wordIdx, bitOffset uint64 }{
			wordIdx: wordIdx, bitOffset: bitOffset,
		})
	} else {
		// Track first use of this cache line index
		if len(s.mapOpsSet[cacheLineIdx]) == 0 {
			s.usedIndicesSet = append(s.usedIndicesSet, cacheLineIdx)
		}
		s.mapOpsSet[cacheLineIdx] = append(s.mapOpsSet[cacheLineIdx], struct{ wordIdx, bitOffset uint64 }{
			wordIdx: wordIdx, bitOffset: bitOffset,
		})
	}
}

// getSetOperations returns all set operations for a given cache line.
func (s *storageMode) getSetOperations(cacheLineIdx uint64) []struct{ wordIdx, bitOffset uint64 } {
	if s.useArrayMode {
		return s.arrayOpsSet[cacheLineIdx]
	}
	return s.mapOpsSet[cacheLineIdx]
}

// getUsedSetIndices returns the list of cache line indices that have set operations.
func (s *storageMode) getUsedSetIndices() []uint64 {
	return s.usedIndicesSet
}

// clearGetMap clears the get operation map efficiently.
func (s *storageMode) clearGetMap() {
	if s.useArrayMode {
		// Clear only used indices - O(used) instead of O(capacity)
		for _, idx := range s.usedIndicesGet {
			s.arrayOps[idx] = s.arrayOps[idx][:0]
		}
		s.usedIndicesGet = s.usedIndicesGet[:0]
	} else {
		// Clear the map efficiently with Go 1.21+ built-in
		clear(s.mapOps)
		s.usedIndicesGet = s.usedIndicesGet[:0]
	}
}

// addGetOperation adds a get operation for a given cache line.
func (s *storageMode) addGetOperation(cacheLineIdx, wordIdx, bitOffset uint64) {
	if s.useArrayMode {
		// Track first use of this cache line index
		if len(s.arrayOps[cacheLineIdx]) == 0 {
			s.usedIndicesGet = append(s.usedIndicesGet, cacheLineIdx)
		}
		s.arrayOps[cacheLineIdx] = append(s.arrayOps[cacheLineIdx], opDetail{
			wordIdx: wordIdx, bitOffset: bitOffset,
		})
	} else {
		// Track first use of this cache line index
		if len(s.mapOps[cacheLineIdx]) == 0 {
			s.usedIndicesGet = append(s.usedIndicesGet, cacheLineIdx)
		}
		s.mapOps[cacheLineIdx] = append(s.mapOps[cacheLineIdx], opDetail{
			wordIdx: wordIdx, bitOffset: bitOffset,
		})
	}
}

// getGetOperations returns all get operations for a given cache line.
func (s *storageMode) getGetOperations(cacheLineIdx uint64) []opDetail {
	if s.useArrayMode {
		return s.arrayOps[cacheLineIdx]
	}
	return s.mapOps[cacheLineIdx]
}

// getUsedGetIndices returns the list of cache line indices that have get operations.
func (s *storageMode) getUsedGetIndices() []uint64 {
	return s.usedIndicesGet
}
