package snailtracer

import (
	_ "embed"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

func TestNativeSnailtracer(b *testing.T) {
	s := NewBenchmarkScene()

	// Trace a few pixels and collect their colors (sanity check)
	color := &Vector{0, 0, 0}

	color = color.add(s.trace(512, 384, 1)) // Flat diffuse surface, opposite wall
	// color = color.add(s.trace(512, 384, 8)) // Flat diffuse surface, opposite wall
	// color = color.add(s.trace(325, 540, 8)) // Reflective surface mirroring left wall
	// color = color.add(s.trace(600, 600, 8)) // Refractive surface reflecting right wall
	// color = color.add(s.trace(522, 524, 8)) // Reflective surface mirroring the refractive surface reflecting the light
	color = color.scaleDiv(4)

	b.Log(color)
}

//go:embed testdata/bytecode.txt
var bytecodeHex []byte

func TestEVMSnailtracer(b *testing.T) {
	var (
		address        = common.HexToAddress("0xc0ffee")
		origin         = common.HexToAddress("0xc0ffee0001")
		bytecode       = common.Hex2Bytes(string(bytecodeHex)[2:])
		initInput      = common.Hex2Bytes("57a86f7d")
		benchmarkInput = common.Hex2Bytes("30627b7c")
		gasLimit       = uint64(1e8)
		txContext      = vm.TxContext{
			Origin:   origin,
			GasPrice: big.NewInt(1),
		}
		context = vm.BlockContext{
			CanTransfer: core.CanTransfer,
			Transfer:    core.Transfer,
			Coinbase:    common.Address{},
			BlockNumber: big.NewInt(1),
			Time:        1,
			Difficulty:  big.NewInt(1),
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
	var leftoverGas uint64

	ret, leftoverGas, err = evm.Call(vm.AccountRef(origin), address, initInput, gasLimit, common.Big0)
	b.Log(ret, gasLimit-leftoverGas, err)
	if err != nil {
		b.Fatal(err)
	}

	ret, leftoverGas, err = evm.Call(vm.AccountRef(origin), address, benchmarkInput, gasLimit, common.Big0)
	b.Log(ret, gasLimit-leftoverGas, err)
	if err != nil {
		b.Fatal(err)
	}
}
