package ecs

// System represents a system that processes entities
type System interface {
	// Update is called every frame/tick
	Update(world *World, deltaTime float64)

	// GetName returns the system name for debugging
	GetName() string
}

// SystemManager manages all systems in the ECS
type SystemManager struct {
	systems []System
	enabled map[System]bool
}

// NewSystemManager creates a new system manager
func NewSystemManager() *SystemManager {
	return &SystemManager{
		systems: make([]System, 0),
		enabled: make(map[System]bool),
	}
}

// AddSystem adds a system to the manager
func (sm *SystemManager) AddSystem(system System) {
	sm.systems = append(sm.systems, system)
	sm.enabled[system] = true
}

// RemoveSystem removes a system from the manager
func (sm *SystemManager) RemoveSystem(system System) {
	for i, s := range sm.systems {
		if s == system {
			// Remove system from slice
			sm.systems = append(sm.systems[:i], sm.systems[i+1:]...)
			delete(sm.enabled, system)
			break
		}
	}
}

// EnableSystem enables a system
func (sm *SystemManager) EnableSystem(system System) {
	sm.enabled[system] = true
}

// DisableSystem disables a system
func (sm *SystemManager) DisableSystem(system System) {
	sm.enabled[system] = false
}

// IsEnabled checks if a system is enabled
func (sm *SystemManager) IsEnabled(system System) bool {
	enabled, exists := sm.enabled[system]
	return exists && enabled
}

// Update updates all enabled systems
func (sm *SystemManager) Update(world *World, deltaTime float64) {
	for _, system := range sm.systems {
		if sm.IsEnabled(system) {
			system.Update(world, deltaTime)
		}
	}
}

// GetSystems returns all systems
func (sm *SystemManager) GetSystems() []System {
	return sm.systems
}

// GetEnabledSystems returns all enabled systems
func (sm *SystemManager) GetEnabledSystems() []System {
	enabled := make([]System, 0)
	for _, system := range sm.systems {
		if sm.IsEnabled(system) {
			enabled = append(enabled, system)
		}
	}
	return enabled
}

// Clear removes all systems
func (sm *SystemManager) Clear() {
	sm.systems = sm.systems[:0]
	sm.enabled = make(map[System]bool)
}

// BaseSystem provides a basic implementation of System interface
type BaseSystem struct {
	name string
}

// NewBaseSystem creates a new base system
func NewBaseSystem(name string) *BaseSystem {
	return &BaseSystem{name: name}
}

// GetName returns the system name
func (bs *BaseSystem) GetName() string {
	return bs.name
}

// Update is a default implementation that does nothing
// Override this in your concrete systems
func (bs *BaseSystem) Update(world *World, deltaTime float64) {
	// Default implementation does nothing
}

// System1 is a convenience system that processes entities with one component type
type System1[T1 any] struct {
	*BaseSystem
	updateFunc func(*World, float64, Entity, *T1)
}

// NewSystem1 creates a new single-component system
func NewSystem1[T1 any](name string, updateFunc func(*World, float64, Entity, *T1)) *System1[T1] {
	return &System1[T1]{
		BaseSystem: NewBaseSystem(name),
		updateFunc: updateFunc,
	}
}

// Update processes all entities with the required component
func (s *System1[T1]) Update(world *World, deltaTime float64) {
	Iter1[T1](world).ForEach(func(entity Entity, comp1 *T1) {
		s.updateFunc(world, deltaTime, entity, comp1)
	})
}

// System2 is a convenience system that processes entities with two component types
type System2[T1, T2 any] struct {
	*BaseSystem
	updateFunc func(*World, float64, Entity, *T1, *T2)
}

// NewSystem2 creates a new two-component system
func NewSystem2[T1, T2 any](name string, updateFunc func(*World, float64, Entity, *T1, *T2)) *System2[T1, T2] {
	return &System2[T1, T2]{
		BaseSystem: NewBaseSystem(name),
		updateFunc: updateFunc,
	}
}

// Update processes all entities with the required components
func (s *System2[T1, T2]) Update(world *World, deltaTime float64) {
	Iter2[T1, T2](world).ForEach(func(entity Entity, comp1 *T1, comp2 *T2) {
		s.updateFunc(world, deltaTime, entity, comp1, comp2)
	})
}

// System3 is a convenience system that processes entities with three component types
type System3[T1, T2, T3 any] struct {
	*BaseSystem
	updateFunc func(*World, float64, Entity, *T1, *T2, *T3)
}

// NewSystem3 creates a new three-component system
func NewSystem3[T1, T2, T3 any](name string, updateFunc func(*World, float64, Entity, *T1, *T2, *T3)) *System3[T1, T2, T3] {
	return &System3[T1, T2, T3]{
		BaseSystem: NewBaseSystem(name),
		updateFunc: updateFunc,
	}
}

// Update processes all entities with the required components
func (s *System3[T1, T2, T3]) Update(world *World, deltaTime float64) {
	Iter3[T1, T2, T3](world).ForEach(func(entity Entity, comp1 *T1, comp2 *T2, comp3 *T3) {
		s.updateFunc(world, deltaTime, entity, comp1, comp2, comp3)
	})
}
