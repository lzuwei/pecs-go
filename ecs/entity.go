package ecs

import "fmt"

// Entity represents a unique identifier for an entity in the ECS
// Uses generational index pattern: high bits for generation, low bits for index
type Entity uint32

const (
	// EntityIndexBits defines how many bits are used for the entity index
	EntityIndexBits = 20
	// EntityGenerationBits defines how many bits are used for the generation
	EntityGenerationBits = 12
	// EntityIndexMask is the mask for extracting the index part
	EntityIndexMask = (1 << EntityIndexBits) - 1
	// EntityGenerationMask is the mask for extracting the generation part
	EntityGenerationMask = (1 << EntityGenerationBits) - 1
	// NullEntity represents an invalid entity
	NullEntity Entity = 0xFFFFFFFF
)

// Index returns the index part of the entity
func (e Entity) Index() uint32 {
	return uint32(e) & EntityIndexMask
}

// Generation returns the generation part of the entity
func (e Entity) Generation() uint32 {
	return (uint32(e) >> EntityIndexBits) & EntityGenerationMask
}

// IsValid checks if the entity is valid (not null)
func (e Entity) IsValid() bool {
	return e != NullEntity
}

// String returns string representation of entity
func (e Entity) String() string {
	if !e.IsValid() {
		return "Entity(NULL)"
	}
	return fmt.Sprintf("Entity(%d.%d)", e.Index(), e.Generation())
}

// makeEntity creates an entity from index and generation
func makeEntity(index, generation uint32) Entity {
	return Entity((generation&EntityGenerationMask)<<EntityIndexBits | (index & EntityIndexMask))
}

// EntityManager manages entity creation, destruction, and recycling
type EntityManager struct {
	// entities stores generation for each entity index
	entities []uint32
	// freeHead points to the first free entity index, or -1 if none
	freeHead int32
}

// NewEntityManager creates a new entity manager
func NewEntityManager() *EntityManager {
	return &EntityManager{
		entities: make([]uint32, 0),
		freeHead: -1,
	}
}

// Create creates a new entity with proper ID recycling
func (em *EntityManager) Create() Entity {
	var index uint32
	var generation uint32

	if em.freeHead >= 0 {
		// Reuse a freed entity index
		index = uint32(em.freeHead)

		// The stored value is either the next free index or generation
		stored := em.entities[index]
		if stored == uint32(em.freeHead) {
			// This was the last free entity, no more in the chain
			em.freeHead = -1
			generation = 0 // Reset generation for reused entity
		} else {
			// Point to next free entity in the chain
			em.freeHead = int32(stored)
			generation = 0 // Reset generation for reused entity
		}

		// Store the new generation
		em.entities[index] = generation
	} else {
		// Create a new entity index
		index = uint32(len(em.entities))
		generation = 0
		em.entities = append(em.entities, generation)
	}

	return makeEntity(index, generation)
}

// Destroy marks an entity for reuse and increments its generation
func (em *EntityManager) Destroy(entity Entity) bool {
	if !entity.IsValid() {
		return false
	}

	index := entity.Index()
	if index >= uint32(len(em.entities)) {
		return false
	}

	currentGen := em.entities[index]
	expectedGen := entity.Generation()

	// Check if this is the current generation of the entity
	if currentGen != expectedGen {
		return false // Entity is stale
	}

	// Add to free list - store the previous free head
	if em.freeHead >= 0 {
		em.entities[index] = uint32(em.freeHead)
	} else {
		em.entities[index] = index // Point to itself if no free list
	}

	em.freeHead = int32(index)

	return true
}

// IsValid checks if an entity is valid and current
func (em *EntityManager) IsValid(entity Entity) bool {
	if !entity.IsValid() {
		return false
	}

	index := entity.Index()
	if index >= uint32(len(em.entities)) {
		return false
	}

	return em.entities[index] == entity.Generation()
}

// Size returns the number of entities that have been created
func (em *EntityManager) Size() int {
	return len(em.entities)
}

// Clear removes all entities
func (em *EntityManager) Clear() {
	em.entities = em.entities[:0]
	em.freeHead = -1
}
