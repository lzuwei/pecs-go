package ecs

import (
	"reflect"
	"unsafe"
)

// ComponentPool stores components of a specific type using sparse set architecture
type ComponentPool[T any] struct {
	entities   *SparseSet // Tracks which entities have this component
	components []T        // Component data aligned with entities dense array
}

// NewComponentPool creates a new component pool for type T
func NewComponentPool[T any]() *ComponentPool[T] {
	return &ComponentPool[T]{
		entities:   NewSparseSet(),
		components: make([]T, 0),
	}
}

// Insert adds a component to an entity
func (cp *ComponentPool[T]) Insert(entity Entity, component T) {
	if cp.entities.Contains(entity) {
		// Update existing component
		index := cp.entities.Index(entity)
		cp.components[index] = component
		return
	}

	// Add new component
	if cp.entities.Insert(entity) {
		// Grow component array if needed
		if len(cp.components) <= cp.entities.Size()-1 {
			cp.components = append(cp.components, component)
		} else {
			cp.components[cp.entities.Size()-1] = component
		}
	}
}

// Remove removes a component from an entity
func (cp *ComponentPool[T]) Remove(entity Entity) bool {
	if !cp.entities.Contains(entity) {
		return false
	}

	index := cp.entities.Index(entity)
	lastIndex := cp.entities.Size() - 1

	// Move last component to removed position before removing from sparse set
	if index != lastIndex {
		cp.components[index] = cp.components[lastIndex]
	}

	return cp.entities.Remove(entity)
}

// Get retrieves a component for an entity
func (cp *ComponentPool[T]) Get(entity Entity) (T, bool) {
	var zero T
	if !cp.entities.Contains(entity) {
		return zero, false
	}

	index := cp.entities.Index(entity)
	return cp.components[index], true
}

// GetPtr returns a pointer to the component for an entity
func (cp *ComponentPool[T]) GetPtr(entity Entity) *T {
	if !cp.entities.Contains(entity) {
		return nil
	}

	index := cp.entities.Index(entity)
	return &cp.components[index]
}

// Contains checks if an entity has this component
func (cp *ComponentPool[T]) Contains(entity Entity) bool {
	return cp.entities.Contains(entity)
}

// Size returns the number of entities with this component
func (cp *ComponentPool[T]) Size() int {
	return cp.entities.Size()
}

// Empty checks if the pool is empty
func (cp *ComponentPool[T]) Empty() bool {
	return cp.entities.Empty()
}

// Clear removes all components
func (cp *ComponentPool[T]) Clear() {
	cp.entities.Clear()
	cp.components = cp.components[:0]
}

// Entities returns the sparse set of entities
func (cp *ComponentPool[T]) Entities() *SparseSet {
	return cp.entities
}

// Data returns raw component data (aligned with entities.Data())
func (cp *ComponentPool[T]) Data() []T {
	return cp.components[:cp.entities.Size()]
}

// ForEach iterates over all entities and their components
func (cp *ComponentPool[T]) ForEach(fn func(Entity, *T)) {
	entities := cp.entities.Data()
	for i, entity := range entities {
		fn(entity, &cp.components[i])
	}
}

// Sort sorts components by the given comparison function
func (cp *ComponentPool[T]) Sort(less func(Entity, *T, Entity, *T) bool) {
	cp.entities.Sort(func(a, b Entity) bool {
		indexA := cp.entities.Index(a)
		indexB := cp.entities.Index(b)
		return less(a, &cp.components[indexA], b, &cp.components[indexB])
	})
}

// Respect reorders this pool to match another sparse set's order
func (cp *ComponentPool[T]) Respect(other *SparseSet) {
	if other.Size() == 0 {
		return
	}

	// Create new component array in the order of other
	newComponents := make([]T, 0, cp.entities.Size())

	// First, add components for entities that exist in other
	for i := 0; i < other.Size(); i++ {
		entity := other.At(i)
		if cp.entities.Contains(entity) {
			index := cp.entities.Index(entity)
			newComponents = append(newComponents, cp.components[index])
		}
	}

	// Then add remaining components
	entities := cp.entities.Data()
	for i, entity := range entities {
		found := false
		for j := 0; j < other.Size(); j++ {
			if other.At(j) == entity {
				found = true
				break
			}
		}
		if !found {
			newComponents = append(newComponents, cp.components[i])
		}
	}

	// Update entities order and components
	cp.entities.Respect(other)
	copy(cp.components[:len(newComponents)], newComponents)
}

// IComponentStorage is the interface for type-erased component storage
type IComponentStorage interface {
	Remove(entity Entity) bool
	Contains(entity Entity) bool
	Size() int
	Clear()
	Entities() *SparseSet
	TypeName() string
}

// TypedStorage wraps ComponentPool to implement IComponentStorage
type TypedStorage[T any] struct {
	pool     *ComponentPool[T]
	typeName string
}

// NewTypedStorage creates a new typed storage wrapper
func NewTypedStorage[T any]() *TypedStorage[T] {
	var zero T
	typeName := reflect.TypeOf(zero).String()
	return &TypedStorage[T]{
		pool:     NewComponentPool[T](),
		typeName: typeName,
	}
}

// Pool returns the underlying component pool
func (ts *TypedStorage[T]) Pool() *ComponentPool[T] {
	return ts.pool
}

// Remove removes a component from an entity
func (ts *TypedStorage[T]) Remove(entity Entity) bool {
	return ts.pool.Remove(entity)
}

// Contains checks if an entity has this component
func (ts *TypedStorage[T]) Contains(entity Entity) bool {
	return ts.pool.Contains(entity)
}

// Size returns the number of entities with this component
func (ts *TypedStorage[T]) Size() int {
	return ts.pool.Size()
}

// Clear removes all components
func (ts *TypedStorage[T]) Clear() {
	ts.pool.Clear()
}

// Entities returns the sparse set of entities
func (ts *TypedStorage[T]) Entities() *SparseSet {
	return ts.pool.Entities()
}

// TypeName returns the component type name
func (ts *TypedStorage[T]) TypeName() string {
	return ts.typeName
}

// ComponentID represents a unique identifier for a component type
type ComponentID uint32

// ComponentRegistry manages component type registration and storage
type ComponentRegistry struct {
	nextID   ComponentID
	typeToID map[reflect.Type]ComponentID
	idToType map[ComponentID]reflect.Type
	storages map[ComponentID]IComponentStorage
	names    map[ComponentID]string
}

// NewComponentRegistry creates a new component registry
func NewComponentRegistry() *ComponentRegistry {
	return &ComponentRegistry{
		nextID:   0,
		typeToID: make(map[reflect.Type]ComponentID),
		idToType: make(map[ComponentID]reflect.Type),
		storages: make(map[ComponentID]IComponentStorage),
		names:    make(map[ComponentID]string),
	}
}

// Register registers a component type and returns its ID
func Register[T any](cr *ComponentRegistry) ComponentID {
	var zero T
	componentType := reflect.TypeOf(zero)

	// Check if already registered
	if id, exists := cr.typeToID[componentType]; exists {
		return id
	}

	// Register new component type
	id := cr.nextID
	cr.nextID++

	storage := NewTypedStorage[T]()

	cr.typeToID[componentType] = id
	cr.idToType[id] = componentType
	cr.storages[id] = storage
	cr.names[id] = componentType.String()

	return id
}

// GetComponentID returns the component ID for a given type
func GetComponentID[T any](cr *ComponentRegistry) (ComponentID, bool) {
	var zero T
	componentType := reflect.TypeOf(zero)
	id, exists := cr.typeToID[componentType]
	return id, exists
}

// GetStorage returns the typed storage for a component type
func GetStorage[T any](cr *ComponentRegistry) (*ComponentPool[T], bool) {
	id, exists := GetComponentID[T](cr)
	if !exists {
		return nil, false
	}

	storage, exists := cr.storages[id]
	if !exists {
		return nil, false
	}

	typedStorage, ok := storage.(*TypedStorage[T])
	if !ok {
		return nil, false
	}

	return typedStorage.Pool(), true
}

// GetStorageByID returns the type-erased storage for a component ID
func (cr *ComponentRegistry) GetStorageByID(id ComponentID) (IComponentStorage, bool) {
	storage, exists := cr.storages[id]
	return storage, exists
}

// RemoveAllComponents removes all components from an entity
func (cr *ComponentRegistry) RemoveAllComponents(entity Entity) {
	for _, storage := range cr.storages {
		storage.Remove(entity)
	}
}

// GetComponentName returns the name of a component type by ID
func (cr *ComponentRegistry) GetComponentName(id ComponentID) string {
	if name, exists := cr.names[id]; exists {
		return name
	}
	return "Unknown"
}

// GetRegisteredTypes returns all registered component types
func (cr *ComponentRegistry) GetRegisteredTypes() map[ComponentID]string {
	result := make(map[ComponentID]string)
	for id, name := range cr.names {
		result[id] = name
	}
	return result
}

// UnsafePointer returns an unsafe pointer to component data for entity
// This is used for advanced operations and should be used with caution
func UnsafeComponentPointer[T any](cr *ComponentRegistry, entity Entity) unsafe.Pointer {
	storage, exists := GetStorage[T](cr)
	if !exists {
		return nil
	}

	ptr := storage.GetPtr(entity)
	if ptr == nil {
		return nil
	}

	return unsafe.Pointer(ptr)
}
