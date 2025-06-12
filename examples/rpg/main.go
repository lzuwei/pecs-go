package main

import (
	"fmt"
	"math"
	"pecs-go/ecs"
)

// Example components
type Position struct {
	X, Y float64
}

type Velocity struct {
	X, Y float64
}

type Health struct {
	Current, Max int
}

type Name struct {
	Value string
}

type Damage struct {
	Amount int
}

type Player struct {
	Level int
}

type Enemy struct {
	AggroRange float64
}

// MovementSystem moves entities with Position and Velocity components
type MovementSystem struct {
	*ecs.BaseSystem
}

func NewMovementSystem() *MovementSystem {
	return &MovementSystem{
		BaseSystem: ecs.NewBaseSystem("MovementSystem"),
	}
}

func (ms *MovementSystem) Update(world *ecs.World, deltaTime float64) {
	ecs.Iter2[Position, Velocity](world).ForEach(func(entity ecs.Entity, pos *Position, vel *Velocity) {
		pos.X += vel.X * deltaTime
		pos.Y += vel.Y * deltaTime
	})
}

// CombatSystem handles combat between entities
type CombatSystem struct {
	*ecs.BaseSystem
}

func NewCombatSystem() *CombatSystem {
	return &CombatSystem{
		BaseSystem: ecs.NewBaseSystem("CombatSystem"),
	}
}

func (cs *CombatSystem) Update(world *ecs.World, deltaTime float64) {
	// Process damage dealers
	query := world.Query()
	ecs.With[Damage](query)
	damageDealer := query.Build()

	// Apply damage to entities with Health
	ecs.Iter1[Health](world).ForEach(func(entity ecs.Entity, health *Health) {
		// For simplicity, apply damage from all damage dealers
		for _, dealer := range damageDealer.Entities() {
			if dealer != entity { // Don't damage self
				if damage, hasDamage := ecs.GetComponent[Damage](world, dealer); hasDamage {
					health.Current -= damage.Amount
					if health.Current <= 0 {
						health.Current = 0
						fmt.Printf("Entity %s died!\n", entity)
					}
				}
			}
		}
	})
}

// DebugSystem prints entity information
type DebugSystem struct {
	*ecs.BaseSystem
	printInterval float64
	timer         float64
}

func NewDebugSystem(printInterval float64) *DebugSystem {
	return &DebugSystem{
		BaseSystem:    ecs.NewBaseSystem("DebugSystem"),
		printInterval: printInterval,
		timer:         0,
	}
}

func (ds *DebugSystem) Update(world *ecs.World, deltaTime float64) {
	ds.timer += deltaTime
	if ds.timer < ds.printInterval {
		return
	}
	ds.timer = 0

	fmt.Println("=== World State ===")

	// Print entities with names
	ecs.Iter1[Name](world).ForEach(func(entity ecs.Entity, name *Name) {
		fmt.Printf("Entity %s [%s]:", entity, name.Value)

		if pos, hasPos := ecs.GetComponent[Position](world, entity); hasPos {
			fmt.Printf(" Pos(%.1f,%.1f)", pos.X, pos.Y)
		}

		if vel, hasVel := ecs.GetComponent[Velocity](world, entity); hasVel {
			fmt.Printf(" Vel(%.1f,%.1f)", vel.X, vel.Y)
		}

		if health, hasHealth := ecs.GetComponent[Health](world, entity); hasHealth {
			fmt.Printf(" HP(%d/%d)", health.Current, health.Max)
		}

		if ecs.HasComponent[Player](world, entity) {
			fmt.Printf(" [PLAYER]")
		}

		if ecs.HasComponent[Enemy](world, entity) {
			fmt.Printf(" [ENEMY]")
		}

		fmt.Println()
	})

	stats := world.Stats()
	fmt.Printf("Stats: %d entities, %d component types, %d total components, %d systems\n\n",
		stats.EntityCount, stats.ComponentTypes, stats.TotalComponents, stats.SystemCount)
}

// ExampleUsage demonstrates the ECS functionality
func ExampleUsage() {
	fmt.Println("=== PECS-GO: Sparse Set-Based ECS Example ===\n")

	// Create world
	world := ecs.NewWorld()

	// Add systems
	world.AddSystem(NewMovementSystem())
	world.AddSystem(NewCombatSystem())
	world.AddSystem(NewDebugSystem(2.0)) // Print every 2 seconds

	// Create player entity
	player := world.CreateEntity()
	ecs.AddComponent(world, player, Name{Value: "Hero"})
	ecs.AddComponent(world, player, Position{X: 0, Y: 0})
	ecs.AddComponent(world, player, Velocity{X: 1, Y: 0.5})
	ecs.AddComponent(world, player, Health{Current: 100, Max: 100})
	ecs.AddComponent(world, player, Player{Level: 1})

	// Create enemy entities
	for i := 0; i < 3; i++ {
		enemy := world.CreateEntity()
		ecs.AddComponent(world, enemy, Name{Value: fmt.Sprintf("Orc%d", i+1)})
		ecs.AddComponent(world, enemy, Position{X: float64(10 + i*5), Y: float64(i * 2)})
		ecs.AddComponent(world, enemy, Velocity{X: -0.5, Y: 0})
		ecs.AddComponent(world, enemy, Health{Current: 50, Max: 50})
		ecs.AddComponent(world, enemy, Enemy{AggroRange: 5.0})
		if i == 0 {
			ecs.AddComponent(world, enemy, Damage{Amount: 10}) // First enemy can attack
		}
	}

	// Create some projectiles (moving damage dealers)
	for i := 0; i < 2; i++ {
		projectile := world.CreateEntity()
		ecs.AddComponent(world, projectile, Name{Value: fmt.Sprintf("Arrow%d", i+1)})
		ecs.AddComponent(world, projectile, Position{X: float64(i * 3), Y: 1})
		ecs.AddComponent(world, projectile, Velocity{X: 2, Y: 0})
		ecs.AddComponent(world, projectile, Damage{Amount: 15})
	}

	fmt.Println("Initial state:")
	world.Update(0.0)

	// Simulate some time steps
	timeSteps := []float64{1.0, 1.0, 1.0, 2.0, 1.0}
	for i, dt := range timeSteps {
		fmt.Printf("=== Tick %d (dt=%.1f) ===\n", i+1, dt)
		world.Update(dt)
	}

	// Demonstrate entity recycling
	fmt.Println("=== Testing Entity Recycling ===")

	// Get the first enemy and destroy it
	query := world.Query()
	ecs.With[Enemy](query)
	enemies := query.Build()

	if !enemies.Empty() {
		firstEnemy := enemies.Entities()[0]
		fmt.Printf("Destroying entity: %s\n", firstEnemy)
		world.DestroyEntity(firstEnemy)

		// Create a new entity - should reuse the ID
		newEntity := world.CreateEntity()
		ecs.AddComponent(world, newEntity, Name{Value: "Recycled"})
		ecs.AddComponent(world, newEntity, Position{X: 999, Y: 999})
		fmt.Printf("Created new entity: %s (recycled ID expected)\n", newEntity)
	}

	world.Update(0.1)

	// Demonstrate complex queries
	fmt.Println("=== Complex Query Examples ===")

	// Find all entities with Position but without Velocity (stationary entities)
	stationaryQuery := world.Query()
	ecs.With[Position](stationaryQuery)
	ecs.Without[Velocity](stationaryQuery)
	stationary := stationaryQuery.Build()
	fmt.Printf("Stationary entities: %d\n", stationary.Size())

	// Find all entities with either Player or Enemy components
	combatantQuery := world.Query()
	ecs.WithAny[Player](combatantQuery)
	ecs.WithAny[Enemy](combatantQuery)
	combatants := combatantQuery.Build()
	fmt.Printf("Combat entities (Player or Enemy): %d\n", combatants.Size())

	// Find all entities with Health and Position (living, positioned entities)
	livingQuery := world.Query()
	ecs.With[Health](livingQuery)
	ecs.With[Position](livingQuery)
	living := livingQuery.Build()
	fmt.Printf("Living positioned entities: %d\n", living.Size())

	// Demonstrate the convenience systems
	fmt.Println("\n=== Convenience System Demo ===")

	// Create a system using System2 helper
	healingSystem := ecs.NewSystem2[Health, Player]("HealingSystem",
		func(world *ecs.World, deltaTime float64, entity ecs.Entity, health *Health, player *Player) {
			// Players regenerate health
			if health.Current < health.Max {
				healAmount := int(float64(player.Level) * deltaTime * 2)
				health.Current += healAmount
				if health.Current > health.Max {
					health.Current = health.Max
				}
			}
		})

	world.AddSystem(healingSystem)
	fmt.Println("Added healing system for players")
	world.Update(3.0) // Run for 3 seconds to see healing

	fmt.Println("\n=== Final Statistics ===")
	finalStats := world.Stats()
	fmt.Printf("Final state: %d entities, %d component types, %d total components, %d systems\n",
		finalStats.EntityCount, finalStats.ComponentTypes, finalStats.TotalComponents, finalStats.SystemCount)
}

// BenchmarkExample demonstrates performance characteristics
func BenchmarkExample() {
	fmt.Println("=== Performance Benchmark Example ===")

	world := ecs.NewWorld()

	// Create many entities
	numEntities := 10000
	fmt.Printf("Creating %d entities...\n", numEntities)

	for i := 0; i < numEntities; i++ {
		entity := world.CreateEntity()
		ecs.AddComponent(world, entity, Position{
			X: math.Sin(float64(i)) * 100,
			Y: math.Cos(float64(i)) * 100,
		})
		ecs.AddComponent(world, entity, Velocity{
			X: (float64(i%10) - 5) * 0.1,
			Y: (float64(i%7) - 3) * 0.1,
		})

		if i%100 == 0 {
			ecs.AddComponent(world, entity, Health{Current: 100, Max: 100})
		}

		if i%1000 == 0 {
			ecs.AddComponent(world, entity, Name{Value: fmt.Sprintf("Entity%d", i)})
		}
	}

	// Add movement system
	world.AddSystem(NewMovementSystem())

	// Benchmark queries
	fmt.Println("Benchmarking queries...")

	// Query all entities with Position and Velocity
	query1 := world.Query()
	ecs.With[Position](query1)
	ecs.With[Velocity](query1)
	result1 := query1.Build()
	fmt.Printf("Entities with Position+Velocity: %d\n", result1.Size())

	// Query entities with Health
	query2 := world.Query()
	ecs.With[Health](query2)
	result2 := query2.Build()
	fmt.Printf("Entities with Health: %d\n", result2.Size())

	// Benchmark system updates
	fmt.Println("Running movement system...")
	world.Update(0.016) // 60 FPS frame time

	fmt.Printf("Benchmark complete. Final entity count: %d\n", world.Stats().EntityCount)
}

func main() {
	// Run the comprehensive example
	ExampleUsage()

	// Uncomment to run performance benchmark
	// BenchmarkExample()
}
