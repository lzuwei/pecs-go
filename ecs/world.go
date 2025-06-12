package ecs

// World represents the main ECS world containing entities, components, and systems
type World struct {
	entityManager     *EntityManager
	componentRegistry *ComponentRegistry
	systemManager     *SystemManager
}

// NewWorld creates a new ECS world
func NewWorld() *World {
	return &World{
		entityManager:     NewEntityManager(),
		componentRegistry: NewComponentRegistry(),
		systemManager:     NewSystemManager(),
	}
}

// CreateEntity creates a new entity
func (w *World) CreateEntity() Entity {
	return w.entityManager.Create()
}

// DestroyEntity destroys an entity and removes all its components
func (w *World) DestroyEntity(entity Entity) bool {
	if !w.entityManager.IsValid(entity) {
		return false
	}

	w.componentRegistry.RemoveAllComponents(entity)
	return w.entityManager.Destroy(entity)
}

// IsValidEntity checks if an entity is valid
func (w *World) IsValidEntity(entity Entity) bool {
	return w.entityManager.IsValid(entity)
}

// AddComponent adds a component to an entity
func AddComponent[T any](w *World, entity Entity, component T) {
	if !w.entityManager.IsValid(entity) {
		return
	}

	Register[T](w.componentRegistry)
	if storage, exists := GetStorage[T](w.componentRegistry); exists {
		storage.Insert(entity, component)
	}
}

// RemoveComponent removes a component from an entity
func RemoveComponent[T any](w *World, entity Entity) bool {
	if !w.entityManager.IsValid(entity) {
		return false
	}

	if storage, exists := GetStorage[T](w.componentRegistry); exists {
		return storage.Remove(entity)
	}
	return false
}

// GetComponent retrieves a component from an entity
func GetComponent[T any](w *World, entity Entity) (T, bool) {
	var zero T
	if !w.entityManager.IsValid(entity) {
		return zero, false
	}

	if storage, exists := GetStorage[T](w.componentRegistry); exists {
		return storage.Get(entity)
	}
	return zero, false
}

// GetComponentPtr returns a pointer to a component for an entity
func GetComponentPtr[T any](w *World, entity Entity) *T {
	if !w.entityManager.IsValid(entity) {
		return nil
	}

	if storage, exists := GetStorage[T](w.componentRegistry); exists {
		return storage.GetPtr(entity)
	}
	return nil
}

// HasComponent checks if an entity has a specific component
func HasComponent[T any](w *World, entity Entity) bool {
	if !w.entityManager.IsValid(entity) {
		return false
	}

	if storage, exists := GetStorage[T](w.componentRegistry); exists {
		return storage.Contains(entity)
	}
	return false
}

// Query creates a new query for this world
func (w *World) Query() *Query {
	return NewQuery(w)
}

// View creates a new view builder for this world
func (w *World) View() *ViewBuilder {
	return NewViewBuilder(w)
}

// Iter1 creates a new single-component iterator
func Iter1[T1 any](w *World) *Iterator1[T1] {
	return NewIterator1[T1](w)
}

// Iter2 creates a new two-component iterator
func Iter2[T1, T2 any](w *World) *Iterator2[T1, T2] {
	return NewIterator2[T1, T2](w)
}

// Iter3 creates a new three-component iterator
func Iter3[T1, T2, T3 any](w *World) *Iterator3[T1, T2, T3] {
	return NewIterator3[T1, T2, T3](w)
}

// GetEntityManager returns the entity manager
func (w *World) GetEntityManager() *EntityManager {
	return w.entityManager
}

// GetComponentRegistry returns the component registry
func (w *World) GetComponentRegistry() *ComponentRegistry {
	return w.componentRegistry
}

// GetSystemManager returns the system manager
func (w *World) GetSystemManager() *SystemManager {
	return w.systemManager
}

// AddSystem adds a system to the world
func (w *World) AddSystem(system System) {
	w.systemManager.AddSystem(system)
}

// RemoveSystem removes a system from the world
func (w *World) RemoveSystem(system System) {
	w.systemManager.RemoveSystem(system)
}

// EnableSystem enables a system
func (w *World) EnableSystem(system System) {
	w.systemManager.EnableSystem(system)
}

// DisableSystem disables a system
func (w *World) DisableSystem(system System) {
	w.systemManager.DisableSystem(system)
}

// Update updates all enabled systems
func (w *World) Update(deltaTime float64) {
	w.systemManager.Update(w, deltaTime)
}

// Clear removes all entities, components, and systems
func (w *World) Clear() {
	w.systemManager.Clear()
	w.componentRegistry = NewComponentRegistry()
	w.entityManager.Clear()
}

// Stats returns statistics about the world
func (w *World) Stats() WorldStats {
	entityCount := w.entityManager.Size()
	componentTypes := len(w.componentRegistry.GetRegisteredTypes())
	systemCount := len(w.systemManager.GetSystems())

	var totalComponents int
	for _, storage := range w.componentRegistry.storages {
		totalComponents += storage.Size()
	}

	return WorldStats{
		EntityCount:     entityCount,
		ComponentTypes:  componentTypes,
		TotalComponents: totalComponents,
		SystemCount:     systemCount,
	}
}

// WorldStats contains statistics about the world
type WorldStats struct {
	EntityCount     int
	ComponentTypes  int
	TotalComponents int
	SystemCount     int
}
