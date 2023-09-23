.PHONY: prepare solidity tinygo render benchmark

prepare:
	mkdir -p snailtracer/testdata
	mkdir -p results

solidity:
	forge build --optimizer-runs 1000 --sizes
	jq -r '.deployedBytecode.object' out/Snailtracer.sol/SnailTracer.json > snailtracer/testdata/snailtracer.evm

# deploy: solidity
# 	@address=$$(forge create snailtracer-sol/Snailtracer.sol:SnailTracer --private-key 0x2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6 --json | jq -r '.deployedTo'); \
# 	cast send $$address "Benchmark()" --private-key 0x2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6 --gas-limit 100000000

tinygo:
	tinygo build -opt=2 -no-debug -o snailtracer/testdata/snailtracer_o2.wasm -target wasi tinygo/main.go
	tinygo build -opt=z -no-debug -o snailtracer/testdata/snailtracer_oz.wasm -target wasi tinygo/main.go

render:
	go run cmd/render.go

benchmark:
	cd snailtracer && go test -bench . -benchmem | tee ../results/benchmark_output.txt
	echo "Benchmark,Iterations,ns/op,Bytes/op,Allocs/op" > results/benchmark_results.csv
	awk '/Benchmark/ { print $$1 "," $$2 "," $$3 "," $$5 "," $$7 }' results/benchmark_output.txt >> results/benchmark_results.csv
	rm results/benchmark_output.txt
