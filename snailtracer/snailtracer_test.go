//go:build !tinygo

package snailtracer

import (
	_ "embed"
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/precompiles"
	"github.com/ethereum/go-ethereum/concrete/wasm"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/wasmerio/wasmer-go/wasmer"
)

func validResult(r, g, b uint64) bool {
	return r == 17 && g == 17 && b == 53
}

// big.Int
// NativeSnailtracer-8                        32          34383703 ns/op        11873171 B/op     475931 allocs/op
// Parallel4NativeSnailtracer-8               93          12365558 ns/op        11878374 B/op     475990 allocs/op
// EVMSnailtracer-8                            3         447541079 ns/op        41799312 B/op        701 allocs/op
// TinygoSnailtracer/wazero-8                  4         318123902 ns/op             112 B/op          6 allocs/op
// TinygoSnailtracer/wasmer/singlepass-8       4         261384293 ns/op             560 B/op         34 allocs/op
// TinygoSnailtracer/wasmer/cranelift-8        7         153933651 ns/op             561 B/op         34 allocs/op

// uint256.Int
// NativeSnailtracer-8                          73          15308502 ns/op         2794699 B/op      87334 allocs/op
// Parallel4NativeSnailtracer-8                229           5708114 ns/op         2795407 B/op      87345 allocs/op
// EVMSnailtracer-8                              3         443557845 ns/op        41799274 B/op        700 allocs/op
// TinygoSnailtracer/wazero-8                    7         144487708 ns/op             113 B/op          6 allocs/op
// TinygoSnailtracer/wasmer/singlepass-8         9         123184179 ns/op             560 B/op        34 allocs/op
// TinygoSnailtracer/wasmer/cranelift-8          16          84588197 ns/op             560 B/op        34 allocs/op

func BenchmarkNativeSnailtracer(b *testing.B) {
	s := NewBenchmarkScene(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		color := NewVector(0, 0, 0)
		color = color.Add(s.Trace(512, 384, 8))
		color = color.Add(s.Trace(325, 540, 8))
		color = color.Add(s.Trace(600, 600, 8))
		color = color.Add(s.Trace(522, 524, 8))
		color = color.ScaleDiv(uint256.NewInt(4))

		cr := color.X.Uint64()
		cg := color.Y.Uint64()
		cb := color.Z.Uint64()

		if !validResult(cr, cg, cb) {
			b.Fatal("invalid result:", cr, cg, cb)
		}
	}
}

func BenchmarkParallel4NativeSnailtracer(b *testing.B) {
	tasks := []struct{ x, y, spp int }{
		{512, 384, 8},
		{325, 540, 8},
		{600, 600, 8},
		{522, 524, 8},
	}
	scenes := make([]*Scene, len(tasks))
	for i := 0; i < len(scenes); i++ {
		scenes[i] = NewBenchmarkScene(0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		color := NewVector(0, 0, 0)
		outputChan := make(chan Vector, len(tasks))

		var wg sync.WaitGroup
		wg.Add(len(tasks))

		for i, task := range tasks {
			go func(i, x, y, spp int) {
				defer wg.Done()
				outputChan <- scenes[i].Trace(x, y, spp)
			}(i, task.x, task.y, task.spp)
		}

		wg.Wait()
		close(outputChan)

		for output := range outputChan {
			color = color.Add(output)
		}

		color = color.ScaleDiv(uint256.NewInt(uint64(len(tasks))))

		cr := color.X.Uint64()
		cg := color.Y.Uint64()
		cb := color.Z.Uint64()

		if !validResult(cr, cg, cb) {
			b.Fatal("invalid result:", cr, cg, cb)
		}
	}
}

//go:embed testdata/snailtracer.evm
var evmBytecodeHex []byte

func BenchmarkEVMSnailtracer(b *testing.B) {
	var (
		address        = common.HexToAddress("0xc0ffee")
		origin         = common.HexToAddress("0xc0ffee0001")
		bytecode       = common.Hex2Bytes(string(evmBytecodeHex)[2:])
		initInput      = common.Hex2Bytes("57a86f7d")
		benchmarkInput = common.Hex2Bytes("30627b7c")
		gasLimit       = uint64(1e9)
		txContext      = vm.TxContext{
			Origin:   origin,
			GasPrice: common.Big1,
		}
		context = vm.BlockContext{
			CanTransfer: core.CanTransfer,
			Transfer:    core.Transfer,
			Coinbase:    common.Address{},
			BlockNumber: common.Big1,
			Time:        1,
			Difficulty:  common.Big1,
			GasLimit:    uint64(1e8),
		}
	)

	statedb, err := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
	if err != nil {
		b.Fatal(err)
	}

	statedb.CreateAccount(address)
	statedb.SetCode(address, bytecode)
	statedb.AddAddressToAccessList(address)
	statedb.CreateAccount(origin)
	statedb.SetBalance(origin, big.NewInt(1e18))

	evm := vm.NewEVM(context, txContext, statedb, params.TestChainConfig, vm.Config{})

	var ret []byte

	_, _, err = evm.Call(vm.AccountRef(origin), address, initInput, gasLimit, common.Big0)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ret, _, err = evm.Call(vm.AccountRef(origin), address, benchmarkInput, gasLimit, common.Big0)
		if err != nil {
			b.Fatal(err)
		}

		cr := uint64(ret[0])
		cg := uint64(ret[32])
		cb := uint64(ret[64])

		if !validResult(cr, cg, cb) {
			b.Fatal("invalid result:", cr, cg, cb)
		}
	}
}

//go:embed testdata/snailtracer.wasm
var wasmBytecode []byte

func BenchmarkTinygoSnailtracer(b *testing.B) {
	runtimes := []struct {
		name string
		pc   precompiles.Precompile
	}{
		{"wazero", wasm.NewWazeroPrecompile(wasmBytecode)},
		{"wasmer/singlepass", wasm.NewWasmerPrecompileWithConfig(wasmBytecode, wasmer.NewConfig().UseSinglepassCompiler())},
		{"wasmer/cranelift", wasm.NewWasmerPrecompileWithConfig(wasmBytecode, wasmer.NewConfig().UseCraneliftCompiler())},
	}
	for _, runtime := range runtimes {
		b.Run(runtime.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ret, err := runtime.pc.Run(nil, nil)
				if err != nil {
					b.Fatal(err)
				}
				cr := uint64(ret[0])
				cg := uint64(ret[32])
				cb := uint64(ret[64])
				if !validResult(cr, cg, cb) {
					b.Fatal("invalid result:", cr, cg, cb)
				}
			}
		})
	}
}
