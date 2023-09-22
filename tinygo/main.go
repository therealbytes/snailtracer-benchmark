package main

import (
	"github.com/holiman/uint256"
	"github.com/therealbytes/snailtracer-benchmark/snailtracer"
)

var (
	scene *snailtracer.Scene
)

//export run
func run(seed int32) int32 {
	if scene == nil {
		// Global variables behave unexpectedly in Wasmer, so we need to initialize
		// the scene here.
		scene = snailtracer.NewBenchmarkScene(0, int(seed))
	}

	color := snailtracer.NewVector(0, 0, 0)
	color = color.Add(scene.Trace(512, 384, 8))
	color = color.Add(scene.Trace(325, 540, 8))
	color = color.Add(scene.Trace(600, 600, 8))
	color = color.Add(scene.Trace(522, 524, 8))
	color = color.ScaleDiv(uint256.NewInt(4))

	cr := color.X.Uint64() & 0xff
	cg := color.Y.Uint64() & 0xff
	cb := color.Z.Uint64() & 0xff

	// return int32(cr) + int32(cg), int32(cb)
	return int32(cr<<16 + cg<<8 + cb)
}

// main is REQUIRED for TinyGo to compile to WASM
func main() {}
