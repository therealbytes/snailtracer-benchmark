package main

import (
	"math/big"

	"github.com/ethereum/go-ethereum/concrete/api"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/ethereum/go-ethereum/tinygo"
	"github.com/therealbytes/concrete-snailtrace/snailtracer"
)

type snailtracerPrecompile struct {
	scene *snailtracer.Scene
	lib.BlankPrecompile
}

func (t *snailtracerPrecompile) Run(env api.Environment, input []byte) ([]byte, error) {
	if t.scene == nil {
		t.scene = snailtracer.NewBenchmarkScene(0) // Wasmer won't work unless we do this every time
	}
	color := snailtracer.NewVector(0, 0, 0)
	color = color.Add(t.scene.Trace(512, 384, 8))
	color = color.Add(t.scene.Trace(325, 540, 8))
	color = color.Add(t.scene.Trace(600, 600, 8))
	color = color.Add(t.scene.Trace(522, 524, 8))
	color = color.ScaleDiv(big.NewInt(4))

	cr := color.X.Int64()
	cg := color.Y.Int64()
	cb := color.Z.Int64()

	result := make([]byte, 96)
	result[0] = byte(cr)
	result[32] = byte(cg)
	result[64] = byte(cb)

	return result, nil
}

func init() {
	tinygo.WasmWrap(&snailtracerPrecompile{
		scene: snailtracer.NewBenchmarkScene(0),
	})
}

// main is REQUIRED for TinyGo to compile to WASM
func main() {}
