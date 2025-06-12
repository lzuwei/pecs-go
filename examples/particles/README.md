# Particle Simulation

A real-time particle simulation using the PECS-GO ECS system and tfriedel6/canvas for native desktop rendering.

## Features

- 5000 bouncing particles with random colors and velocities
- Real-time physics simulation at 60+ FPS
- Native desktop window using SDL2 and OpenGL
- Demonstrates ECS architecture with MovementSystem and BounceSystem

## Prerequisites

Before running the particle simulation, you need to install the required system dependencies:

### macOS

```bash
brew install pkg-config sdl2
```

### Ubuntu/Debian

```bash
sudo apt-get update
sudo apt-get install pkg-config libsdl2-dev
```

### CentOS/RHEL/Fedora

```bash
# CentOS/RHEL
sudo yum install pkgconfig SDL2-devel

# Fedora
sudo dnf install pkgconf-devel SDL2-devel
```

### Windows

1. Install [pkg-config for Windows](http://ftp.gnome.org/pub/gnome/binaries/win32/dependencies/)
2. Download SDL2 development libraries from [libsdl.org](https://www.libsdl.org/download-2.0.php)
3. Extract and add to your system PATH

## Running the Simulation

```bash
cd examples/particles
go run main.go
```

## Controls

- Close the window to exit the simulation
- The simulation runs automatically with physics and collision detection

## Configuration

You can modify the following constants in `main.go`:

- `ParticleCount`: Number of particles (default: 1000)
- `CanvasWidth`/`CanvasHeight`: Window dimensions
- `MaxVelocity`/`MinVelocity`: Particle speed range
- `ParticleSize`: Size of each particle

## Troubleshooting

If you encounter build errors:

1. **"pkg-config: executable file not found"**: Install pkg-config using the instructions above
2. **SDL2 not found**: Make sure SDL2 development libraries are installed
3. **Window not showing on macOS**: This is normal - the window should appear after a moment

The simulation outputs FPS information to the console to confirm it's running properly. 