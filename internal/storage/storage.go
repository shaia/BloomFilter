package storage

// OpDetail represents a bit operation within a cache line (word index and bit offset).
type OpDetail struct {
	WordIdx   uint64
	BitOffset uint64
}

// SetDetail represents a set operation within a cache line (word index and bit offset).
// This is used specifically for setBitCacheOptimized operations.
type SetDetail struct {
	WordIdx   uint64
	BitOffset uint64
}

// Mode handles the hybrid array/map storage abstraction.
// This encapsulates the logic for choosing between array mode (small filters)
// and map mode (large filters) without duplicating code.
type Mode struct {
	UseArrayMode bool

	// Array-based storage (for small filters, zero-overhead indexing)
	ArrayOps    *[10000][]OpDetail
	ArrayOpsSet *[10000][]SetDetail
	ArrayMap    *[10000][]uint64

	// Map-based storage (for large filters, dynamic scaling)
	MapOps    map[uint64][]OpDetail
	MapOpsSet map[uint64][]SetDetail
	MapMap    map[uint64][]uint64

	// Track which indices are in use for fast clearing
	UsedIndicesGet  []uint64
	UsedIndicesSet  []uint64
	UsedIndicesHash []uint64
}

// New creates a new storage mode instance based on the cache line count.
func New(cacheLineCount uint64, hashCount uint32, arrayModeThreshold uint64) *Mode {
	useArrayMode := cacheLineCount <= arrayModeThreshold

	s := &Mode{
		UseArrayMode:    useArrayMode,
		UsedIndicesGet:  make([]uint64, 0, hashCount/8+1),
		UsedIndicesSet:  make([]uint64, 0, hashCount/8+1),
		UsedIndicesHash: make([]uint64, 0, hashCount/8+1),
	}

	if useArrayMode {
		// Small filter: use arrays for zero-overhead indexing
		s.ArrayOps = &[10000][]OpDetail{}
		s.ArrayOpsSet = &[10000][]SetDetail{}
		s.ArrayMap = &[10000][]uint64{}
	} else {
		// Large filter: use maps for dynamic scaling
		estimatedCapacity := int(hashCount / 4)
		s.MapOps = make(map[uint64][]OpDetail, estimatedCapacity)
		s.MapOpsSet = make(map[uint64][]SetDetail, estimatedCapacity)
		s.MapMap = make(map[uint64][]uint64, estimatedCapacity)
	}

	return s
}

// clearHashMap clears the hash position map efficiently.
func (s *Mode) ClearHashMap() {
	if s.UseArrayMode {
		// Clear only used indices - O(used) instead of O(capacity)
		for _, idx := range s.UsedIndicesHash {
			s.ArrayMap[idx] = s.ArrayMap[idx][:0]
		}
		s.UsedIndicesHash = s.UsedIndicesHash[:0]
	} else {
		// Clear the map efficiently with Go 1.21+ built-in
		clear(s.MapMap)
		s.UsedIndicesHash = s.UsedIndicesHash[:0]
	}
}

// addHashPosition adds a bit position to the hash map for a given cache line.
func (s *Mode) AddHashPosition(cacheLineIdx uint64, bitPos uint64) {
	if s.UseArrayMode {
		// Track first use of this cache line index
		if len(s.ArrayMap[cacheLineIdx]) == 0 {
			s.UsedIndicesHash = append(s.UsedIndicesHash, cacheLineIdx)
		}
		s.ArrayMap[cacheLineIdx] = append(s.ArrayMap[cacheLineIdx], bitPos)
	} else {
		// Track first use of this cache line index
		// Check length to avoid double map lookup (auto-initializes on first append)
		if len(s.MapMap[cacheLineIdx]) == 0 {
			s.UsedIndicesHash = append(s.UsedIndicesHash, cacheLineIdx)
		}
		s.MapMap[cacheLineIdx] = append(s.MapMap[cacheLineIdx], bitPos)
	}
}

// getUsedHashIndices returns the list of cache line indices that have hash positions.
func (s *Mode) GetUsedHashIndices() []uint64 {
	return s.UsedIndicesHash
}

// clearSetMap clears the set operation map efficiently.
func (s *Mode) ClearSetMap() {
	if s.UseArrayMode {
		// Clear only used indices - O(used) instead of O(capacity)
		for _, idx := range s.UsedIndicesSet {
			s.ArrayOpsSet[idx] = s.ArrayOpsSet[idx][:0]
		}
		s.UsedIndicesSet = s.UsedIndicesSet[:0]
	} else {
		// Clear the map efficiently with Go 1.21+ built-in
		clear(s.MapOpsSet)
		s.UsedIndicesSet = s.UsedIndicesSet[:0]
	}
}

// addSetOperation adds a set operation for a given cache line.
func (s *Mode) AddSetOperation(cacheLineIdx, WordIdx, BitOffset uint64) {
	if s.UseArrayMode {
		// Track first use of this cache line index
		if len(s.ArrayOpsSet[cacheLineIdx]) == 0 {
			s.UsedIndicesSet = append(s.UsedIndicesSet, cacheLineIdx)
		}
		s.ArrayOpsSet[cacheLineIdx] = append(s.ArrayOpsSet[cacheLineIdx], SetDetail{
			WordIdx: WordIdx, BitOffset: BitOffset,
		})
	} else {
		// Track first use of this cache line index
		if len(s.MapOpsSet[cacheLineIdx]) == 0 {
			s.UsedIndicesSet = append(s.UsedIndicesSet, cacheLineIdx)
		}
		s.MapOpsSet[cacheLineIdx] = append(s.MapOpsSet[cacheLineIdx], SetDetail{
			WordIdx: WordIdx, BitOffset: BitOffset,
		})
	}
}

// getSetOperations returns all set operations for a given cache line.
func (s *Mode) GetSetOperations(cacheLineIdx uint64) []SetDetail {
	if s.UseArrayMode {
		return s.ArrayOpsSet[cacheLineIdx]
	}
	return s.MapOpsSet[cacheLineIdx]
}

// getUsedSetIndices returns the list of cache line indices that have set operations.
func (s *Mode) GetUsedSetIndices() []uint64 {
	return s.UsedIndicesSet
}

// clearGetMap clears the get operation map efficiently.
func (s *Mode) ClearGetMap() {
	if s.UseArrayMode {
		// Clear only used indices - O(used) instead of O(capacity)
		for _, idx := range s.UsedIndicesGet {
			s.ArrayOps[idx] = s.ArrayOps[idx][:0]
		}
		s.UsedIndicesGet = s.UsedIndicesGet[:0]
	} else {
		// Clear the map efficiently with Go 1.21+ built-in
		clear(s.MapOps)
		s.UsedIndicesGet = s.UsedIndicesGet[:0]
	}
}

// addGetOperation adds a get operation for a given cache line.
func (s *Mode) AddGetOperation(cacheLineIdx, WordIdx, BitOffset uint64) {
	if s.UseArrayMode {
		// Track first use of this cache line index
		if len(s.ArrayOps[cacheLineIdx]) == 0 {
			s.UsedIndicesGet = append(s.UsedIndicesGet, cacheLineIdx)
		}
		s.ArrayOps[cacheLineIdx] = append(s.ArrayOps[cacheLineIdx], OpDetail{
			WordIdx: WordIdx, BitOffset: BitOffset,
		})
	} else {
		// Track first use of this cache line index
		if len(s.MapOps[cacheLineIdx]) == 0 {
			s.UsedIndicesGet = append(s.UsedIndicesGet, cacheLineIdx)
		}
		s.MapOps[cacheLineIdx] = append(s.MapOps[cacheLineIdx], OpDetail{
			WordIdx: WordIdx, BitOffset: BitOffset,
		})
	}
}

// getGetOperations returns all get operations for a given cache line.
func (s *Mode) GetGetOperations(cacheLineIdx uint64) []OpDetail {
	if s.UseArrayMode {
		return s.ArrayOps[cacheLineIdx]
	}
	return s.MapOps[cacheLineIdx]
}

// getUsedGetIndices returns the list of cache line indices that have get operations.
func (s *Mode) GetUsedGetIndices() []uint64 {
	return s.UsedIndicesGet
}
