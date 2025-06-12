# PECS-GO: Sparse Set-Based Entity Component System

A high-performance Entity Component System (ECS) implementation in Go using sparse sets, inspired by EnTT and modern ECS architectures.

## Features

### Core ECS Architecture
- **Sparse Set-based Storage**: O(1) component access and iteration with cache-friendly data layout
- **Generational Entity IDs**: 20-bit entity index + 12-bit generation for safe entity recycling
- **Type-safe Components**: Full Go generics support for compile-time type safety
- **Advanced Query System**: Complex queries with include/exclude/any patterns
- **System Management**: Flexible system registration and execution order

### Performance Optimizations
- **Cache-efficient Iteration**: Dense arrays for optimal memory access patterns
- **O(1) Component Operations**: Add, remove, and access components in constant time
- **Minimal Memory Overhead**: Sparse sets minimize memory usage for sparse component data
- **Entity Recycling**: Reuse entity IDs with generational safety to prevent use-after-free

### Query Capabilities
- **Multi-component Queries**: Query entities with specific component combinations
- **Exclusion Filters**: Exclude entities with certain components
- **Any-of Queries**: Match entities with any of the specified components
- **Iterator Types**: Specialized iterators for 1, 2, and 3 component queries

## Architecture

### Core Components

```
ecs/
├── entity.go           # Entity management with generational IDs
├── sparse_set.go       # Core sparse set data structure
├── component_storage.go # Type-safe component pools using generics
├── query.go           # Advanced query system with multiple patterns
├── world.go           # Central coordinator for entities/components/systems
└── system.go          # System management and execution
```

### Entity Management
- **Generational IDs**: Prevents accessing destroyed entities
- **Efficient Recycling**: Reuse entity slots while maintaining safety
- **Packed Storage**: Dense entity arrays for fast iteration

### Component Storage
- **Sparse Sets**: Each component type uses its own sparse set
- **Type Registry**: Automatic component type registration
- **Generic Interface**: Type-safe component access via Go generics

### Query System
- **Builder Pattern**: Fluent API for constructing complex queries
- **Multiple Patterns**: Support for AND, OR, NOT operations
- **Optimized Iteration**: Specialized iterators for common use cases

## Examples

### Basic Usage

```go
// Create world
world := ecs.NewWorld()

// Create entity
entity := world.CreateEntity()

// Add components
ecs.AddComponent(world, entity, Position{X: 10, Y: 20})
ecs.AddComponent(world, entity, Velocity{X: 1, Y: 0})

// Query entities
ecs.Iter2[Position, Velocity](world).ForEach(func(entity ecs.Entity, pos *Position, vel *Velocity) {
    pos.X += vel.X
    pos.Y += vel.Y
})
```

### Complex Queries

```go
// Find entities with Position but without Velocity
query := world.Query()
ecs.With[Position](query)
ecs.Without[Velocity](query)
result := query.Build()

// Find entities with either Player or Enemy components
query := world.Query()
ecs.WithAny[Player](query)
ecs.WithAny[Enemy](query)
combatants := query.Build()
```

### System Implementation

```go
type MovementSystem struct {
    *ecs.BaseSystem
}

func (ms *MovementSystem) Update(world *ecs.World, deltaTime float64) {
    ecs.Iter2[Position, Velocity](world).ForEach(func(entity ecs.Entity, pos *Position, vel *Velocity) {
        pos.X += vel.X * deltaTime
        pos.Y += vel.Y * deltaTime
    })
}
```

## Demo Applications

### RPG Example (`examples/rpg/`)
- Demonstrates core ECS concepts with game-like entities
- Shows system interactions and component relationships
- Includes player, enemy, and projectile entities
- Features combat, movement, and debug systems

### Particle Simulation (`examples/particles/`)
- Real-time physics simulation with 2000+ particles
- Native desktop rendering using SDL2 and OpenGL
- Demonstrates high-performance ECS with visual output
- 60+ FPS performance with optimized rendering

## Installation

### Prerequisites

For the particle simulation, install system dependencies:

**macOS:**
```bash
brew install pkg-config sdl2
```

**Ubuntu/Debian:**
```bash
sudo apt-get install pkg-config libsdl2-dev
```

### Running Examples

```bash
# RPG example
cd examples/rpg
go run main.go

# Particle simulation
cd examples/particles
go run main.go
```

## Performance Characteristics

- **Entity Creation**: O(1) amortized
- **Component Add/Remove**: O(1)
- **Component Access**: O(1)
- **Query Iteration**: O(n) where n = entities with components
- **Memory Usage**: Minimal overhead, sparse data friendly

## References and Inspiration

### Primary References
- **@dakom's EnTT Gist**: Comprehensive sparse set ECS architecture [overview](https://gist.github.com/dakom/82551fff5d2b843cbe1601bbaff2acbf)
  - Sparse set implementation patterns
  - Entity recycling strategies
  - Performance optimization techniques

- **EnTT Library**: Modern C++ ECS [library](https://github.com/skypjack/entt) by Michele Caini
  - Sparse set-based component storage
  - Type-safe component access
  - Advanced query capabilities

### Academic Papers
- **[Sparse Set Data Structure](https://dl.acm.org/doi/pdf/10.1145/176454.176484)**: Briggs & Torczon (1993)
  - Efficient set operations for sparse data
  - O(1) membership testing and iteration

### ECS Architecture References
- **[gecs](https://github.com/tutumagi/gecs) Library**: Pre-generics Go ECS implementation
  - Comparison baseline for modern Go approaches
  - Interface-based component storage patterns

- **Unity DOTS**: Data-Oriented Technology Stack
  - Archetype-based ECS concepts
  - Performance-oriented design principles

### Performance Optimization Sources
- **Data-Oriented Design**: Mike Acton's principles
  - Cache-friendly data layouts
  - Structure of Arrays (SoA) patterns
  - Memory access optimization

## Technical Details

### Sparse Set Implementation
- **Dense Array**: Packed component data for cache efficiency
- **Sparse Array**: O(1) entity-to-component mapping
- **Swap-and-Pop**: Efficient component removal maintaining density

### Generational Entity IDs
- **20-bit Index**: Supports up to 1M entities
- **12-bit Generation**: Prevents use-after-free with 4096 generations
- **Packed Format**: Single uint32 for efficient storage and comparison

### Memory Layout
```
Entity: [Generation:12][Index:20]
Sparse Set: [Dense Array][Sparse Array][Entity Array]
Component Pool: [Sparse Set][Component Data Array]
```

## Contributing

Contributions are welcome! Areas of interest:
- Performance optimizations
- Additional query patterns
- More example applications
- Documentation improvements
- Benchmark comparisons

## License

MIT License - see LICENSE file for details.
