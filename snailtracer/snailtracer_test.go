package snailtracer

import (
	_ "embed"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/wasm"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

func validResult(r, g, b int64) bool {
	return r == 17 && g == 17 && b == 53
}

// BenchmarkNativeSnailtracer-8          32          34383703 ns/op        11873171 B/op     475931 allocs/op
// BenchmarkEVMSnailtracer-8              3         447541079 ns/op        41799312 B/op        701 allocs/op
// BenchmarkTinygoSnailtracer-8           3         349629869 ns/op            4728 B/op          7 allocs/op

func BenchmarkNativeSnailtracer(b *testing.B) {
	s := NewBenchmarkScene(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		color := NewVector(0, 0, 0)
		color = color.Add(s.Trace(512, 384, 8))
		color = color.Add(s.Trace(325, 540, 8))
		color = color.Add(s.Trace(600, 600, 8))
		color = color.Add(s.Trace(522, 524, 8))
		color = color.ScaleDiv(big.NewInt(4))

		cr := color.X.Int64()
		cg := color.Y.Int64()
		cb := color.Z.Int64()

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

		cr := int64(ret[0])
		cg := int64(ret[32])
		cb := int64(ret[64])

		if !validResult(cr, cg, cb) {
			b.Fatal("invalid result:", cr, cg, cb)
		}
	}
}

//go:embed testdata/snailtracer.wasm
var wasmBytecode []byte

func BenchmarkTinygoSnailtracer(b *testing.B) {
	pc := wasm.NewWasmPrecompile(wasmBytecode)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ret, err := pc.Run(nil, []byte{})
		if err != nil {
			b.Fatal(err)
		}
		cr := int64(ret[0])
		cg := int64(ret[32])
		cb := int64(ret[64])
		if !validResult(cr, cg, cb) {
			b.Fatal("invalid result:", cr, cg, cb)
		}
	}
}
