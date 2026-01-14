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
	Bin: "0x610fc361004d600b8282823980515f1a6073146041577f4e487b71000000000000000000000000000000000000000000000000000000005f525f60045260245ffd5b305f52607381538281f3fe730000000000000000000000000000000000000000301460806040526004361061003f575f3560e01c80636f4db2d914610043578063aa1e84de14610073575b5f5ffd5b61005d60048036038101906100589190610dfe565b6100a3565b60405161006a9190610e99565b60405180910390f35b61008d60048036038101906100889190610f1a565b61013c565b60405161009a9190610f74565b60405180910390f35b6060604051905060408152602081015f60f060e05f5b60088110156100e85761ffff87841c1680831b85179450601084039350602083039250506001810190506100b9565b508284525050505f607060e05f5b60088110156101255761ffff87841c1680831b85179450601084039350602083039250506001810190506100f6565b508260208501525050506040810160405250919050565b5f601f82161561016e577fc2cab26c000000000000000000000000000000000000000000000000000000005f5260045ffd5b6040516102008101604052805b6101008201811015610195575f815260208101905061017b565b50838381015b8082101561023c578135610100840160e05f5b60088110156101d95763ffffffff84831c1683526020830192506020820391506001810190506101ae565b5050506101e584610805565b83610100850160e05f5b600881101561022c5763ffffffff85831c16637f0000018185510880865260208601955060208501945060208403935050506001810190506101ef565b505050505060208201915061019b565b6102458361024c565b9350610dba565b5f602082015160c01b5f83015160e01b179050604082015160a01b81179050606082015160801b81179050608082015160601b8117905060a082015160401b8117905060c082015160201b8117905060e082015181179050919050565b5f5f5f5f637f000001868608637f000001898908637f000001818308637f000001898208637f0000018c8308637f00000181637f0000018c8d08089550637f00000182637f0000018e8f08089750637f0000018286089850637f00000181850896505050505050945094509450949050565b5f5f5f5f845f81015160208201516040830151606084015161033f818385876102a9565b835f8a01528260208a01528160408a01528060608a0152838d019c50828c019b50818b019a50808a019950505050505050505050608085015f810151602082015160408301516060840151610396818385876102a9565b835f8a01528260208a01528160408a01528060608a0152838d019c50828c019b50818b019a50808a01995050505050505050505061010085015f8101516020820151604083015160608401516103ee818385876102a9565b835f8a01528260208a01528160408a01528060608a0152838d019c50828c019b50818b019a50808a01995050505050505050505061018085015f810151602082015160408301516060840151610446818385876102a9565b835f8a01528260208a01528160408a01528060608a0152838d019c50828c019b50818b019a50808a019950505050505050505050637f00000184069350637f00000183069250637f00000182069150637f00000181069050845b61020086018110156104f957637f000001855f830151085f820152637f000001846020830151086020820152637f000001836040830151086040820152637f0000018260608301510860608201526080810190506104a0565b505050505050565b5f5f8201518101905060208201518101905060408201518101905060608201518101905060808201518101905060a08201518101905060c08201518101905060e08201518101905061010082015181019050610120820151810190506101408201518101905061016082015181019050610180820151810190506101a0820151810190506101c0820151810190506101e082015181019050637f000001810690505f820151637f0000018060028309637f0000010383085f84015260208301519050637f000001818308602084015260408301519050637f00000180600283098308604084015260608301519050637f00000180633f80000183098308606084015260808301519050637f00000180600383098308608084015260a08301519050637f0000018060048309830860a084015260c08301519050637f00000180633f8000008309830860c084015260e08301519050637f0000018060038309637f00000103830860e08401526101008301519050637f0000018060048309637f0000010383086101008401526101208301519050637f00000180637e810001830983086101208401526101408301519050637f00000180636f200001830983086101408401526101608301519050637f00000180637effff82830983086101608401526101808301519050637f00000180627f0000830983086101808401526101a08301519050637f00000180630fe00000830983086101a08401526101c08301519050637f000001806307f00000830983086101c08401526101e08301519050637f00000180607f830983086101e0840152505050565b8060e05f5b60088110156107a75763ffffffff85831c16637f00000181855108637f000001818209637f000001828209915081865260208601955060208503945050505060018101905061075d565b505050610100810160e05f5b60088110156107fd5763ffffffff86831c16637f00000181855108637f000001818209637f00000182820991508186526020860195506020850394505050506001810190506107b3565b505050505050565b61080e8161031b565b6108597f4734c87b4d3a8a8e078f1df51a1cd0b879bffc5d60b1b050583a8f517a64f2cc7f747e80c97d5ff3d179440df5000413c541d0a44f0a1c440d4d06e84c5c15b82283610758565b6108628161031b565b6108ad7f735b0fea6340a19359973615451586193b15995317fa62952d9bdf646488ca547f310042f41c53094b1c5cba696d6ffbdf0901f29c325e35b30f56ce7d016a13d683610758565b6108b68161031b565b6109017f2830cef50de1216556f3b44502c06c3b632610e36d42c9e970946c65540dcc057f2c58464623f97e046792307e19a4384b16dd8606260ba9654202495b7ea189de83610758565b61090a8161031b565b637f00000163102b26d0825108637f000001818209637f000001828209915081835261093583610501565b5050637f0000016326a27e02825108637f000001818209637f000001828209915081835261096283610501565b5050637f00000163631a12ee825108637f000001818209637f000001828209915081835261098f83610501565b5050637f00000163764381a0825108637f000001818209637f00000182820991508183526109bc83610501565b5050637f00000163237670af825108637f000001818209637f00000182820991508183526109e983610501565b5050637f0000016306704ee2825108637f000001818209637f0000018282099150818352610a1683610501565b5050637f0000016332a30b70825108637f000001818209637f0000018282099150818352610a4383610501565b5050637f0000016374f1a6d2825108637f000001818209637f0000018282099150818352610a7083610501565b5050637f000001635c2b3ab3825108637f000001818209637f0000018282099150818352610a9d83610501565b5050637f000001633dc9785e825108637f000001818209637f0000018282099150818352610aca83610501565b5050637f000001631af253b6825108637f000001818209637f0000018282099150818352610af783610501565b5050637f000001631070cad8825108637f000001818209637f0000018282099150818352610b2483610501565b5050637f000001632d7e0ec3825108637f000001818209637f0000018282099150818352610b5183610501565b5050637f000001633fb3540b825108637f000001818209637f0000018282099150818352610b7e83610501565b5050637f000001635e3b82d3825108637f000001818209637f0000018282099150818352610bab83610501565b5050637f00000163520a1a14825108637f000001818209637f0000018282099150818352610bd883610501565b5050637f000001633536b508825108637f000001818209637f0000018282099150818352610c0583610501565b5050637f00000163502ceafe825108637f000001818209637f0000018282099150818352610c3283610501565b5050637f00000163362a42a9825108637f000001818209637f0000018282099150818352610c5f83610501565b5050637f00000163362c4558825108637f000001818209637f0000018282099150818352610c8c83610501565b5050637f000001633ecbe928825108637f000001818209637f0000018282099150818352610cb983610501565b5050610d067f1f79642b294d9199367cc09f1b12a5d96fc835223df8b572654b046e68c005fa7f11218dc0167014be7c8eb2e23a6c3b9b7a033707120e62e50e3b6de86f58ab9c83610758565b610d0f8161031b565b610d5a7f2324d022013e8d86701e300c22d1a4b0273731494f0bb2c7703e164c5cb8526d7f09b4f5ef60c076d7231e594c7ba2a2692c9185f83a2840fd718175b52eaf547583610758565b610d638161031b565b610dae7f65b9070333c90dc57e45ae51298fb39a4ed1df724fbb42a63d33c0e65d61e9df7f52a0a9505c2abe9833fd545013ce64d563a948284730b1264e15994247931fcc83610758565b610db78161031b565b50565b50505092915050565b5f5ffd5b5f5ffd5b5f819050919050565b610ddd81610dcb565b8114610de7575f5ffd5b50565b5f81359050610df881610dd4565b92915050565b5f60208284031215610e1357610e12610dc3565b5b5f610e2084828501610dea565b91505092915050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f601f19601f8301169050919050565b5f610e6b82610e29565b610e758185610e33565b9350610e85818560208601610e43565b610e8e81610e51565b840191505092915050565b5f6020820190508181035f830152610eb18184610e61565b905092915050565b5f5ffd5b5f5ffd5b5f5ffd5b5f5f83601f840112610eda57610ed9610eb9565b5b8235905067ffffffffffffffff811115610ef757610ef6610ebd565b5b602083019150836001820283011115610f1357610f12610ec1565b5b9250929050565b5f5f60208385031215610f3057610f2f610dc3565b5b5f83013567ffffffffffffffff811115610f4d57610f4c610dc7565b5b610f5985828601610ec5565b92509250509250929050565b610f6e81610dcb565b82525050565b5f602082019050610f875f830184610f65565b9291505056fea264697066735822122084f5d64193f800a8551363c1f91b15881e3b549ab98b936cc10fd1d28c6e213e64736f6c634300081f0033",
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
