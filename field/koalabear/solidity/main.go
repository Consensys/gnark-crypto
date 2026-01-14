package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	contract "github.com/consensys/gnark-crypto/field/koalabear/solidity/gopkg"
	"github.com/consensys/gnark-crypto/hash"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
)

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func createSimulatedBackend(privateKey *ecdsa.PrivateKey) (*simulated.Backend, *bind.TransactOpts, error) {

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1337))
	if err != nil {
		return nil, nil, err
	}

	balance := new(big.Int)
	balance.SetString("10000000000000000000", 10) // 10 eth in wei

	address := auth.From
	genesisAlloc := map[common.Address]core.GenesisAccount{
		address: {
			Balance: balance,
		},
	}

	// create simulated backend & deploy the contract
	// blockGasLimit := uint64(14712388)
	// client := backends.NewSimulatedBackend(genesisAlloc, blockGasLimit)
	client := simulated.NewBackend(genesisAlloc)

	return client, auth, nil

}

// func getTransactionOpts(privateKey *ecdsa.PrivateKey, auth *bind.TransactOpts, client *backends.SimulatedBackend) (*bind.TransactOpts, error) {
func getTransactionOpts(privateKey *ecdsa.PrivateKey, auth *bind.TransactOpts, client simulated.Client) (*bind.TransactOpts, error) {

	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, err
	}

	gasprice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(1000000) // -> + add the require for the pairing... +20k
	auth.GasPrice = gasprice

	return auth, nil

}

func packUint256Single(a []koalabear.Element) big.Int {
	var ba, tmp big.Int
	for i := 0; i < 8; i++ {
		a[i].BigInt(&tmp)
		tmp.Lsh(&tmp, uint((7-i)*32))
		ba.Add(&ba, &tmp)
	}
	return ba
}

func PrintBytes(b []byte) {
	for i := 0; i < len(b); i++ {
		if b[i] < 16 {
			fmt.Printf("0x0%x,", b[i])
		} else {
			fmt.Printf("0x%x,", b[i])
		}
	}
	fmt.Println("")
}

type TestData struct {
	List []Entry
}

type Entry struct {
	Input  string
	Output string
}

func Decode() {
	fileTestData, err := os.ReadFile("test_vectors.json")
	checkError(err)
	var tdata TestData
	json.Unmarshal(fileTestData, &tdata)
}

func genRandomHashableInput(size int) []byte {
	data := make([]byte, size)
	for i := 0; i < size/4; i++ {
		var tmp koalabear.Element
		tmp.SetRandom()
		copy(data[4*i:], tmp.Marshal())
	}
	return data
}

func generateTestData(nbEntriesPerSize int) TestData {

	var tmpIn, tmpOut big.Int
	var res TestData
	res.List = make([]Entry, 3*nbEntriesPerSize)

	var sizeInput int

	h := poseidon2.NewPermutation(16, 6, 21)
	initialState := make([]byte, 32)
	md := hash.NewMerkleDamgardHasher(h, initialState)

	// 32 bytes input
	offset := 0
	sizeInput = 32
	for i := 0; i < nbEntriesPerSize; i++ {
		data := genRandomHashableInput(sizeInput)
		md.Write(data[:])
		cc := md.Sum(nil)
		md.Reset()

		tmpIn.SetBytes(data)
		tmpOut.SetBytes(cc)
		res.List[offset+i] = Entry{Input: fmt.Sprintf("0x%s", tmpIn.Text(16)), Output: fmt.Sprintf("0x%s", tmpOut.Text(16))}
	}

	// 64 bytes input
	offset += nbEntriesPerSize
	sizeInput = 64
	for i := 0; i < nbEntriesPerSize; i++ {
		data := genRandomHashableInput(sizeInput)
		md.Write(data[:])
		cc := md.Sum(nil)
		md.Reset()

		tmpIn.SetBytes(data)
		tmpOut.SetBytes(cc)
		res.List[offset+i] = Entry{Input: fmt.Sprintf("0x%s", tmpIn.Text(16)), Output: fmt.Sprintf("0x%s", tmpOut.Text(16))}
	}

	// 96 bytes input
	offset += nbEntriesPerSize
	sizeInput = 96
	for i := 0; i < nbEntriesPerSize; i++ {
		data := genRandomHashableInput(sizeInput)
		md.Write(data[:])
		cc := md.Sum(nil)
		md.Reset()

		tmpIn.SetBytes(data)
		tmpOut.SetBytes(cc)
		res.List[offset+i] = Entry{Input: fmt.Sprintf("0x%s", tmpIn.Text(16)), Output: fmt.Sprintf("0x%s", tmpOut.Text(16))}
	}

	return res

}

func main() {

	// generate test vectors
	td := generateTestData(100)
	b, err := json.MarshalIndent(&td, "", "\t")
	checkError(err)
	outFile, err := os.Create("test_vectors.json")
	checkError(err)
	_, err = outFile.Write(b)
	checkError(err)
	outFile.Close()

	// create account
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	// create simulated backend
	client, auth, err := createSimulatedBackend(privateKey)
	checkError(err)

	// deploy the contract
	contractAddress, _, instance, err := contract.DeployContract(auth, client.Client())
	checkError(err)
	client.Commit()

	// Interact with the contract
	auth, err = getTransactionOpts(privateKey, auth, client.Client())
	checkError(err)

	var data [32]byte
	for i := 0; i < 4; i++ {
		var tmp koalabear.Element
		tmp.SetRandom()
		copy(data[4*i:], tmp.Marshal())
	}
	h := poseidon2.NewPermutation(16, 6, 21)
	initialState := make([]byte, 32)
	md := hash.NewMerkleDamgardHasher(h, initialState)
	md.Write(data[:])
	cc := md.Sum(nil)
	PrintBytes(cc)

	// fileTestData, err := os.ReadFile("test_vectors.json")
	// checkError(err)
	// var tdata TestData
	// json.Unmarshal(fileTestData, &tdata)
	// fmt.Printf("size tdata = %d\n", len(tdata.List))

	// h := poseidon2.NewPermutation(16, 6, 21)
	// initialState := make([]byte, 32)
	// md := hash.NewMerkleDamgardHasher(h, initialState)
	// md.Write(tdata.List[0].Input[:])
	// cc := md.Sum(nil)
	// PrintBytes(cc)
	// PrintBytes(tdata.List[0].Output[:])

	// _, err = instance.Hash(nil)
	// res, err := instance.Hash(nil, tdata.List[3].Input[:])
	res, err := instance.Hash(nil, data[:])
	checkError(err)
	client.Commit()
	PrintBytes(res[:])
	// PrintBytes(tdata.List[3].Output[:])

	// var sk [8]koalabear.Element
	// for i := 0; i < 8; i++ {
	// 	sk[i].SetBytes(res[4*i : 4*i+4])
	// }
	// bsk := packUint256Single(sk[:])
	// fmt.Println(bsk.String())

	// query event
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(0),
		ToBlock:   big.NewInt(2),
		Addresses: []common.Address{
			contractAddress,
		},
	}

	logs, err := client.Client().FilterLogs(context.Background(), query)
	checkError(err)

	contractABI, err := abi.JSON(strings.NewReader(string(contract.ContractABI)))
	checkError(err)

	for _, vLog := range logs {

		var event interface{}
		// err = contractABI.UnpackIntoInterface(&event, "PrintUint32", vLog.Data)
		err = contractABI.UnpackIntoInterface(&event, "PrintUint256", vLog.Data)
		checkError(err)
		fmt.Println(event)
	}
}
