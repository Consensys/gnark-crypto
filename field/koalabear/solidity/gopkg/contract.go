// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// ContractMetaData contains all meta data concerning the Contract contract.
var ContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"DataIsNotMod32\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_msg\",\"type\":\"bytes\"}],\"name\":\"hash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"poseidon2Hash\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"input\",\"type\":\"bytes32\"}],\"name\":\"padBytes32\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"out\",\"type\":\"bytes\"}],\"stateMutability\":\"pure\",\"type\":\"function\"}]",
	Bin: "0x608060405234602057600e6024565b61105761003082393081505061105790f35b602a565b60405190565b600080fdfe60806040526004361015610013575b61021e565b61001e60003561003d565b80636f4db2d9146100385763aa1e84de0361000e576101f2565b61011f565b60e01c90565b60405190565b600080fd5b600080fd5b90565b61005f81610053565b0361006657565b600080fd5b9050359061007882610056565b565b90602082820312610094576100919160000161006b565b90565b610049565b5190565b60209181520190565b60005b8381106100ba575050906000910152565b8060209183015181850152016100a9565b601f801991011690565b6100f46100fd602093610102936100eb81610099565b9384809361009d565b958691016100a6565b6100cb565b0190565b61011c91602082019160008184039101526100d5565b90565b61014661013561013036600461007a565b610228565b61013d610043565b91829182610106565b0390f35b600080fd5b600080fd5b600080fd5b909182601f830112156101935781359167ffffffffffffffff831161018e57602001926001830284011161018957565b610154565b61014f565b61014a565b906020828203126101ca57600082013567ffffffffffffffff81116101c5576101c19201610159565b9091565b61004e565b610049565b6101d890610053565b9052565b91906101f0906000602085019401906101cf565b565b61021a610209610203366004610198565b90610fca565b610211610043565b918291826101dc565b0390f35b600080fd5b606090565b90610231610223565b506040519060408252602082016000805b6008811061028d57508152600060085b6010811061026c5750604092939450602082015201604052565b9060019061ffff87601085600f03021c16602084600f03021b179101610252565b9060019061ffff87601085600f03021c16602084600703021b179101610242565b600090565b906102bd91610bae565b907f747e80c97d5ff3d179440df5000413c541d0a44f0a1c440d4d06e84c5c15b8226102e891610ded565b907f4734c87b4d3a8a8e078f1df51a1cd0b879bffc5d60b1b050583a8f517a64f2cc61031391610ded565b9061031d90610f11565b9061032790610f11565b61033091610bae565b907f310042f41c53094b1c5cba696d6ffbdf0901f29c325e35b30f56ce7d016a13d661035b91610ded565b907f735b0fea6340a19359973615451586193b15995317fa62952d9bdf646488ca5461038691610ded565b9061039090610f11565b9061039a90610f11565b6103a391610bae565b907f2c58464623f97e046792307e19a4384b16dd8606260ba9654202495b7ea189de6103ce91610ded565b907f2830cef50de1216556f3b44502c06c3b632610e36d42c9e970946c65540dcc056103f991610ded565b9061040390610f11565b9061040d90610f11565b61041691610bae565b9063102b26d061042591610efc565b61042e90610f9b565b906104398183610a0b565b918261044491610a7f565b9161044e91610b0f565b906326a27e0261045d91610efc565b61046690610f9b565b906104718183610a0b565b918261047c91610a7f565b9161048691610b0f565b9063631a12ee61049591610efc565b61049e90610f9b565b906104a98183610a0b565b91826104b491610a7f565b916104be91610b0f565b9063764381a06104cd91610efc565b6104d690610f9b565b906104e18183610a0b565b91826104ec91610a7f565b916104f691610b0f565b9063237670af61050591610efc565b61050e90610f9b565b906105198183610a0b565b918261052491610a7f565b9161052e91610b0f565b906306704ee261053d91610efc565b61054690610f9b565b906105518183610a0b565b918261055c91610a7f565b9161056691610b0f565b906332a30b7061057591610efc565b61057e90610f9b565b906105898183610a0b565b918261059491610a7f565b9161059e91610b0f565b906374f1a6d26105ad91610efc565b6105b690610f9b565b906105c18183610a0b565b91826105cc91610a7f565b916105d691610b0f565b90635c2b3ab36105e591610efc565b6105ee90610f9b565b906105f98183610a0b565b918261060491610a7f565b9161060e91610b0f565b90633dc9785e61061d91610efc565b61062690610f9b565b906106318183610a0b565b918261063c91610a7f565b9161064691610b0f565b90631af253b661065591610efc565b61065e90610f9b565b906106698183610a0b565b918261067491610a7f565b9161067e91610b0f565b90631070cad861068d91610efc565b61069690610f9b565b906106a18183610a0b565b91826106ac91610a7f565b916106b691610b0f565b90632d7e0ec36106c591610efc565b6106ce90610f9b565b906106d98183610a0b565b91826106e491610a7f565b916106ee91610b0f565b90633fb3540b6106fd91610efc565b61070690610f9b565b906107118183610a0b565b918261071c91610a7f565b9161072691610b0f565b90635e3b82d361073591610efc565b61073e90610f9b565b906107498183610a0b565b918261075491610a7f565b9161075e91610b0f565b9063520a1a1461076d91610efc565b61077690610f9b565b906107818183610a0b565b918261078c91610a7f565b9161079691610b0f565b90633536b5086107a591610efc565b6107ae90610f9b565b906107b98183610a0b565b91826107c491610a7f565b916107ce91610b0f565b9063502ceafe6107dd91610efc565b6107e690610f9b565b906107f18183610a0b565b91826107fc91610a7f565b9161080691610b0f565b9063362a42a961081591610efc565b61081e90610f9b565b906108298183610a0b565b918261083491610a7f565b9161083e91610b0f565b9063362c455861084d91610efc565b61085690610f9b565b906108618183610a0b565b918261086c91610a7f565b9161087691610b0f565b90633ecbe92861088591610efc565b61088e90610f9b565b906108998183610a0b565b91826108a491610a7f565b916108ae91610b0f565b907f11218dc0167014be7c8eb2e23a6c3b9b7a033707120e62e50e3b6de86f58ab9c6108d991610ded565b907f1f79642b294d9199367cc09f1b12a5d96fc835223df8b572654b046e68c005fa61090491610ded565b9061090e90610f11565b9061091890610f11565b61092191610bae565b907f09b4f5ef60c076d7231e594c7ba2a2692c9185f83a2840fd718175b52eaf547561094c91610ded565b907f2324d022013e8d86701e300c22d1a4b0273731494f0bb2c7703e164c5cb8526d61097791610ded565b9061098190610f11565b9061098b90610f11565b61099491610bae565b907f52a0a9505c2abe9833fd545013ce64d563a948284730b1264e15994247931fcc6109bf91610ded565b907f65b9070333c90dc57e45ae51298fb39a4ed1df724fbb42a63d33c0e65d61e9df6109ea91610ded565b906109f490610f11565b906109fe90610f11565b610a0791610bae565b9091565b90637f0000019163ffffffff808360e01c8360e01c01818560c01c16828560c01c160101818560a01c16828560a01c160101818560801c16828560801c160101818560601c16828560601c160101818560401c16828560401c160101818560201c16828560201c1601019316911601010690565b90637f0000019081600363ffffffff828060028860e01c098103850860e01b83828860c01c16860860c01b178380838960a01c168008860860a01b178380633f800001848a60801c1609860860801b17838084848a60601c1609860860601b1783806004848a60401c1609860860401b178380633f800000848a60201c1609860860201b17951609820390081790565b90637f0000019081607f63ffffffff828060048860e01c098103850860e01b8380637e810001848a60c01c1609860860c01b178380636f200001848a60a01c1609860860a01b178380637effff82848a60801c1609860860801b178380627f0000848a60601c1609860860601b178380630fe00000848a60401c1609860860401b1783806307f00000848a60201c1609860860201b1795160990081790565b610be3610bc9610bc3610be993959495610d4b565b94610d4b565b93610bd48582610cd0565b81838583989497969596610bec565b96610c5e565b90565b939192637f00000180948163ffffffff9481808a60e01c830860e01b81888c60c01c16860860c01b1781888c60a01c16880860a01b1781888c60801c168a0860801b1791878b60601c16900860601b1791858960401c16900860401b1791838760201c16900860201b17931690081790565b939192637f00000180948163ffffffff9481808a60e01c830860e01b81888c60c01c16860860c01b1781888c60a01c16880860a01b1781888c60801c168a0860801b1791878b60601c16900860601b1791858960401c16900860401b1791838760201c16900860201b17931690081790565b9063ffffffff92637f00000180858460601c168460e01c01868660601c168660e01c0101069481818560401c16828660c01c1601828760401c16838860c01c160101069482828660201c16838760a01c1601838360201c16848460a01c1601010694828082169160801c1601918082169160801c1601010690565b610da563ffffffff610d738360e01c828560c01c16838660a01c1690848760801c1692610dba565b92909160c01b9060e01b179060a01b179060801b1792818160601c1691808260401c1690808360201c16921692610dba565b93909260601b179060401b179060201b171790565b637f000001909492919481868201958181860198818a8a0191820193849201868197010106968001010696010695010691565b63ffffffff808360e01c8360e01c0160e01b818560c01c16828560c01c160160c01b17818560a01c16828560a01c160160a01b17818560801c16828560801c160160801b17818560601c16828560601c160160601b17818560401c16828560401c160160401b17818560201c16828560201c160160201b1793169116011790565b637f0000019063ffffffff80838560e01c8460e01c0860e01b84828760c01c16838660c01c160860c01b1784828760a01c16838660a01c160860a01b1784828760801c16838660801c160860801b1784828760601c16838660601c160860601b1784828760401c16838660401c160860401b1784828760201c16838660201c160860201b1794169116081790565b8060e01c8092019160e01b90039060e01b0190565b637f00000163ffffffff818360e01c81818009900960e01b82828560c01c1681818009900960c01b1782828560a01c1681818009900960a01b1782828560801c1681818009900960801b1782828560601c1681818009900960601b1782828560401c1681818009900960401b1782828560201c1681818009900960201b1792168181800990091790565b8060e01c90637f0000018083800983099160e01b90039060e01b0190565b60046040516330b2ac9b60e21b8152fd5b919091610fd56102ae565b92601f811661101c5760051c60005b818110610ff057505050565b909193602061101060019261100888359182906102b3565b919050610e6e565b95019101919091610fe4565b610fb956fea2646970667358221220a9b1b6b3d476b128a353b303ae09aacd4377e7e75ff69aab07f8f04419daafc164736f6c634300081e0033",
}

// ContractABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractMetaData.ABI instead.
var ContractABI = ContractMetaData.ABI

// ContractBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ContractMetaData.Bin instead.
var ContractBin = ContractMetaData.Bin

// DeployContract deploys a new Ethereum contract, binding an instance of Contract to it.
func DeployContract(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Contract, error) {
	parsed, err := ContractMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ContractBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Contract{ContractCaller: ContractCaller{contract: contract}, ContractTransactor: ContractTransactor{contract: contract}, ContractFilterer: ContractFilterer{contract: contract}}, nil
}

// Contract is an auto generated Go binding around an Ethereum contract.
type Contract struct {
	ContractCaller     // Read-only binding to the contract
	ContractTransactor // Write-only binding to the contract
	ContractFilterer   // Log filterer for contract events
}

// ContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractSession struct {
	Contract     *Contract         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractCallerSession struct {
	Contract *ContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// ContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractTransactorSession struct {
	Contract     *ContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractRaw struct {
	Contract *Contract // Generic contract binding to access the raw methods on
}

// ContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractCallerRaw struct {
	Contract *ContractCaller // Generic read-only contract binding to access the raw methods on
}

// ContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractTransactorRaw struct {
	Contract *ContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContract creates a new instance of Contract, bound to a specific deployed contract.
func NewContract(address common.Address, backend bind.ContractBackend) (*Contract, error) {
	contract, err := bindContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Contract{ContractCaller: ContractCaller{contract: contract}, ContractTransactor: ContractTransactor{contract: contract}, ContractFilterer: ContractFilterer{contract: contract}}, nil
}

// NewContractCaller creates a new read-only instance of Contract, bound to a specific deployed contract.
func NewContractCaller(address common.Address, caller bind.ContractCaller) (*ContractCaller, error) {
	contract, err := bindContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractCaller{contract: contract}, nil
}

// NewContractTransactor creates a new write-only instance of Contract, bound to a specific deployed contract.
func NewContractTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractTransactor, error) {
	contract, err := bindContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractTransactor{contract: contract}, nil
}

// NewContractFilterer creates a new log filterer instance of Contract, bound to a specific deployed contract.
func NewContractFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractFilterer, error) {
	contract, err := bindContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractFilterer{contract: contract}, nil
}

// bindContract binds a generic wrapper to an already deployed contract.
func bindContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.ContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transact(opts, method, params...)
}

// Hash is a free data retrieval call binding the contract method 0xaa1e84de.
//
// Solidity: function hash(bytes _msg) pure returns(bytes32 poseidon2Hash)
func (_Contract *ContractCaller) Hash(opts *bind.CallOpts, _msg []byte) ([32]byte, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "hash", _msg)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// Hash is a free data retrieval call binding the contract method 0xaa1e84de.
//
// Solidity: function hash(bytes _msg) pure returns(bytes32 poseidon2Hash)
func (_Contract *ContractSession) Hash(_msg []byte) ([32]byte, error) {
	return _Contract.Contract.Hash(&_Contract.CallOpts, _msg)
}

// Hash is a free data retrieval call binding the contract method 0xaa1e84de.
//
// Solidity: function hash(bytes _msg) pure returns(bytes32 poseidon2Hash)
func (_Contract *ContractCallerSession) Hash(_msg []byte) ([32]byte, error) {
	return _Contract.Contract.Hash(&_Contract.CallOpts, _msg)
}

// PadBytes32 is a free data retrieval call binding the contract method 0x6f4db2d9.
//
// Solidity: function padBytes32(bytes32 input) pure returns(bytes out)
func (_Contract *ContractCaller) PadBytes32(opts *bind.CallOpts, input [32]byte) ([]byte, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "padBytes32", input)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// PadBytes32 is a free data retrieval call binding the contract method 0x6f4db2d9.
//
// Solidity: function padBytes32(bytes32 input) pure returns(bytes out)
func (_Contract *ContractSession) PadBytes32(input [32]byte) ([]byte, error) {
	return _Contract.Contract.PadBytes32(&_Contract.CallOpts, input)
}

// PadBytes32 is a free data retrieval call binding the contract method 0x6f4db2d9.
//
// Solidity: function padBytes32(bytes32 input) pure returns(bytes out)
func (_Contract *ContractCallerSession) PadBytes32(input [32]byte) ([]byte, error) {
	return _Contract.Contract.PadBytes32(&_Contract.CallOpts, input)
}
