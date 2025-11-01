package storage

import "sync"

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

// OperationStorage holds temporary storage for a single operation.
// This is pooled to avoid allocations and enable thread-safe concurrent operations.
type OperationStorage struct {
	UseArrayMode bool

	// Array-based storage (for small filters)
	ArrayOps    *[10000][]OpDetail
	ArrayOpsSet *[10000][]SetDetail
	ArrayMap    *[10000][]uint64

	// Map-based storage (for large filters)
	MapOps    map[uint64][]OpDetail
	MapOpsSet map[uint64][]SetDetail
	MapMap    map[uint64][]uint64

	// Track which indices are in use
	UsedIndicesGet  []uint64
	UsedIndicesSet  []uint64
	UsedIndicesHash []uint64
}

// clear resets the operation storage for reuse
func (os *OperationStorage) clear() {
	if os.UseArrayMode {
		// Clear only used indices
		for _, idx := range os.UsedIndicesGet {
			os.ArrayOps[idx] = os.ArrayOps[idx][:0]
		}
		for _, idx := range os.UsedIndicesSet {
			os.ArrayOpsSet[idx] = os.ArrayOpsSet[idx][:0]
		}
		for _, idx := range os.UsedIndicesHash {
			os.ArrayMap[idx] = os.ArrayMap[idx][:0]
		}
	} else {
		// Clear maps
		clear(os.MapOps)
		clear(os.MapOpsSet)
		clear(os.MapMap)
	}

	// Reset used indices
	os.UsedIndicesGet = os.UsedIndicesGet[:0]
	os.UsedIndicesSet = os.UsedIndicesSet[:0]
	os.UsedIndicesHash = os.UsedIndicesHash[:0]
}

// ClearGetMap clears the get operation map
func (os *OperationStorage) ClearGetMap() {
	if os.UseArrayMode {
		for _, idx := range os.UsedIndicesGet {
			os.ArrayOps[idx] = os.ArrayOps[idx][:0]
		}
		os.UsedIndicesGet = os.UsedIndicesGet[:0]
	} else {
		clear(os.MapOps)
		os.UsedIndicesGet = os.UsedIndicesGet[:0]
	}
}

// AddGetOperation adds a get operation for a given cache line
func (os *OperationStorage) AddGetOperation(cacheLineIdx, WordIdx, BitOffset uint64) {
	if os.UseArrayMode {
		if len(os.ArrayOps[cacheLineIdx]) == 0 {
			os.UsedIndicesGet = append(os.UsedIndicesGet, cacheLineIdx)
		}
		os.ArrayOps[cacheLineIdx] = append(os.ArrayOps[cacheLineIdx], OpDetail{
			WordIdx: WordIdx, BitOffset: BitOffset,
		})
	} else {
		if len(os.MapOps[cacheLineIdx]) == 0 {
			os.UsedIndicesGet = append(os.UsedIndicesGet, cacheLineIdx)
		}
		os.MapOps[cacheLineIdx] = append(os.MapOps[cacheLineIdx], OpDetail{
			WordIdx: WordIdx, BitOffset: BitOffset,
		})
	}
}

// GetGetOperations returns all get operations for a given cache line
func (os *OperationStorage) GetGetOperations(cacheLineIdx uint64) []OpDetail {
	if os.UseArrayMode {
		return os.ArrayOps[cacheLineIdx]
	}
	return os.MapOps[cacheLineIdx]
}

// GetUsedGetIndices returns the list of cache line indices that have get operations
func (os *OperationStorage) GetUsedGetIndices() []uint64 {
	return os.UsedIndicesGet
}

// ClearSetMap clears the set operation map
func (os *OperationStorage) ClearSetMap() {
	if os.UseArrayMode {
		for _, idx := range os.UsedIndicesSet {
			os.ArrayOpsSet[idx] = os.ArrayOpsSet[idx][:0]
		}
		os.UsedIndicesSet = os.UsedIndicesSet[:0]
	} else {
		clear(os.MapOpsSet)
		os.UsedIndicesSet = os.UsedIndicesSet[:0]
	}
}

// AddSetOperation adds a set operation for a given cache line
func (os *OperationStorage) AddSetOperation(cacheLineIdx, WordIdx, BitOffset uint64) {
	if os.UseArrayMode {
		if len(os.ArrayOpsSet[cacheLineIdx]) == 0 {
			os.UsedIndicesSet = append(os.UsedIndicesSet, cacheLineIdx)
		}
		os.ArrayOpsSet[cacheLineIdx] = append(os.ArrayOpsSet[cacheLineIdx], SetDetail{
			WordIdx: WordIdx, BitOffset: BitOffset,
		})
	} else {
		if len(os.MapOpsSet[cacheLineIdx]) == 0 {
			os.UsedIndicesSet = append(os.UsedIndicesSet, cacheLineIdx)
		}
		os.MapOpsSet[cacheLineIdx] = append(os.MapOpsSet[cacheLineIdx], SetDetail{
			WordIdx: WordIdx, BitOffset: BitOffset,
		})
	}
}

// GetSetOperations returns all set operations for a given cache line
func (os *OperationStorage) GetSetOperations(cacheLineIdx uint64) []SetDetail {
	if os.UseArrayMode {
		return os.ArrayOpsSet[cacheLineIdx]
	}
	return os.MapOpsSet[cacheLineIdx]
}

// GetUsedSetIndices returns the list of cache line indices that have set operations
func (os *OperationStorage) GetUsedSetIndices() []uint64 {
	return os.UsedIndicesSet
}

// ClearHashMap clears the hash position map
func (os *OperationStorage) ClearHashMap() {
	if os.UseArrayMode {
		for _, idx := range os.UsedIndicesHash {
			os.ArrayMap[idx] = os.ArrayMap[idx][:0]
		}
		os.UsedIndicesHash = os.UsedIndicesHash[:0]
	} else {
		clear(os.MapMap)
		os.UsedIndicesHash = os.UsedIndicesHash[:0]
	}
}

// AddHashPosition adds a bit position to the hash map for a given cache line
func (os *OperationStorage) AddHashPosition(cacheLineIdx uint64, bitPos uint64) {
	if os.UseArrayMode {
		if len(os.ArrayMap[cacheLineIdx]) == 0 {
			os.UsedIndicesHash = append(os.UsedIndicesHash, cacheLineIdx)
		}
		os.ArrayMap[cacheLineIdx] = append(os.ArrayMap[cacheLineIdx], bitPos)
	} else {
		if len(os.MapMap[cacheLineIdx]) == 0 {
			os.UsedIndicesHash = append(os.UsedIndicesHash, cacheLineIdx)
		}
		os.MapMap[cacheLineIdx] = append(os.MapMap[cacheLineIdx], bitPos)
	}
}

// GetUsedHashIndices returns the list of cache line indices that have hash positions
func (os *OperationStorage) GetUsedHashIndices() []uint64 {
	return os.UsedIndicesHash
}

// Pool for operation storage - separate pools for array and map modes
var (
	arrayOpsPool = sync.Pool{
		New: func() interface{} {
			return &OperationStorage{
				UseArrayMode:    true,
				ArrayOps:        &[10000][]OpDetail{},
				ArrayOpsSet:     &[10000][]SetDetail{},
				ArrayMap:        &[10000][]uint64{},
				UsedIndicesGet:  make([]uint64, 0, 8),
				UsedIndicesSet:  make([]uint64, 0, 8),
				UsedIndicesHash: make([]uint64, 0, 8),
			}
		},
	}

	mapOpsPool = sync.Pool{
		New: func() interface{} {
			return &OperationStorage{
				UseArrayMode:    false,
				MapOps:          make(map[uint64][]OpDetail, 32),
				MapOpsSet:       make(map[uint64][]SetDetail, 32),
				MapMap:          make(map[uint64][]uint64, 32),
				UsedIndicesGet:  make([]uint64, 0, 32),
				UsedIndicesSet:  make([]uint64, 0, 32),
				UsedIndicesHash: make([]uint64, 0, 32),
			}
		},
	}
)

// GetOperationStorage retrieves an operation storage from the pool
// Objects from pool are already clean (either new or cleared on Put)
func GetOperationStorage(useArrayMode bool) *OperationStorage {
	if useArrayMode {
		return arrayOpsPool.Get().(*OperationStorage)
	}
	return mapOpsPool.Get().(*OperationStorage)
}

// PutOperationStorage returns an operation storage to the pool after clearing it
func PutOperationStorage(ops *OperationStorage) {
	// Clear before returning to pool to ensure next Get receives clean object
	ops.clear()

	if ops.UseArrayMode {
		arrayOpsPool.Put(ops)
	} else {
		mapOpsPool.Put(ops)
	}
}

// Mode handles the hybrid array/map storage configuration.
// With sync.Pool, this just tracks the mode setting, not the actual storage.
type Mode struct {
	UseArrayMode bool
}

// New creates a new storage mode instance based on the cache line count.
func New(cacheLineCount uint64, hashCount uint32, arrayModeThreshold uint64) *Mode {
	useArrayMode := cacheLineCount <= arrayModeThreshold

	return &Mode{
		UseArrayMode: useArrayMode,
	}
}
