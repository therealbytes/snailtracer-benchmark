.PHONY: solidity deploy tinygo render

solidity:
	forge build --optimizer-runs 1000 --sizes
	jq -r '.deployedBytecode.object' out/Snailtracer.sol/SnailTracer.json > snailtracer/testdata/snailtracer.evm

deploy: solidity
	@address=$$(forge create ./snailtracer-sol/Snailtracer.sol:SnailTracer --private-key 0x2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6 --json | jq -r '.deployedTo'); \
	cast send $$address "Benchmark()" --private-key 0x2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6 --gas-limit 100000000

tinygo:
	tinygo build -opt=2 -no-debug -o ./snailtracer/testdata/snailtracer.wasm -target wasi ./tinygo/main.go

render:
	go run ./cmd/render.go
