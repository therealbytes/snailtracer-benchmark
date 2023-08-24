package snailtracer

import "testing"

func BenchmarkSnailtracer(b *testing.B) {
	s := newScene(1024, 768)

	// Configure the scene for benchmarking
	// s.width = 1024
	// s.height = 768

	s.deltaX = &Vector{s.width * 513500 / s.height, 0, 0}
	s.deltaY = s.deltaX.cross(s.camera.direction).norm().scaleMul(513500).scaleDiv(1000000)

	// Trace a few pixels and collect their colors (sanity check)
	color := &Vector{0, 0, 0}

	color = color.add(s.trace(512, 384, 8)) // Flat diffuse surface, opposite wall
	color = color.add(s.trace(325, 540, 8)) // Reflective surface mirroring left wall
	color = color.add(s.trace(600, 600, 8)) // Refractive surface reflecting right wall
	color = color.add(s.trace(522, 524, 8)) // Reflective surface mirroring the refractive surface reflecting the light
	color = color.scaleDiv(4)

	b.Log(color)
}
