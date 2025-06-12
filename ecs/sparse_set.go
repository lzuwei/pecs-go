package ecs

// SparseSet is a data structure that provides O(1) insertion, deletion, and lookup
// It's the foundation for efficient component storage in the ECS
type SparseSet struct {
	sparse []int32  // Maps entity index to dense array index (-1 means not present)
	dense  []Entity // Packed array of entities
	size   int      // Current number of elements
}

// NewSparseSet creates a new sparse set
func NewSparseSet() *SparseSet {
	return &SparseSet{
		sparse: make([]int32, 0),
		dense:  make([]Entity, 0),
		size:   0,
	}
}

// ensureCapacity ensures the sparse array can hold the given entity index
func (ss *SparseSet) ensureCapacity(entityIndex uint32) {
	needed := int(entityIndex) + 1
	if len(ss.sparse) < needed {
		// Grow sparse array to accommodate entity
		oldLen := len(ss.sparse)
		newSparse := make([]int32, needed)
		copy(newSparse, ss.sparse)
		// Initialize new slots to -1 (not present)
		for i := oldLen; i < needed; i++ {
			newSparse[i] = -1
		}
		ss.sparse = newSparse
	}
}

// Contains checks if an entity exists in the set
func (ss *SparseSet) Contains(entity Entity) bool {
	if !entity.IsValid() {
		return false
	}

	entityIndex := entity.Index()
	if int(entityIndex) >= len(ss.sparse) {
		return false
	}

	denseIndex := ss.sparse[entityIndex]
	return denseIndex >= 0 && int(denseIndex) < ss.size && ss.dense[denseIndex] == entity
}

// Insert adds an entity to the set
func (ss *SparseSet) Insert(entity Entity) bool {
	if !entity.IsValid() {
		return false
	}

	entityIndex := entity.Index()
	ss.ensureCapacity(entityIndex)

	if ss.Contains(entity) {
		return false // Already present
	}

	// Add new entity
	ss.sparse[entityIndex] = int32(ss.size)

	// Grow dense array if needed
	if len(ss.dense) <= ss.size {
		ss.dense = append(ss.dense, entity)
	} else {
		ss.dense[ss.size] = entity
	}

	ss.size++
	return true
}

// Remove removes an entity from the set
func (ss *SparseSet) Remove(entity Entity) bool {
	if !ss.Contains(entity) {
		return false
	}

	entityIndex := entity.Index()
	denseIndex := ss.sparse[entityIndex]
	lastIndex := int32(ss.size - 1)

	if denseIndex != lastIndex {
		// Move last element to the removed element's position (swap-and-pop)
		lastEntity := ss.dense[lastIndex]
		ss.dense[denseIndex] = lastEntity
		ss.sparse[lastEntity.Index()] = denseIndex
	}

	ss.sparse[entityIndex] = -1
	ss.size--

	return true
}

// Size returns the number of entities in the set
func (ss *SparseSet) Size() int {
	return ss.size
}

// Empty checks if the set is empty
func (ss *SparseSet) Empty() bool {
	return ss.size == 0
}

// Clear removes all entities from the set
func (ss *SparseSet) Clear() {
	ss.size = 0
	// Reset sparse array
	for i := range ss.sparse {
		ss.sparse[i] = -1
	}
}

// Data returns the raw dense array (for iteration)
func (ss *SparseSet) Data() []Entity {
	return ss.dense[:ss.size]
}

// At returns the entity at the given dense index
func (ss *SparseSet) At(index int) Entity {
	if index < 0 || index >= ss.size {
		return NullEntity
	}
	return ss.dense[index]
}

// Index returns the dense index of an entity, or -1 if not found
func (ss *SparseSet) Index(entity Entity) int {
	if !ss.Contains(entity) {
		return -1
	}
	return int(ss.sparse[entity.Index()])
}

// ForEach iterates over all entities in the set
func (ss *SparseSet) ForEach(fn func(Entity)) {
	for i := 0; i < ss.size; i++ {
		fn(ss.dense[i])
	}
}

// Swap swaps two entities in the dense array (useful for sorting)
func (ss *SparseSet) Swap(i, j int) {
	if i < 0 || i >= ss.size || j < 0 || j >= ss.size {
		return
	}

	entityI := ss.dense[i]
	entityJ := ss.dense[j]

	// Swap in dense array
	ss.dense[i] = entityJ
	ss.dense[j] = entityI

	// Update sparse array
	ss.sparse[entityI.Index()] = int32(j)
	ss.sparse[entityJ.Index()] = int32(i)
}

// Sort sorts the entities using the provided comparison function
func (ss *SparseSet) Sort(less func(Entity, Entity) bool) {
	// Simple bubble sort for now - could be optimized with quicksort/introsort
	for i := 0; i < ss.size-1; i++ {
		for j := 0; j < ss.size-i-1; j++ {
			if less(ss.dense[j+1], ss.dense[j]) {
				ss.Swap(j, j+1)
			}
		}
	}
}

// Respect maintains the order of entities according to another sparse set
// This is useful for implementing groups
func (ss *SparseSet) Respect(other *SparseSet) {
	if other.size == 0 {
		return
	}

	// Create temporary arrays for reordering
	newDense := make([]Entity, 0, ss.size)

	// First, add entities that exist in other in the same order
	for i := 0; i < other.size; i++ {
		entity := other.dense[i]
		if ss.Contains(entity) {
			newDense = append(newDense, entity)
		}
	}

	// Then add remaining entities
	for i := 0; i < ss.size; i++ {
		entity := ss.dense[i]
		found := false
		for j := 0; j < len(newDense); j++ {
			if newDense[j] == entity {
				found = true
				break
			}
		}
		if !found {
			newDense = append(newDense, entity)
		}
	}

	// Update dense array and sparse indices
	copy(ss.dense[:len(newDense)], newDense)
	for i, entity := range newDense {
		ss.sparse[entity.Index()] = int32(i)
	}
}
