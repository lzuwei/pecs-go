package main

import (
	"fmt"
	"math/rand"
	"time"

	"pecs-go/ecs"

	"github.com/tfriedel6/canvas"
	"github.com/tfriedel6/canvas/sdlcanvas"
)

// Configuration
const (
	// Canvas dimensions
	CanvasWidth  = 1600
	CanvasHeight = 1200

	// Particle settings
	ParticleCount = 10000 // Reduced for better performance
	ParticleSize  = 3.0   // Smaller size for better performance

	// Physics settings
	MaxVelocity = 200.0 // Maximum velocity in pixels per second
	MinVelocity = 50.0  // Minimum velocity in pixels per second
)

// Components
type Position struct {
	X, Y float64
}

type Velocity struct {
	X, Y float64
}

type Color struct {
	R, G, B uint8
}

type Particle struct {
	Size float64
}

// MovementSystem handles particle movement
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

// BounceSystem handles collision with canvas edges
type BounceSystem struct {
	*ecs.BaseSystem
	canvas *canvas.Canvas
}

func NewBounceSystem(cv *canvas.Canvas) *BounceSystem {
	return &BounceSystem{
		BaseSystem: ecs.NewBaseSystem("BounceSystem"),
		canvas:     cv,
	}
}

func (bs *BounceSystem) Update(world *ecs.World, deltaTime float64) {
	// Get current canvas dimensions
	width := float64(bs.canvas.Width())
	height := float64(bs.canvas.Height())

	ecs.Iter3[Position, Velocity, Particle](world).ForEach(func(entity ecs.Entity, pos *Position, vel *Velocity, particle *Particle) {
		radius := particle.Size / 2

		// Bounce off left and right edges
		if pos.X-radius <= 0 {
			pos.X = radius
			vel.X = -vel.X
		} else if pos.X+radius >= width {
			pos.X = width - radius
			vel.X = -vel.X
		}

		// Bounce off top and bottom edges
		if pos.Y-radius <= 0 {
			pos.Y = radius
			vel.Y = -vel.Y
		} else if pos.Y+radius >= height {
			pos.Y = height - radius
			vel.Y = -vel.Y
		}
	})
}

// ParticleSimulation manages the entire simulation
type ParticleSimulation struct {
	world      *ecs.World
	lastTime   time.Time
	frameCount int
	fpsTimer   time.Time
	canvas     *canvas.Canvas
}

func NewParticleSimulation(cv *canvas.Canvas) *ParticleSimulation {
	world := ecs.NewWorld()

	// Add systems
	world.AddSystem(NewMovementSystem())
	world.AddSystem(NewBounceSystem(cv))

	sim := &ParticleSimulation{
		world:    world,
		lastTime: time.Now(),
		fpsTimer: time.Now(),
		canvas:   cv,
	}

	// Create particles
	sim.createParticles()

	return sim
}

func (ps *ParticleSimulation) createParticles() {
	rand.Seed(time.Now().UnixNano())

	// Get actual canvas dimensions
	canvasWidth := float64(ps.canvas.Width())
	canvasHeight := float64(ps.canvas.Height())

	for i := 0; i < ParticleCount; i++ {
		entity := ps.world.CreateEntity()

		// Random position (ensuring particles start within bounds)
		margin := ParticleSize / 2
		x := margin + rand.Float64()*(canvasWidth-2*margin)
		y := margin + rand.Float64()*(canvasHeight-2*margin)
		ecs.AddComponent(ps.world, entity, Position{X: x, Y: y})

		// Random velocity
		vx := MinVelocity + rand.Float64()*(MaxVelocity-MinVelocity)
		vy := MinVelocity + rand.Float64()*(MaxVelocity-MinVelocity)
		if rand.Intn(2) == 0 {
			vx = -vx
		}
		if rand.Intn(2) == 0 {
			vy = -vy
		}
		ecs.AddComponent(ps.world, entity, Velocity{X: vx, Y: vy})

		// Random color
		r := uint8(rand.Intn(256))
		g := uint8(rand.Intn(256))
		b := uint8(rand.Intn(256))
		ecs.AddComponent(ps.world, entity, Color{R: r, G: g, B: b})

		// Fixed size
		ecs.AddComponent(ps.world, entity, Particle{Size: ParticleSize})
	}
}

func (ps *ParticleSimulation) Update(deltaTime float64) {
	// Update ECS world
	ps.world.Update(deltaTime)

	// Update FPS counter
	ps.frameCount++
	if time.Since(ps.fpsTimer) >= time.Second {
		fmt.Printf("FPS: %d, Particles: %d\n", ps.frameCount, ParticleCount)
		ps.frameCount = 0
		ps.fpsTimer = time.Now()
	}
}

func (ps *ParticleSimulation) Render(cv *canvas.Canvas) {
	w, h := float64(cv.Width()), float64(cv.Height())

	// Clear canvas with black background
	cv.SetFillStyle("#000000")
	cv.FillRect(0, 0, w, h)

	// Use a limited color palette for efficient batching
	colors := [][3]int{
		{255, 100, 100}, // Red
		{100, 255, 100}, // Green
		{100, 100, 255}, // Blue
		{255, 255, 100}, // Yellow
		{255, 100, 255}, // Magenta
		{100, 255, 255}, // Cyan
	}

	// Collect positions by color bucket
	colorBuckets := make([][]Position, len(colors))

	ecs.Iter3[Position, Color, Particle](ps.world).ForEach(func(entity ecs.Entity, pos *Position, color *Color, particle *Particle) {
		// Map to color bucket based on original color
		bucket := (int(color.R) + int(color.G) + int(color.B)) % len(colors)
		colorBuckets[bucket] = append(colorBuckets[bucket], *pos)
	})

	// Draw each color bucket
	for i, positions := range colorBuckets {
		if len(positions) > 0 {
			cv.SetFillStyle(colors[i][0], colors[i][1], colors[i][2])

			// Draw all particles of this color
			for _, pos := range positions {
				cv.FillRect(pos.X-1, pos.Y-1, 2, 2)
			}
		}
	}
}

func main() {
	fmt.Printf("Starting Particle Simulation with %d particles...\n", ParticleCount)

	// Create SDL window and canvas
	wnd, cv, err := sdlcanvas.CreateWindow(CanvasWidth, CanvasHeight, "Particle Simulation - PECS-GO ECS")
	if err != nil {
		panic(err)
	}
	defer wnd.Destroy()

	// Create simulation
	sim := NewParticleSimulation(cv)

	fmt.Printf("Particle simulation running with %d particles\n", ParticleCount)
	fmt.Printf("Close the window to exit\n")

	// Main loop
	wnd.MainLoop(func() {
		// Calculate delta time
		now := time.Now()
		deltaTime := now.Sub(sim.lastTime).Seconds()
		sim.lastTime = now

		// Limit delta time to prevent large jumps
		if deltaTime > 1.0/30.0 { // Cap at 30 FPS minimum
			deltaTime = 1.0 / 30.0
		}

		// Update simulation
		sim.Update(deltaTime)

		// Render
		sim.Render(cv)
	})
}
