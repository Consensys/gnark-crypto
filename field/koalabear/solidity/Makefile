SHELL=/bin/bash

.PHONY: clean clean-abi clean-bin clean-go solc

clean-abi:
	cd abi/ && rm -f *.abi

clean-bin:
	cd abi/ && rm -f *.bin

clean-go:
	cd gopkg/ && rm -f *.go

solc:
	solc --bin --abi contracts/poseidon2_lazy.sol --optimize-runs=200 --evm-version paris --via-ir -o abi && abigen --abi abi/Poseidon2.abi --bin abi/Poseidon2.bin --pkg contract --out gopkg/contract.go --type Contract

solc-original:
	solc --bin --abi contracts/poseidon2.sol --optimize-runs=200 --evm-version paris --via-ir -o abi && abigen --abi abi/Poseidon2.abi --bin abi/Poseidon2.bin --pkg contract --out gopkg/contract.go --type Contract

solc-legacy:
	solc --bin --abi contracts/poseidon2.sol --optimize-runs=200 --evm-version paris -o abi && abigen --abi abi/Poseidon2.abi --bin abi/Poseidon2.bin --pkg contract --out gopkg/contract.go --type Contract

clean: clean-abi clean-bin clean-go

all: clean solc
