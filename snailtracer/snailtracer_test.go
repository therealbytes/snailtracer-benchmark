//go:build !tinygo

package snailtracer

import (
	_ "embed"
	"math/big"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
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

func BenchmarkNativeSnailtracer(b *testing.B) {
	s := NewBenchmarkScene(0, 0)
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
		scenes[i] = NewBenchmarkScene(0, 0)
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
		benchmarkInput = common.Hex2Bytes("351578bc0000000000000000000000000000000000000000000000000000000000000000")
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

//go:embed testdata/snailtracer_o2.wasm
var wasmBytecode_o2 []byte

//go:embed testdata/snailtracer_oz.wasm
var wasmBytecode_oz []byte

func BenchmarkTinygoSnailtracer(b *testing.B) {
	runtimes := []struct {
		name     string
		instance *wasmer.Instance
	}{
		{"wasmer/singlepass/o2", newWasmerInstance(b, wasmBytecode_o2, wasmer.NewConfig().UseSinglepassCompiler())},
		{"wasmer/singlepass/oz", newWasmerInstance(b, wasmBytecode_oz, wasmer.NewConfig().UseSinglepassCompiler())},
		{"wasmer/cranelift/o2", newWasmerInstance(b, wasmBytecode_o2, wasmer.NewConfig().UseCraneliftCompiler())},
		{"wasmer/cranelift/oz", newWasmerInstance(b, wasmBytecode_oz, wasmer.NewConfig().UseCraneliftCompiler())},
	}
	for _, runtime := range runtimes {
		b.Run(runtime.name, func(b *testing.B) {
			run, err := runtime.instance.Exports.GetFunction("run")
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ret, err := run(0)
				if err != nil {
					b.Fatal(err)
				}
				ret := _ret.(int32)
				cr := uint64(ret >> 16 & 0xff)
				cg := uint64(ret >> 8 & 0xff)
				cb := uint64(ret & 0xff)
				if !validResult(cr, cg, cb) {
					b.Fatal("invalid result:", cr, cg, cb)
				}
			}
		})
	}
}

func newWasmerInstance(b *testing.B, code []byte, config *wasmer.Config) *wasmer.Instance {
	engine := wasmer.NewEngineWithConfig(config)
	store := wasmer.NewStore(engine)
	module, err := wasmer.NewModule(store, code)
	if err != nil {
		b.Fatal(err)
	}
	wasiEnv, err := wasmer.NewWasiStateBuilder("wasi-program").Finalize()
	if err != nil {
		b.Fatal(err)
	}
	importObject, err := wasiEnv.GenerateImportObject(store, module)
	if err != nil {
		b.Fatal(err)
	}
	instance, err := wasmer.NewInstance(module, importObject)
	if err != nil {
		b.Fatal(err)
	}
	return instance
}
