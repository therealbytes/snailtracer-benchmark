.PHONY: contract

contract:
	forge build --optimizer-runs 1000
	jq -r '.deployedBytecode.object' out/Snailtracer.sol/SnailTracer.json > snailtracer/testdata/bytecode.txt

deploy: contract
	@address=$$(forge create ./snailtracer-sol/Snailtracer.sol:SnailTracer --private-key 0x2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6 --json | jq -r '.deployedTo'); \
	cast send $$address "Benchmark()" --private-key 0x2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6 --gas-limit 100000000

image:
	go run ./cmd/render.go
