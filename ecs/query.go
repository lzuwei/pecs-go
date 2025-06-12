package ecs

// QueryResult represents the result of a query operation
type QueryResult struct {
	entities []Entity
	world    *World
}

// NewQueryResult creates a new query result
func NewQueryResult(entities []Entity, world *World) *QueryResult {
	return &QueryResult{
		entities: entities,
		world:    world,
	}
}

// Entities returns the entities that match the query
func (qr *QueryResult) Entities() []Entity {
	return qr.entities
}

// Size returns the number of entities in the result
func (qr *QueryResult) Size() int {
	return len(qr.entities)
}

// Empty checks if the result is empty
func (qr *QueryResult) Empty() bool {
	return len(qr.entities) == 0
}

// ForEach iterates over all entities in the result
func (qr *QueryResult) ForEach(fn func(Entity)) {
	for _, entity := range qr.entities {
		fn(entity)
	}
}

// Query provides a fluent interface for querying entities
type Query struct {
	world      *World
	include    []ComponentID
	exclude    []ComponentID
	includeAny []ComponentID
	excludeAny []ComponentID
}

// NewQuery creates a new query for the world
func NewQuery(world *World) *Query {
	return &Query{
		world:      world,
		include:    make([]ComponentID, 0),
		exclude:    make([]ComponentID, 0),
		includeAny: make([]ComponentID, 0),
		excludeAny: make([]ComponentID, 0),
	}
}

// With adds component types that entities must have (AND operation)
func With[T any](q *Query) *Query {
	id := Register[T](q.world.componentRegistry)
	q.include = append(q.include, id)
	return q
}

// Without adds component types that entities must not have (NOT operation)
func Without[T any](q *Query) *Query {
	id := Register[T](q.world.componentRegistry)
	q.exclude = append(q.exclude, id)
	return q
}

// WithAny adds component types where entities must have at least one (OR operation)
func WithAny[T any](q *Query) *Query {
	id := Register[T](q.world.componentRegistry)
	q.includeAny = append(q.includeAny, id)
	return q
}

// WithoutAny adds component types where entities must not have any (NOR operation)
func WithoutAny[T any](q *Query) *Query {
	id := Register[T](q.world.componentRegistry)
	q.excludeAny = append(q.excludeAny, id)
	return q
}

// Build executes the query and returns the results
func (q *Query) Build() *QueryResult {
	if len(q.include) == 0 && len(q.includeAny) == 0 {
		// No inclusion criteria, return empty result
		return NewQueryResult([]Entity{}, q.world)
	}

	var candidates []Entity

	// Start with the smallest required component set
	if len(q.include) > 0 {
		// Find the smallest component pool to start with
		smallestSize := int(^uint(0) >> 1) // Max int
		var smallestStorage IComponentStorage

		for _, id := range q.include {
			if storage, exists := q.world.componentRegistry.GetStorageByID(id); exists {
				if storage.Size() < smallestSize {
					smallestSize = storage.Size()
					smallestStorage = storage
				}
			}
		}

		if smallestStorage != nil {
			candidates = smallestStorage.Entities().Data()
		} else {
			return NewQueryResult([]Entity{}, q.world)
		}
	} else if len(q.includeAny) > 0 {
		// Collect entities from any of the includeAny components
		entitySet := make(map[Entity]bool)
		for _, id := range q.includeAny {
			if storage, exists := q.world.componentRegistry.GetStorageByID(id); exists {
				entities := storage.Entities().Data()
				for _, entity := range entities {
					entitySet[entity] = true
				}
			}
		}

		candidates = make([]Entity, 0, len(entitySet))
		for entity := range entitySet {
			candidates = append(candidates, entity)
		}
	}

	// Filter candidates
	result := make([]Entity, 0, len(candidates))

	for _, entity := range candidates {
		if q.matchesEntity(entity) {
			result = append(result, entity)
		}
	}

	return NewQueryResult(result, q.world)
}

// matchesEntity checks if an entity matches all query criteria
func (q *Query) matchesEntity(entity Entity) bool {
	// Check include (must have ALL)
	for _, id := range q.include {
		if storage, exists := q.world.componentRegistry.GetStorageByID(id); exists {
			if !storage.Contains(entity) {
				return false
			}
		} else {
			return false // Component type not registered
		}
	}

	// Check exclude (must have NONE)
	for _, id := range q.exclude {
		if storage, exists := q.world.componentRegistry.GetStorageByID(id); exists {
			if storage.Contains(entity) {
				return false
			}
		}
	}

	// Check includeAny (must have AT LEAST ONE)
	if len(q.includeAny) > 0 {
		hasAny := false
		for _, id := range q.includeAny {
			if storage, exists := q.world.componentRegistry.GetStorageByID(id); exists {
				if storage.Contains(entity) {
					hasAny = true
					break
				}
			}
		}
		if !hasAny {
			return false
		}
	}

	// Check excludeAny (must have NONE)
	for _, id := range q.excludeAny {
		if storage, exists := q.world.componentRegistry.GetStorageByID(id); exists {
			if storage.Contains(entity) {
				return false
			}
		}
	}

	return true
}

// Iterator provides convenient iteration over query results with components
type Iterator1[T1 any] struct {
	result         *QueryResult
	component1Pool *ComponentPool[T1]
}

// NewIterator1 creates a new single-component iterator
func NewIterator1[T1 any](world *World) *Iterator1[T1] {
	pool1, _ := GetStorage[T1](world.componentRegistry)

	query := NewQuery(world)
	With[T1](query)
	result := query.Build()

	return &Iterator1[T1]{
		result:         result,
		component1Pool: pool1,
	}
}

// ForEach iterates over entities with their components
func (it *Iterator1[T1]) ForEach(fn func(Entity, *T1)) {
	for _, entity := range it.result.entities {
		if comp1 := it.component1Pool.GetPtr(entity); comp1 != nil {
			fn(entity, comp1)
		}
	}
}

// Iterator2 provides iteration over entities with two component types
type Iterator2[T1, T2 any] struct {
	result         *QueryResult
	component1Pool *ComponentPool[T1]
	component2Pool *ComponentPool[T2]
}

// NewIterator2 creates a new two-component iterator
func NewIterator2[T1, T2 any](world *World) *Iterator2[T1, T2] {
	pool1, _ := GetStorage[T1](world.componentRegistry)
	pool2, _ := GetStorage[T2](world.componentRegistry)

	query := NewQuery(world)
	With[T1](query)
	With[T2](query)
	result := query.Build()

	return &Iterator2[T1, T2]{
		result:         result,
		component1Pool: pool1,
		component2Pool: pool2,
	}
}

// ForEach iterates over entities with their components
func (it *Iterator2[T1, T2]) ForEach(fn func(Entity, *T1, *T2)) {
	for _, entity := range it.result.entities {
		comp1 := it.component1Pool.GetPtr(entity)
		comp2 := it.component2Pool.GetPtr(entity)
		if comp1 != nil && comp2 != nil {
			fn(entity, comp1, comp2)
		}
	}
}

// Iterator3 provides iteration over entities with three component types
type Iterator3[T1, T2, T3 any] struct {
	result         *QueryResult
	component1Pool *ComponentPool[T1]
	component2Pool *ComponentPool[T2]
	component3Pool *ComponentPool[T3]
}

// NewIterator3 creates a new three-component iterator
func NewIterator3[T1, T2, T3 any](world *World) *Iterator3[T1, T2, T3] {
	pool1, _ := GetStorage[T1](world.componentRegistry)
	pool2, _ := GetStorage[T2](world.componentRegistry)
	pool3, _ := GetStorage[T3](world.componentRegistry)

	query := NewQuery(world)
	With[T1](query)
	With[T2](query)
	With[T3](query)
	result := query.Build()

	return &Iterator3[T1, T2, T3]{
		result:         result,
		component1Pool: pool1,
		component2Pool: pool2,
		component3Pool: pool3,
	}
}

// ForEach iterates over entities with their components
func (it *Iterator3[T1, T2, T3]) ForEach(fn func(Entity, *T1, *T2, *T3)) {
	for _, entity := range it.result.entities {
		comp1 := it.component1Pool.GetPtr(entity)
		comp2 := it.component2Pool.GetPtr(entity)
		comp3 := it.component3Pool.GetPtr(entity)
		if comp1 != nil && comp2 != nil && comp3 != nil {
			fn(entity, comp1, comp2, comp3)
		}
	}
}

// ViewBuilder provides a more flexible way to build queries
type ViewBuilder struct {
	world *World
	query *Query
}

// NewViewBuilder creates a new view builder
func NewViewBuilder(world *World) *ViewBuilder {
	return &ViewBuilder{
		world: world,
		query: NewQuery(world),
	}
}

// Include adds required components (AND)
func (vb *ViewBuilder) Include(componentIDs ...ComponentID) *ViewBuilder {
	vb.query.include = append(vb.query.include, componentIDs...)
	return vb
}

// Exclude adds forbidden components (NOT)
func (vb *ViewBuilder) Exclude(componentIDs ...ComponentID) *ViewBuilder {
	vb.query.exclude = append(vb.query.exclude, componentIDs...)
	return vb
}

// IncludeAny adds components where at least one must be present (OR)
func (vb *ViewBuilder) IncludeAny(componentIDs ...ComponentID) *ViewBuilder {
	vb.query.includeAny = append(vb.query.includeAny, componentIDs...)
	return vb
}

// ExcludeAny adds components where none must be present (NOR)
func (vb *ViewBuilder) ExcludeAny(componentIDs ...ComponentID) *ViewBuilder {
	vb.query.excludeAny = append(vb.query.excludeAny, componentIDs...)
	return vb
}

// Build executes the query
func (vb *ViewBuilder) Build() *QueryResult {
	return vb.query.Build()
}
