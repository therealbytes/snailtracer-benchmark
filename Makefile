.PHONY: contract

contract:
	forge build --optimizer-runs 1000
	jq -r '.deployedBytecode.object' out/Snailtracer.sol/SnailTracer.json > snailtracer/testdata/bytecode.txt

image:
	go run ./cmd/render.go
