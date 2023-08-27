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

func validResult(r, g, b int64) bool {
	return r == 25 && g == 24 && b == 99
}

func TestNativeSnailtracer(t *testing.T) {
	s := NewBenchmarkScene()

	color := NewVector(0, 0, 0)

	color = color.Add(s.trace(512, 384, 8)) // Flat diffuse surface, opposite wall
	color = color.Add(s.trace(325, 540, 8)) // Reflective surface mirroring left wall
	color = color.Add(s.trace(600, 600, 8)) // Refractive surface reflecting right wall
	color = color.Add(s.trace(522, 524, 8)) // Reflective surface mirroring the refractive surface reflecting the light
	color = color.ScaleDiv(big.NewInt(4))

	r := color.x.Int64()
	g := color.y.Int64()
	b := color.z.Int64()

	if !validResult(r, g, b) {
		t.Fatal("invalid result:", r, g, b)
	}
}

//go:embed testdata/bytecode.txt
var bytecodeHex []byte

func TestEVMSnailtracer(t *testing.T) {
	var (
		address        = common.HexToAddress("0xc0ffee")
		origin         = common.HexToAddress("0xc0ffee0001")
		bytecode       = common.Hex2Bytes(string(bytecodeHex)[2:])
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
		t.Fatal(err)
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
		t.Fatal(err)
	}

	ret, _, err = evm.Call(vm.AccountRef(origin), address, benchmarkInput, gasLimit, common.Big0)
	if err != nil {
		t.Fatal(err)
	}

	r := int64(ret[0])
	g := int64(ret[32])
	b := int64(ret[64])

	if !validResult(r, g, b) {
		t.Fatal("invalid result:", r, g, b)
	}
}
