package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	contract "github.com/consensys/gnark-crypto/field/koalabear/solidity/gopkg"
	"github.com/consensys/gnark-crypto/hash"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
)

func setupTestEnv(t testing.TB) (*simulated.Backend, *bind.TransactOpts, *contract.Contract, common.Address, common.Address) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1337))
	if err != nil {
		t.Fatal(err)
	}

	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	balance := new(big.Int)
	balance.SetString("10000000000000000000", 10)

	genesisAlloc := map[common.Address]core.GenesisAccount{
		auth.From: {Balance: balance},
	}

	client := simulated.NewBackend(genesisAlloc)

	contractAddress, _, instance, err := contract.DeployContract(auth, client.Client())
	if err != nil {
		t.Fatal(err)
	}
	client.Commit()

	return client, auth, instance, contractAddress, fromAddress
}

func genRandomHashableInputBytes(size int) []byte {
	data := make([]byte, size)
	for i := 0; i < size/4; i++ {
		var tmp koalabear.Element
		tmp.SetRandom()
		copy(data[4*i:], tmp.Marshal())
	}
	return data
}

func TestCorrectness(t *testing.T) {
	client, _, instance, _, _ := setupTestEnv(t)
	defer client.Close()

	// Load test vectors
	fileTestData, err := os.ReadFile("test_vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var tdata TestData
	json.Unmarshal(fileTestData, &tdata)

	// Test against reference implementation
	h := poseidon2.NewPermutation(16, 6, 21)
	initialState := make([]byte, 32)
	md := hash.NewMerkleDamgardHasher(h, initialState)

	for i := 0; i < 10; i++ {
		data := genRandomHashableInputBytes(32)
		md.Reset()
		md.Write(data[:])
		expected := md.Sum(nil)

		got, err := instance.Hash(nil, data)
		if err != nil {
			t.Fatalf("Hash call failed: %v", err)
		}

		if string(expected) != string(got[:]) {
			t.Fatalf("Hash mismatch at test %d:\nexpected: %x\ngot:      %x", i, expected, got)
		}
	}

	t.Log("All correctness tests passed!")
}

func TestGasUsage(t *testing.T) {
	client, _, instance, contractAddress, _ := setupTestEnv(t)
	defer client.Close()

	sizes := []int{32, 64, 96, 128, 256}
	results := make(map[int]uint64)

	for _, size := range sizes {
		data := genRandomHashableInputBytes(size)

		// Verify correctness first
		h := poseidon2.NewPermutation(16, 6, 21)
		initialState := make([]byte, 32)
		md := hash.NewMerkleDamgardHasher(h, initialState)
		md.Write(data)
		expected := md.Sum(nil)

		got, err := instance.Hash(nil, data)
		if err != nil {
			t.Fatalf("Hash call failed: %v", err)
		}

		if string(expected) != string(got[:]) {
			t.Fatalf("Hash mismatch for size %d:\nexpected: %x\ngot:      %x", size, expected, got)
		}

		// Create the transaction data to estimate gas
		parsed, _ := contract.ContractMetaData.GetAbi()
		callData, _ := parsed.Pack("hash", data)

		// Estimate gas using eth_estimateGas
		gasUsed, err := client.Client().EstimateGas(context.Background(), ethereum.CallMsg{
			To:   &contractAddress,
			Data: callData,
		})

		if err != nil {
			t.Fatalf("Failed to estimate gas for size %d: %v", size, err)
		}

		results[size] = gasUsed
	}

	fmt.Println("\n=== Gas Usage Results ===")
	for _, size := range sizes {
		fmt.Printf("Input size %3d bytes: %6d gas\n", size, results[size])
	}
	fmt.Println("========================")
}
