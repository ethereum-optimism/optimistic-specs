// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package l2oo

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
)

// L2OutputOracleMetaData contains all meta data concerning the L2OutputOracle contract.
var L2OutputOracleMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_submissionInterval\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_l2BlockTime\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"_genesisL2Output\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"_historicalTotalBlocks\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"sequencer\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"_l2Output\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"_l2timestamp\",\"type\":\"uint256\"}],\"name\":\"l2OutputAppended\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_l2Output\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"_l2timestamp\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"_blockhash\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"_blocknumber\",\"type\":\"uint256\"}],\"name\":\"appendL2Output\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_l2timestamp\",\"type\":\"uint256\"}],\"name\":\"computeL2BlockNumber\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_l2Timestamp\",\"type\":\"uint256\"}],\"name\":\"getL2Output\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"historicalTotalBlocks\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"l2BlockTime\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"latestBlockTimestamp\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"nextTimestamp\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"startingBlockTimestamp\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"submissionInterval\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x6101006040523480156200001257600080fd5b50604051620013583803806200135883398181016040528101906200003891906200025f565b620000586200004c620000b360201b60201c565b620000bb60201b60201c565b84608081815250508360a081815250508260026000428152602001908152602001600020819055508160c08181525050426001819055504260e08181525050620000a881620000bb60201b60201c565b5050505050620002e7565b600033905090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050816000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508173ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e060405160405180910390a35050565b600080fd5b6000819050919050565b620001998162000184565b8114620001a557600080fd5b50565b600081519050620001b9816200018e565b92915050565b6000819050919050565b620001d481620001bf565b8114620001e057600080fd5b50565b600081519050620001f481620001c9565b92915050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006200022782620001fa565b9050919050565b62000239816200021a565b81146200024557600080fd5b50565b60008151905062000259816200022e565b92915050565b600080600080600060a086880312156200027e576200027d6200017f565b5b60006200028e88828901620001a8565b9550506020620002a188828901620001a8565b9450506040620002b488828901620001e3565b9350506060620002c788828901620001a8565b9250506080620002da8882890162000248565b9150509295509295909350565b60805160a05160c05160e05161101462000344600039600081816102b30152818161033701526106c601526000818161036f01526106ea01526000818161031601526106850152600081816105a1015261070e01526110146000f3fe6080604052600436106100a75760003560e01c806393991af31161006457806393991af31461019d578063a25ae557146101c8578063c5095d6814610205578063c90ec2da14610230578063e1a41bcf1461025b578063f2fde38b14610286576100a7565b806302e51345146100ac5780630c1952d3146100e95780632518810414610114578063357e951f14610130578063715018a61461015b5780638da5cb5b14610172575b600080fd5b3480156100b857600080fd5b506100d360048036038101906100ce919061092f565b6102af565b6040516100e0919061096b565b60405180910390f35b3480156100f557600080fd5b506100fe6103a9565b60405161010b919061096b565b60405180910390f35b61012e600480360381019061012991906109bc565b6103af565b005b34801561013c57600080fd5b5061014561059d565b604051610152919061096b565b60405180910390f35b34801561016757600080fd5b506101706105d2565b005b34801561017e57600080fd5b5061018761065a565b6040516101949190610a64565b60405180910390f35b3480156101a957600080fd5b506101b2610683565b6040516101bf919061096b565b60405180910390f35b3480156101d457600080fd5b506101ef60048036038101906101ea919061092f565b6106a7565b6040516101fc9190610a8e565b60405180910390f35b34801561021157600080fd5b5061021a6106c4565b604051610227919061096b565b60405180910390f35b34801561023c57600080fd5b506102456106e8565b604051610252919061096b565b60405180910390f35b34801561026757600080fd5b5061027061070c565b60405161027d919061096b565b60405180910390f35b34801561029257600080fd5b506102ad60048036038101906102a89190610ad5565b610730565b005b60007f0000000000000000000000000000000000000000000000000000000000000000821015610314576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161030b90610b85565b60405180910390fd5b7f00000000000000000000000000000000000000000000000000000000000000007f0000000000000000000000000000000000000000000000000000000000000000836103619190610bd4565b61036b9190610c37565b60017f00000000000000000000000000000000000000000000000000000000000000006103989190610c68565b6103a29190610c68565b9050919050565b60015481565b6103b7610828565b73ffffffffffffffffffffffffffffffffffffffff166103d561065a565b73ffffffffffffffffffffffffffffffffffffffff161461042b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161042290610d0a565b60405180910390fd5b42831061046d576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161046490610d9c565b60405180910390fd5b61047561059d565b83146104b6576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016104ad90610e2e565b60405180910390fd5b6000801b8414156104fc576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016104f390610e9a565b60405180910390fd5b6000801b821461054a5781814014610549576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161054090610f2c565b60405180910390fd5b5b8360026000858152602001908152602001600020819055508260018190555082847f92701dc658a5d84c16077ea6de344b9995e21a96a05d45e4cd22f37a3d266f8b60405160405180910390a350505050565b60007f00000000000000000000000000000000000000000000000000000000000000006001546105cd9190610c68565b905090565b6105da610828565b73ffffffffffffffffffffffffffffffffffffffff166105f861065a565b73ffffffffffffffffffffffffffffffffffffffff161461064e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161064590610d0a565b60405180910390fd5b6106586000610830565b565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b7f000000000000000000000000000000000000000000000000000000000000000081565b600060026000838152602001908152602001600020549050919050565b7f000000000000000000000000000000000000000000000000000000000000000081565b7f000000000000000000000000000000000000000000000000000000000000000081565b7f000000000000000000000000000000000000000000000000000000000000000081565b610738610828565b73ffffffffffffffffffffffffffffffffffffffff1661075661065a565b73ffffffffffffffffffffffffffffffffffffffff16146107ac576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016107a390610d0a565b60405180910390fd5b600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff16141561081c576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161081390610fbe565b60405180910390fd5b61082581610830565b50565b600033905090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050816000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508173ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e060405160405180910390a35050565b600080fd5b6000819050919050565b61090c816108f9565b811461091757600080fd5b50565b60008135905061092981610903565b92915050565b600060208284031215610945576109446108f4565b5b60006109538482850161091a565b91505092915050565b610965816108f9565b82525050565b6000602082019050610980600083018461095c565b92915050565b6000819050919050565b61099981610986565b81146109a457600080fd5b50565b6000813590506109b681610990565b92915050565b600080600080608085870312156109d6576109d56108f4565b5b60006109e4878288016109a7565b94505060206109f58782880161091a565b9350506040610a06878288016109a7565b9250506060610a178782880161091a565b91505092959194509250565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000610a4e82610a23565b9050919050565b610a5e81610a43565b82525050565b6000602082019050610a796000830184610a55565b92915050565b610a8881610986565b82525050565b6000602082019050610aa36000830184610a7f565b92915050565b610ab281610a43565b8114610abd57600080fd5b50565b600081359050610acf81610aa9565b92915050565b600060208284031215610aeb57610aea6108f4565b5b6000610af984828501610ac0565b91505092915050565b600082825260208201905092915050565b7f54696d657374616d70207072696f7220746f207374617274696e67426c6f636b60008201527f54696d657374616d700000000000000000000000000000000000000000000000602082015250565b6000610b6f602983610b02565b9150610b7a82610b13565b604082019050919050565b60006020820190508181036000830152610b9e81610b62565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000610bdf826108f9565b9150610bea836108f9565b925082821015610bfd57610bfc610ba5565b5b828203905092915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b6000610c42826108f9565b9150610c4d836108f9565b925082610c5d57610c5c610c08565b5b828204905092915050565b6000610c73826108f9565b9150610c7e836108f9565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff03821115610cb357610cb2610ba5565b5b828201905092915050565b7f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572600082015250565b6000610cf4602083610b02565b9150610cff82610cbe565b602082019050919050565b60006020820190508181036000830152610d2381610ce7565b9050919050565b7f43616e6e6f7420617070656e64204c32206f757470757420696e20667574757260008201527f6500000000000000000000000000000000000000000000000000000000000000602082015250565b6000610d86602183610b02565b9150610d9182610d2a565b604082019050919050565b60006020820190508181036000830152610db581610d79565b9050919050565b7f54696d657374616d70206e6f7420657175616c20746f206e657874206578706560008201527f637465642074696d657374616d70000000000000000000000000000000000000602082015250565b6000610e18602e83610b02565b9150610e2382610dbc565b604082019050919050565b60006020820190508181036000830152610e4781610e0b565b9050919050565b7f43616e6e6f74207375626d697420656d707479204c32206f7574707574000000600082015250565b6000610e84601d83610b02565b9150610e8f82610e4e565b602082019050919050565b60006020820190508181036000830152610eb381610e77565b9050919050565b7f426c6f636b6861736820646f6573206e6f74206d61746368207468652068617360008201527f6820617420746865206578706563746564206865696768742e00000000000000602082015250565b6000610f16603983610b02565b9150610f2182610eba565b604082019050919050565b60006020820190508181036000830152610f4581610f09565b9050919050565b7f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160008201527f6464726573730000000000000000000000000000000000000000000000000000602082015250565b6000610fa8602683610b02565b9150610fb382610f4c565b604082019050919050565b60006020820190508181036000830152610fd781610f9b565b905091905056fea26469706673582212206ad05ccab38d6bd37a3f9bb6a352f9a508d85a42e9e17c7159dd42349f6f296964736f6c634300080a0033",
}

// L2OutputOracleABI is the input ABI used to generate the binding from.
// Deprecated: Use L2OutputOracleMetaData.ABI instead.
var L2OutputOracleABI = L2OutputOracleMetaData.ABI

// L2OutputOracleBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use L2OutputOracleMetaData.Bin instead.
var L2OutputOracleBin = L2OutputOracleMetaData.Bin

// DeployL2OutputOracle deploys a new Ethereum contract, binding an instance of L2OutputOracle to it.
func DeployL2OutputOracle(auth *bind.TransactOpts, backend bind.ContractBackend, _submissionInterval *big.Int, _l2BlockTime *big.Int, _genesisL2Output [32]byte, _historicalTotalBlocks *big.Int, sequencer common.Address) (common.Address, *types.Transaction, *L2OutputOracle, error) {
	parsed, err := L2OutputOracleMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(L2OutputOracleBin), backend, _submissionInterval, _l2BlockTime, _genesisL2Output, _historicalTotalBlocks, sequencer)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &L2OutputOracle{L2OutputOracleCaller: L2OutputOracleCaller{contract: contract}, L2OutputOracleTransactor: L2OutputOracleTransactor{contract: contract}, L2OutputOracleFilterer: L2OutputOracleFilterer{contract: contract}}, nil
}

// L2OutputOracle is an auto generated Go binding around an Ethereum contract.
type L2OutputOracle struct {
	L2OutputOracleCaller     // Read-only binding to the contract
	L2OutputOracleTransactor // Write-only binding to the contract
	L2OutputOracleFilterer   // Log filterer for contract events
}

// L2OutputOracleCaller is an auto generated read-only Go binding around an Ethereum contract.
type L2OutputOracleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// L2OutputOracleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type L2OutputOracleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// L2OutputOracleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type L2OutputOracleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// L2OutputOracleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type L2OutputOracleSession struct {
	Contract     *L2OutputOracle   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// L2OutputOracleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type L2OutputOracleCallerSession struct {
	Contract *L2OutputOracleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// L2OutputOracleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type L2OutputOracleTransactorSession struct {
	Contract     *L2OutputOracleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// L2OutputOracleRaw is an auto generated low-level Go binding around an Ethereum contract.
type L2OutputOracleRaw struct {
	Contract *L2OutputOracle // Generic contract binding to access the raw methods on
}

// L2OutputOracleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type L2OutputOracleCallerRaw struct {
	Contract *L2OutputOracleCaller // Generic read-only contract binding to access the raw methods on
}

// L2OutputOracleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type L2OutputOracleTransactorRaw struct {
	Contract *L2OutputOracleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewL2OutputOracle creates a new instance of L2OutputOracle, bound to a specific deployed contract.
func NewL2OutputOracle(address common.Address, backend bind.ContractBackend) (*L2OutputOracle, error) {
	contract, err := bindL2OutputOracle(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &L2OutputOracle{L2OutputOracleCaller: L2OutputOracleCaller{contract: contract}, L2OutputOracleTransactor: L2OutputOracleTransactor{contract: contract}, L2OutputOracleFilterer: L2OutputOracleFilterer{contract: contract}}, nil
}

// NewL2OutputOracleCaller creates a new read-only instance of L2OutputOracle, bound to a specific deployed contract.
func NewL2OutputOracleCaller(address common.Address, caller bind.ContractCaller) (*L2OutputOracleCaller, error) {
	contract, err := bindL2OutputOracle(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &L2OutputOracleCaller{contract: contract}, nil
}

// NewL2OutputOracleTransactor creates a new write-only instance of L2OutputOracle, bound to a specific deployed contract.
func NewL2OutputOracleTransactor(address common.Address, transactor bind.ContractTransactor) (*L2OutputOracleTransactor, error) {
	contract, err := bindL2OutputOracle(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &L2OutputOracleTransactor{contract: contract}, nil
}

// NewL2OutputOracleFilterer creates a new log filterer instance of L2OutputOracle, bound to a specific deployed contract.
func NewL2OutputOracleFilterer(address common.Address, filterer bind.ContractFilterer) (*L2OutputOracleFilterer, error) {
	contract, err := bindL2OutputOracle(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &L2OutputOracleFilterer{contract: contract}, nil
}

// bindL2OutputOracle binds a generic wrapper to an already deployed contract.
func bindL2OutputOracle(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(L2OutputOracleABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_L2OutputOracle *L2OutputOracleRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _L2OutputOracle.Contract.L2OutputOracleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_L2OutputOracle *L2OutputOracleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _L2OutputOracle.Contract.L2OutputOracleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_L2OutputOracle *L2OutputOracleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _L2OutputOracle.Contract.L2OutputOracleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_L2OutputOracle *L2OutputOracleCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _L2OutputOracle.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_L2OutputOracle *L2OutputOracleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _L2OutputOracle.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_L2OutputOracle *L2OutputOracleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _L2OutputOracle.Contract.contract.Transact(opts, method, params...)
}

// ComputeL2BlockNumber is a free data retrieval call binding the contract method 0x02e51345.
//
// Solidity: function computeL2BlockNumber(uint256 _l2timestamp) view returns(uint256)
func (_L2OutputOracle *L2OutputOracleCaller) ComputeL2BlockNumber(opts *bind.CallOpts, _l2timestamp *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _L2OutputOracle.contract.Call(opts, &out, "computeL2BlockNumber", _l2timestamp)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ComputeL2BlockNumber is a free data retrieval call binding the contract method 0x02e51345.
//
// Solidity: function computeL2BlockNumber(uint256 _l2timestamp) view returns(uint256)
func (_L2OutputOracle *L2OutputOracleSession) ComputeL2BlockNumber(_l2timestamp *big.Int) (*big.Int, error) {
	return _L2OutputOracle.Contract.ComputeL2BlockNumber(&_L2OutputOracle.CallOpts, _l2timestamp)
}

// ComputeL2BlockNumber is a free data retrieval call binding the contract method 0x02e51345.
//
// Solidity: function computeL2BlockNumber(uint256 _l2timestamp) view returns(uint256)
func (_L2OutputOracle *L2OutputOracleCallerSession) ComputeL2BlockNumber(_l2timestamp *big.Int) (*big.Int, error) {
	return _L2OutputOracle.Contract.ComputeL2BlockNumber(&_L2OutputOracle.CallOpts, _l2timestamp)
}

// GetL2Output is a free data retrieval call binding the contract method 0xa25ae557.
//
// Solidity: function getL2Output(uint256 _l2Timestamp) view returns(bytes32)
func (_L2OutputOracle *L2OutputOracleCaller) GetL2Output(opts *bind.CallOpts, _l2Timestamp *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _L2OutputOracle.contract.Call(opts, &out, "getL2Output", _l2Timestamp)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetL2Output is a free data retrieval call binding the contract method 0xa25ae557.
//
// Solidity: function getL2Output(uint256 _l2Timestamp) view returns(bytes32)
func (_L2OutputOracle *L2OutputOracleSession) GetL2Output(_l2Timestamp *big.Int) ([32]byte, error) {
	return _L2OutputOracle.Contract.GetL2Output(&_L2OutputOracle.CallOpts, _l2Timestamp)
}

// GetL2Output is a free data retrieval call binding the contract method 0xa25ae557.
//
// Solidity: function getL2Output(uint256 _l2Timestamp) view returns(bytes32)
func (_L2OutputOracle *L2OutputOracleCallerSession) GetL2Output(_l2Timestamp *big.Int) ([32]byte, error) {
	return _L2OutputOracle.Contract.GetL2Output(&_L2OutputOracle.CallOpts, _l2Timestamp)
}

// HistoricalTotalBlocks is a free data retrieval call binding the contract method 0xc90ec2da.
//
// Solidity: function historicalTotalBlocks() view returns(uint256)
func (_L2OutputOracle *L2OutputOracleCaller) HistoricalTotalBlocks(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _L2OutputOracle.contract.Call(opts, &out, "historicalTotalBlocks")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// HistoricalTotalBlocks is a free data retrieval call binding the contract method 0xc90ec2da.
//
// Solidity: function historicalTotalBlocks() view returns(uint256)
func (_L2OutputOracle *L2OutputOracleSession) HistoricalTotalBlocks() (*big.Int, error) {
	return _L2OutputOracle.Contract.HistoricalTotalBlocks(&_L2OutputOracle.CallOpts)
}

// HistoricalTotalBlocks is a free data retrieval call binding the contract method 0xc90ec2da.
//
// Solidity: function historicalTotalBlocks() view returns(uint256)
func (_L2OutputOracle *L2OutputOracleCallerSession) HistoricalTotalBlocks() (*big.Int, error) {
	return _L2OutputOracle.Contract.HistoricalTotalBlocks(&_L2OutputOracle.CallOpts)
}

// L2BlockTime is a free data retrieval call binding the contract method 0x93991af3.
//
// Solidity: function l2BlockTime() view returns(uint256)
func (_L2OutputOracle *L2OutputOracleCaller) L2BlockTime(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _L2OutputOracle.contract.Call(opts, &out, "l2BlockTime")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// L2BlockTime is a free data retrieval call binding the contract method 0x93991af3.
//
// Solidity: function l2BlockTime() view returns(uint256)
func (_L2OutputOracle *L2OutputOracleSession) L2BlockTime() (*big.Int, error) {
	return _L2OutputOracle.Contract.L2BlockTime(&_L2OutputOracle.CallOpts)
}

// L2BlockTime is a free data retrieval call binding the contract method 0x93991af3.
//
// Solidity: function l2BlockTime() view returns(uint256)
func (_L2OutputOracle *L2OutputOracleCallerSession) L2BlockTime() (*big.Int, error) {
	return _L2OutputOracle.Contract.L2BlockTime(&_L2OutputOracle.CallOpts)
}

// LatestBlockTimestamp is a free data retrieval call binding the contract method 0x0c1952d3.
//
// Solidity: function latestBlockTimestamp() view returns(uint256)
func (_L2OutputOracle *L2OutputOracleCaller) LatestBlockTimestamp(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _L2OutputOracle.contract.Call(opts, &out, "latestBlockTimestamp")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// LatestBlockTimestamp is a free data retrieval call binding the contract method 0x0c1952d3.
//
// Solidity: function latestBlockTimestamp() view returns(uint256)
func (_L2OutputOracle *L2OutputOracleSession) LatestBlockTimestamp() (*big.Int, error) {
	return _L2OutputOracle.Contract.LatestBlockTimestamp(&_L2OutputOracle.CallOpts)
}

// LatestBlockTimestamp is a free data retrieval call binding the contract method 0x0c1952d3.
//
// Solidity: function latestBlockTimestamp() view returns(uint256)
func (_L2OutputOracle *L2OutputOracleCallerSession) LatestBlockTimestamp() (*big.Int, error) {
	return _L2OutputOracle.Contract.LatestBlockTimestamp(&_L2OutputOracle.CallOpts)
}

// NextTimestamp is a free data retrieval call binding the contract method 0x357e951f.
//
// Solidity: function nextTimestamp() view returns(uint256)
func (_L2OutputOracle *L2OutputOracleCaller) NextTimestamp(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _L2OutputOracle.contract.Call(opts, &out, "nextTimestamp")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NextTimestamp is a free data retrieval call binding the contract method 0x357e951f.
//
// Solidity: function nextTimestamp() view returns(uint256)
func (_L2OutputOracle *L2OutputOracleSession) NextTimestamp() (*big.Int, error) {
	return _L2OutputOracle.Contract.NextTimestamp(&_L2OutputOracle.CallOpts)
}

// NextTimestamp is a free data retrieval call binding the contract method 0x357e951f.
//
// Solidity: function nextTimestamp() view returns(uint256)
func (_L2OutputOracle *L2OutputOracleCallerSession) NextTimestamp() (*big.Int, error) {
	return _L2OutputOracle.Contract.NextTimestamp(&_L2OutputOracle.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_L2OutputOracle *L2OutputOracleCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _L2OutputOracle.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_L2OutputOracle *L2OutputOracleSession) Owner() (common.Address, error) {
	return _L2OutputOracle.Contract.Owner(&_L2OutputOracle.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_L2OutputOracle *L2OutputOracleCallerSession) Owner() (common.Address, error) {
	return _L2OutputOracle.Contract.Owner(&_L2OutputOracle.CallOpts)
}

// StartingBlockTimestamp is a free data retrieval call binding the contract method 0xc5095d68.
//
// Solidity: function startingBlockTimestamp() view returns(uint256)
func (_L2OutputOracle *L2OutputOracleCaller) StartingBlockTimestamp(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _L2OutputOracle.contract.Call(opts, &out, "startingBlockTimestamp")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StartingBlockTimestamp is a free data retrieval call binding the contract method 0xc5095d68.
//
// Solidity: function startingBlockTimestamp() view returns(uint256)
func (_L2OutputOracle *L2OutputOracleSession) StartingBlockTimestamp() (*big.Int, error) {
	return _L2OutputOracle.Contract.StartingBlockTimestamp(&_L2OutputOracle.CallOpts)
}

// StartingBlockTimestamp is a free data retrieval call binding the contract method 0xc5095d68.
//
// Solidity: function startingBlockTimestamp() view returns(uint256)
func (_L2OutputOracle *L2OutputOracleCallerSession) StartingBlockTimestamp() (*big.Int, error) {
	return _L2OutputOracle.Contract.StartingBlockTimestamp(&_L2OutputOracle.CallOpts)
}

// SubmissionInterval is a free data retrieval call binding the contract method 0xe1a41bcf.
//
// Solidity: function submissionInterval() view returns(uint256)
func (_L2OutputOracle *L2OutputOracleCaller) SubmissionInterval(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _L2OutputOracle.contract.Call(opts, &out, "submissionInterval")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SubmissionInterval is a free data retrieval call binding the contract method 0xe1a41bcf.
//
// Solidity: function submissionInterval() view returns(uint256)
func (_L2OutputOracle *L2OutputOracleSession) SubmissionInterval() (*big.Int, error) {
	return _L2OutputOracle.Contract.SubmissionInterval(&_L2OutputOracle.CallOpts)
}

// SubmissionInterval is a free data retrieval call binding the contract method 0xe1a41bcf.
//
// Solidity: function submissionInterval() view returns(uint256)
func (_L2OutputOracle *L2OutputOracleCallerSession) SubmissionInterval() (*big.Int, error) {
	return _L2OutputOracle.Contract.SubmissionInterval(&_L2OutputOracle.CallOpts)
}

// AppendL2Output is a paid mutator transaction binding the contract method 0x25188104.
//
// Solidity: function appendL2Output(bytes32 _l2Output, uint256 _l2timestamp, bytes32 _blockhash, uint256 _blocknumber) payable returns()
func (_L2OutputOracle *L2OutputOracleTransactor) AppendL2Output(opts *bind.TransactOpts, _l2Output [32]byte, _l2timestamp *big.Int, _blockhash [32]byte, _blocknumber *big.Int) (*types.Transaction, error) {
	return _L2OutputOracle.contract.Transact(opts, "appendL2Output", _l2Output, _l2timestamp, _blockhash, _blocknumber)
}

// AppendL2Output is a paid mutator transaction binding the contract method 0x25188104.
//
// Solidity: function appendL2Output(bytes32 _l2Output, uint256 _l2timestamp, bytes32 _blockhash, uint256 _blocknumber) payable returns()
func (_L2OutputOracle *L2OutputOracleSession) AppendL2Output(_l2Output [32]byte, _l2timestamp *big.Int, _blockhash [32]byte, _blocknumber *big.Int) (*types.Transaction, error) {
	return _L2OutputOracle.Contract.AppendL2Output(&_L2OutputOracle.TransactOpts, _l2Output, _l2timestamp, _blockhash, _blocknumber)
}

// AppendL2Output is a paid mutator transaction binding the contract method 0x25188104.
//
// Solidity: function appendL2Output(bytes32 _l2Output, uint256 _l2timestamp, bytes32 _blockhash, uint256 _blocknumber) payable returns()
func (_L2OutputOracle *L2OutputOracleTransactorSession) AppendL2Output(_l2Output [32]byte, _l2timestamp *big.Int, _blockhash [32]byte, _blocknumber *big.Int) (*types.Transaction, error) {
	return _L2OutputOracle.Contract.AppendL2Output(&_L2OutputOracle.TransactOpts, _l2Output, _l2timestamp, _blockhash, _blocknumber)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_L2OutputOracle *L2OutputOracleTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _L2OutputOracle.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_L2OutputOracle *L2OutputOracleSession) RenounceOwnership() (*types.Transaction, error) {
	return _L2OutputOracle.Contract.RenounceOwnership(&_L2OutputOracle.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_L2OutputOracle *L2OutputOracleTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _L2OutputOracle.Contract.RenounceOwnership(&_L2OutputOracle.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_L2OutputOracle *L2OutputOracleTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _L2OutputOracle.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_L2OutputOracle *L2OutputOracleSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _L2OutputOracle.Contract.TransferOwnership(&_L2OutputOracle.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_L2OutputOracle *L2OutputOracleTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _L2OutputOracle.Contract.TransferOwnership(&_L2OutputOracle.TransactOpts, newOwner)
}

// L2OutputOracleOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the L2OutputOracle contract.
type L2OutputOracleOwnershipTransferredIterator struct {
	Event *L2OutputOracleOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *L2OutputOracleOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(L2OutputOracleOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(L2OutputOracleOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *L2OutputOracleOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *L2OutputOracleOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// L2OutputOracleOwnershipTransferred represents a OwnershipTransferred event raised by the L2OutputOracle contract.
type L2OutputOracleOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_L2OutputOracle *L2OutputOracleFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*L2OutputOracleOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _L2OutputOracle.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &L2OutputOracleOwnershipTransferredIterator{contract: _L2OutputOracle.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_L2OutputOracle *L2OutputOracleFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *L2OutputOracleOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _L2OutputOracle.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(L2OutputOracleOwnershipTransferred)
				if err := _L2OutputOracle.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_L2OutputOracle *L2OutputOracleFilterer) ParseOwnershipTransferred(log types.Log) (*L2OutputOracleOwnershipTransferred, error) {
	event := new(L2OutputOracleOwnershipTransferred)
	if err := _L2OutputOracle.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// L2OutputOracleL2OutputAppendedIterator is returned from FilterL2OutputAppended and is used to iterate over the raw logs and unpacked data for L2OutputAppended events raised by the L2OutputOracle contract.
type L2OutputOracleL2OutputAppendedIterator struct {
	Event *L2OutputOracleL2OutputAppended // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *L2OutputOracleL2OutputAppendedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(L2OutputOracleL2OutputAppended)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(L2OutputOracleL2OutputAppended)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *L2OutputOracleL2OutputAppendedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *L2OutputOracleL2OutputAppendedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// L2OutputOracleL2OutputAppended represents a L2OutputAppended event raised by the L2OutputOracle contract.
type L2OutputOracleL2OutputAppended struct {
	L2Output    [32]byte
	L2timestamp *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterL2OutputAppended is a free log retrieval operation binding the contract event 0x92701dc658a5d84c16077ea6de344b9995e21a96a05d45e4cd22f37a3d266f8b.
//
// Solidity: event l2OutputAppended(bytes32 indexed _l2Output, uint256 indexed _l2timestamp)
func (_L2OutputOracle *L2OutputOracleFilterer) FilterL2OutputAppended(opts *bind.FilterOpts, _l2Output [][32]byte, _l2timestamp []*big.Int) (*L2OutputOracleL2OutputAppendedIterator, error) {

	var _l2OutputRule []interface{}
	for _, _l2OutputItem := range _l2Output {
		_l2OutputRule = append(_l2OutputRule, _l2OutputItem)
	}
	var _l2timestampRule []interface{}
	for _, _l2timestampItem := range _l2timestamp {
		_l2timestampRule = append(_l2timestampRule, _l2timestampItem)
	}

	logs, sub, err := _L2OutputOracle.contract.FilterLogs(opts, "l2OutputAppended", _l2OutputRule, _l2timestampRule)
	if err != nil {
		return nil, err
	}
	return &L2OutputOracleL2OutputAppendedIterator{contract: _L2OutputOracle.contract, event: "l2OutputAppended", logs: logs, sub: sub}, nil
}

// WatchL2OutputAppended is a free log subscription operation binding the contract event 0x92701dc658a5d84c16077ea6de344b9995e21a96a05d45e4cd22f37a3d266f8b.
//
// Solidity: event l2OutputAppended(bytes32 indexed _l2Output, uint256 indexed _l2timestamp)
func (_L2OutputOracle *L2OutputOracleFilterer) WatchL2OutputAppended(opts *bind.WatchOpts, sink chan<- *L2OutputOracleL2OutputAppended, _l2Output [][32]byte, _l2timestamp []*big.Int) (event.Subscription, error) {

	var _l2OutputRule []interface{}
	for _, _l2OutputItem := range _l2Output {
		_l2OutputRule = append(_l2OutputRule, _l2OutputItem)
	}
	var _l2timestampRule []interface{}
	for _, _l2timestampItem := range _l2timestamp {
		_l2timestampRule = append(_l2timestampRule, _l2timestampItem)
	}

	logs, sub, err := _L2OutputOracle.contract.WatchLogs(opts, "l2OutputAppended", _l2OutputRule, _l2timestampRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(L2OutputOracleL2OutputAppended)
				if err := _L2OutputOracle.contract.UnpackLog(event, "l2OutputAppended", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseL2OutputAppended is a log parse operation binding the contract event 0x92701dc658a5d84c16077ea6de344b9995e21a96a05d45e4cd22f37a3d266f8b.
//
// Solidity: event l2OutputAppended(bytes32 indexed _l2Output, uint256 indexed _l2timestamp)
func (_L2OutputOracle *L2OutputOracleFilterer) ParseL2OutputAppended(log types.Log) (*L2OutputOracleL2OutputAppended, error) {
	event := new(L2OutputOracleL2OutputAppended)
	if err := _L2OutputOracle.contract.UnpackLog(event, "l2OutputAppended", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
