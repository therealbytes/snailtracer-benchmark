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
	s := newScene(1024, 768)

	s.deltaX = &Vector{s.width * 513500 / s.height, 0, 0}
	s.deltaY = s.deltaX.cross(s.camera.direction).norm().scaleMul(513500).scaleDiv(1000000)

	s.spheres = []*Sphere{
		{100000000000, &Vector{100001000000, 40800000, 81600000}, &Vector{0, 0, 0}, &Vector{750000, 250000, 250000}, DiffuseMaterial},
		{100000000000, &Vector{-99901000000, 40800000, 81600000}, &Vector{0, 0, 0}, &Vector{250000, 250000, 750000}, DiffuseMaterial},
		{100000000000, &Vector{50000000, 40800000, 100000000000}, &Vector{0, 0, 0}, &Vector{750000, 750000, 750000}, DiffuseMaterial},
		{100000000000, &Vector{50000000, 40800000, -99830000000}, &Vector{0, 0, 0}, &Vector{0, 0, 0}, DiffuseMaterial},
		{100000000000, &Vector{50000000, 100000000000, 81600000}, &Vector{0, 0, 0}, &Vector{750000, 750000, 750000}, DiffuseMaterial},
		{100000000000, &Vector{50000000, -99918400000, 81600000}, &Vector{0, 0, 0}, &Vector{750000, 750000, 750000}, DiffuseMaterial},
		{16500000, &Vector{27000000, 16500000, 47000000}, &Vector{0, 0, 0}, &Vector{999000, 999000, 999000}, SpecularMaterial},
		{600000000, &Vector{50000000, 681330000, 81600000}, &Vector{12000000, 12000000, 12000000}, &Vector{0, 0, 0}, DiffuseMaterial},
	}

	s.triangles = []*Triangle{
		{&Vector{56500000, 25740000, 78000000}, &Vector{73000000, 25740000, 94500000}, &Vector{73000000, 49500000, 78000000}, &Vector{0, 0, 0}, &Vector{0, 0, 0}, &Vector{999000, 999000, 999000}, SpecularMaterial},
		{&Vector{56500000, 23760000, 78000000}, &Vector{73000000, 0, 78000000}, &Vector{73000000, 23760000, 94500000}, &Vector{0, 0, 0}, &Vector{0, 0, 0}, &Vector{999000, 999000, 999000}, SpecularMaterial},
		{&Vector{89500000, 25740000, 78000000}, &Vector{73000000, 49500000, 78000000}, &Vector{73000000, 25740000, 94500000}, &Vector{0, 0, 0}, &Vector{0, 0, 0}, &Vector{999000, 999000, 999000}, SpecularMaterial},
		{&Vector{89500000, 23760000, 78000000}, &Vector{73000000, 23760000, 94500000}, &Vector{73000000, 0, 78000000}, &Vector{0, 0, 0}, &Vector{0, 0, 0}, &Vector{999000, 999000, 999000}, SpecularMaterial},
		{&Vector{56500000, 25740000, 78000000}, &Vector{73000000, 49500000, 78000000}, &Vector{73000000, 25740000, 61500000}, &Vector{0, 0, 0}, &Vector{0, 0, 0}, &Vector{999000, 999000, 999000}, SpecularMaterial},
		{&Vector{56500000, 23760000, 78000000}, &Vector{73000000, 23760000, 61500000}, &Vector{73000000, 0, 78000000}, &Vector{0, 0, 0}, &Vector{0, 0, 0}, &Vector{999000, 999000, 999000}, SpecularMaterial},
		{&Vector{89500000, 25740000, 78000000}, &Vector{73000000, 25740000, 61500000}, &Vector{73000000, 49500000, 78000000}, &Vector{0, 0, 0}, &Vector{0, 0, 0}, &Vector{999000, 999000, 999000}, SpecularMaterial},
		{&Vector{89500000, 23760000, 78000000}, &Vector{73000000, 0, 78000000}, &Vector{73000000, 23760000, 61500000}, &Vector{0, 0, 0}, &Vector{0, 0, 0}, &Vector{999000, 999000, 999000}, SpecularMaterial},
		{&Vector{56500000, 25740000, 78000000}, &Vector{73000000, 25740000, 61500000}, &Vector{89500000, 25740000, 78000000}, &Vector{0, 0, 0}, &Vector{0, 0, 0}, &Vector{999000, 999000, 999000}, SpecularMaterial},
		{&Vector{56500000, 25740000, 78000000}, &Vector{89500000, 25740000, 78000000}, &Vector{73000000, 25740000, 94500000}, &Vector{0, 0, 0}, &Vector{0, 0, 0}, &Vector{999000, 999000, 999000}, SpecularMaterial},
		{&Vector{56500000, 23760000, 78000000}, &Vector{89500000, 23760000, 78000000}, &Vector{73000000, 23760000, 61500000}, &Vector{0, 0, 0}, &Vector{0, 0, 0}, &Vector{999000, 999000, 999000}, SpecularMaterial},
		{&Vector{56500000, 23760000, 78000000}, &Vector{73000000, 23760000, 94500000}, &Vector{89500000, 23760000, 78000000}, &Vector{0, 0, 0}, &Vector{0, 0, 0}, &Vector{999000, 999000, 999000}, SpecularMaterial},
	}

	// Calculate all the triangle surface normals
	for i := range s.triangles {
		tri := s.triangles[i]
		tri.normal = tri.b.sub(tri.a).cross(tri.c.sub(tri.a)).norm()
	}

	// Trace a few pixels and collect their colors (sanity check)
	color := &Vector{0, 0, 0}

	// color = color.add(s.trace(512, 384, 1)) // Flat diffuse surface, opposite wall
	color = color.add(s.trace(512, 384, 8)) // Flat diffuse surface, opposite wall
	color = color.add(s.trace(325, 540, 8)) // Reflective surface mirroring left wall
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
