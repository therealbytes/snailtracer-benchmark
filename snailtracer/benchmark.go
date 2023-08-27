package snailtracer

import (
	"math/big"
)

func NewBenchmarkScene() *Scene {
	s := newScene(1024/2, 768/2)
	// s := newScene(1024, 768)

	s.camera = &Ray{
		origin:    NewVector(50000000, 50000000, 295600000),
		direction: (NewVector(0, -42612, -1000000)).Norm(),
	}
	s.deltaX = NewVector(s.width*513500/s.height, 0, 0)
	s.deltaY = s.deltaX.Cross(s.camera.direction).Norm().
		ScaleMul(big.NewInt(513500)).
		ScaleDiv(big.NewInt(1000000))

	s.spheres = []*Sphere{
		{big.NewInt(100000000000), NewVector(100001000000, 40800000, 81600000), NewVector(0, 0, 0), NewVector(750000, 250000, 250000), DiffuseMaterial},
		{big.NewInt(100000000000), NewVector(-99901000000, 40800000, 81600000), NewVector(0, 0, 0), NewVector(250000, 250000, 750000), DiffuseMaterial},
		{big.NewInt(100000000000), NewVector(50000000, 40800000, 100000000000), NewVector(0, 0, 0), NewVector(750000, 750000, 750000), DiffuseMaterial},
		{big.NewInt(100000000000), NewVector(50000000, 40800000, -99830000000), NewVector(0, 0, 0), NewVector(0, 0, 0), DiffuseMaterial},
		{big.NewInt(100000000000), NewVector(50000000, 100000000000, 81600000), NewVector(0, 0, 0), NewVector(750000, 750000, 750000), DiffuseMaterial},
		{big.NewInt(100000000000), NewVector(50000000, -99918400000, 81600000), NewVector(0, 0, 0), NewVector(750000, 750000, 750000), DiffuseMaterial},
		{big.NewInt(16500000), NewVector(27000000, 16500000, 47000000), NewVector(0, 0, 0), NewVector(999000, 999000, 999000), SpecularMaterial},
		{big.NewInt(600000000), NewVector(50000000, 681330000, 81600000), NewVector(12000000, 12000000, 12000000), NewVector(0, 0, 0), DiffuseMaterial},
	}

	s.triangles = []*Triangle{
		{NewVector(56500000, 25740000, 78000000), NewVector(73000000, 25740000, 94500000), NewVector(73000000, 49500000, 78000000), NewVector(0, 0, 0), NewVector(0, 0, 0), NewVector(999000, 999000, 999000), SpecularMaterial},
		{NewVector(56500000, 23760000, 78000000), NewVector(73000000, 0, 78000000), NewVector(73000000, 23760000, 94500000), NewVector(0, 0, 0), NewVector(0, 0, 0), NewVector(999000, 999000, 999000), SpecularMaterial},
		{NewVector(89500000, 25740000, 78000000), NewVector(73000000, 49500000, 78000000), NewVector(73000000, 25740000, 94500000), NewVector(0, 0, 0), NewVector(0, 0, 0), NewVector(999000, 999000, 999000), SpecularMaterial},
		{NewVector(89500000, 23760000, 78000000), NewVector(73000000, 23760000, 94500000), NewVector(73000000, 0, 78000000), NewVector(0, 0, 0), NewVector(0, 0, 0), NewVector(999000, 999000, 999000), SpecularMaterial},
		{NewVector(56500000, 25740000, 78000000), NewVector(73000000, 49500000, 78000000), NewVector(73000000, 25740000, 61500000), NewVector(0, 0, 0), NewVector(0, 0, 0), NewVector(999000, 999000, 999000), SpecularMaterial},
		{NewVector(56500000, 23760000, 78000000), NewVector(73000000, 23760000, 61500000), NewVector(73000000, 0, 78000000), NewVector(0, 0, 0), NewVector(0, 0, 0), NewVector(999000, 999000, 999000), SpecularMaterial},
		{NewVector(89500000, 25740000, 78000000), NewVector(73000000, 25740000, 61500000), NewVector(73000000, 49500000, 78000000), NewVector(0, 0, 0), NewVector(0, 0, 0), NewVector(999000, 999000, 999000), SpecularMaterial},
		{NewVector(89500000, 23760000, 78000000), NewVector(73000000, 0, 78000000), NewVector(73000000, 23760000, 61500000), NewVector(0, 0, 0), NewVector(0, 0, 0), NewVector(999000, 999000, 999000), SpecularMaterial},
		{NewVector(56500000, 25740000, 78000000), NewVector(73000000, 25740000, 61500000), NewVector(89500000, 25740000, 78000000), NewVector(0, 0, 0), NewVector(0, 0, 0), NewVector(999000, 999000, 999000), SpecularMaterial},
		{NewVector(56500000, 25740000, 78000000), NewVector(89500000, 25740000, 78000000), NewVector(73000000, 25740000, 94500000), NewVector(0, 0, 0), NewVector(0, 0, 0), NewVector(999000, 999000, 999000), SpecularMaterial},
		{NewVector(56500000, 23760000, 78000000), NewVector(89500000, 23760000, 78000000), NewVector(73000000, 23760000, 61500000), NewVector(0, 0, 0), NewVector(0, 0, 0), NewVector(999000, 999000, 999000), SpecularMaterial},
		{NewVector(56500000, 23760000, 78000000), NewVector(73000000, 23760000, 94500000), NewVector(89500000, 23760000, 78000000), NewVector(0, 0, 0), NewVector(0, 0, 0), NewVector(999000, 999000, 999000), SpecularMaterial},
	}

	// Calculate all the triangle surface normals
	for i := range s.triangles {
		tri := s.triangles[i]
		tri.normal = tri.b.Sub(tri.a).Cross(tri.c.Sub(tri.a)).Norm()
	}

	return s
}
