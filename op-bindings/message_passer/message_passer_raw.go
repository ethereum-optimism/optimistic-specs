// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package messagePasser

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

// MessagePasserMetaData contains all meta data concerning the MessagePasser contract.
var MessagePasserMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"MessagePasserBalanceBurnt\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"WithdrawalInitiated\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_target\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_gasLimit\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_data\",\"type\":\"bytes\"}],\"name\":\"initiateWithdrawal\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"nonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"sentMessages\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Bin: "0x608060405234801561001057600080fd5b506104a1806100206000396000f3fe6080604052600436106100435760003560e01c806344df8e701461006c57806382e3702d14610081578063affed0e0146100c6578063c2b3e5ac146100ea57600080fd5b366100675761006533620186a0604051806020016040528060008152506100f8565b005b600080fd5b34801561007857600080fd5b506100656101ae565b34801561008d57600080fd5b506100b161009c366004610256565b60006020819052908152604090205460ff1681565b60405190151581526020015b60405180910390f35b3480156100d257600080fd5b506100dc60015481565b6040519081526020016100bd565b6100656100f836600461029e565b600061010a600154338634878761020b565b6000818152602081905260409081902080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0016600190811790915554905191925073ffffffffffffffffffffffffffffffffffffffff8616913391907f87bf7b546c8de873abb0db5b579ec131f8d0cf5b14f39933551cf9ced23a6136906101989034908990899061040d565b60405180910390a4505060018054810190555050565b604051479081906101be9061024a565b6040518091039082f09050801580156101db573d6000803e3d6000fd5b505060405181907fa803ee038cd71f0e98c4aef6e78c05c7c44b5fae28e0acfb66190aea75565cff90600090a250565b600086868686868660405160200161022896959493929190610435565b6040516020818303038152906040528051906020012090509695505050505050565b60088061048d83390190565b60006020828403121561026857600080fd5b5035919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6000806000606084860312156102b357600080fd5b833573ffffffffffffffffffffffffffffffffffffffff811681146102d757600080fd5b925060208401359150604084013567ffffffffffffffff808211156102fb57600080fd5b818601915086601f83011261030f57600080fd5b8135818111156103215761032161026f565b604051601f82017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0908116603f011681019083821181831017156103675761036761026f565b8160405282815289602084870101111561038057600080fd5b8260208601602083013760006020848301015280955050505050509250925092565b6000815180845260005b818110156103c8576020818501810151868301820152016103ac565b818111156103da576000602083870101525b50601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0169290920160200192915050565b83815282602082015260606040820152600061042c60608301846103a2565b95945050505050565b868152600073ffffffffffffffffffffffffffffffffffffffff808816602084015280871660408401525084606083015283608083015260c060a083015261048060c08301846103a2565b9897505050505050505056fe608060405230fffea164736f6c634300080a000a",
}

// MessagePasserABI is the input ABI used to generate the binding from.
// Deprecated: Use MessagePasserMetaData.ABI instead.
var MessagePasserABI = MessagePasserMetaData.ABI

// MessagePasserBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use MessagePasserMetaData.Bin instead.
var MessagePasserBin = MessagePasserMetaData.Bin

// DeployMessagePasser deploys a new Ethereum contract, binding an instance of MessagePasser to it.
func DeployMessagePasser(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *MessagePasser, error) {
	parsed, err := MessagePasserMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(MessagePasserBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &MessagePasser{MessagePasserCaller: MessagePasserCaller{contract: contract}, MessagePasserTransactor: MessagePasserTransactor{contract: contract}, MessagePasserFilterer: MessagePasserFilterer{contract: contract}}, nil
}

// MessagePasser is an auto generated Go binding around an Ethereum contract.
type MessagePasser struct {
	MessagePasserCaller     // Read-only binding to the contract
	MessagePasserTransactor // Write-only binding to the contract
	MessagePasserFilterer   // Log filterer for contract events
}

// MessagePasserCaller is an auto generated read-only Go binding around an Ethereum contract.
type MessagePasserCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MessagePasserTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MessagePasserTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MessagePasserFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MessagePasserFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MessagePasserSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MessagePasserSession struct {
	Contract     *MessagePasser    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MessagePasserCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MessagePasserCallerSession struct {
	Contract *MessagePasserCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// MessagePasserTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MessagePasserTransactorSession struct {
	Contract     *MessagePasserTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// MessagePasserRaw is an auto generated low-level Go binding around an Ethereum contract.
type MessagePasserRaw struct {
	Contract *MessagePasser // Generic contract binding to access the raw methods on
}

// MessagePasserCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MessagePasserCallerRaw struct {
	Contract *MessagePasserCaller // Generic read-only contract binding to access the raw methods on
}

// MessagePasserTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MessagePasserTransactorRaw struct {
	Contract *MessagePasserTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMessagePasser creates a new instance of MessagePasser, bound to a specific deployed contract.
func NewMessagePasser(address common.Address, backend bind.ContractBackend) (*MessagePasser, error) {
	contract, err := bindMessagePasser(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MessagePasser{MessagePasserCaller: MessagePasserCaller{contract: contract}, MessagePasserTransactor: MessagePasserTransactor{contract: contract}, MessagePasserFilterer: MessagePasserFilterer{contract: contract}}, nil
}

// NewMessagePasserCaller creates a new read-only instance of MessagePasser, bound to a specific deployed contract.
func NewMessagePasserCaller(address common.Address, caller bind.ContractCaller) (*MessagePasserCaller, error) {
	contract, err := bindMessagePasser(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MessagePasserCaller{contract: contract}, nil
}

// NewMessagePasserTransactor creates a new write-only instance of MessagePasser, bound to a specific deployed contract.
func NewMessagePasserTransactor(address common.Address, transactor bind.ContractTransactor) (*MessagePasserTransactor, error) {
	contract, err := bindMessagePasser(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MessagePasserTransactor{contract: contract}, nil
}

// NewMessagePasserFilterer creates a new log filterer instance of MessagePasser, bound to a specific deployed contract.
func NewMessagePasserFilterer(address common.Address, filterer bind.ContractFilterer) (*MessagePasserFilterer, error) {
	contract, err := bindMessagePasser(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MessagePasserFilterer{contract: contract}, nil
}

// bindMessagePasser binds a generic wrapper to an already deployed contract.
func bindMessagePasser(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(MessagePasserABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MessagePasser *MessagePasserRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MessagePasser.Contract.MessagePasserCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MessagePasser *MessagePasserRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MessagePasser.Contract.MessagePasserTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MessagePasser *MessagePasserRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MessagePasser.Contract.MessagePasserTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MessagePasser *MessagePasserCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MessagePasser.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MessagePasser *MessagePasserTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MessagePasser.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MessagePasser *MessagePasserTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MessagePasser.Contract.contract.Transact(opts, method, params...)
}

// Nonce is a free data retrieval call binding the contract method 0xaffed0e0.
//
// Solidity: function nonce() view returns(uint256)
func (_MessagePasser *MessagePasserCaller) Nonce(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MessagePasser.contract.Call(opts, &out, "nonce")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Nonce is a free data retrieval call binding the contract method 0xaffed0e0.
//
// Solidity: function nonce() view returns(uint256)
func (_MessagePasser *MessagePasserSession) Nonce() (*big.Int, error) {
	return _MessagePasser.Contract.Nonce(&_MessagePasser.CallOpts)
}

// Nonce is a free data retrieval call binding the contract method 0xaffed0e0.
//
// Solidity: function nonce() view returns(uint256)
func (_MessagePasser *MessagePasserCallerSession) Nonce() (*big.Int, error) {
	return _MessagePasser.Contract.Nonce(&_MessagePasser.CallOpts)
}

// SentMessages is a free data retrieval call binding the contract method 0x82e3702d.
//
// Solidity: function sentMessages(bytes32 ) view returns(bool)
func (_MessagePasser *MessagePasserCaller) SentMessages(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var out []interface{}
	err := _MessagePasser.contract.Call(opts, &out, "sentMessages", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SentMessages is a free data retrieval call binding the contract method 0x82e3702d.
//
// Solidity: function sentMessages(bytes32 ) view returns(bool)
func (_MessagePasser *MessagePasserSession) SentMessages(arg0 [32]byte) (bool, error) {
	return _MessagePasser.Contract.SentMessages(&_MessagePasser.CallOpts, arg0)
}

// SentMessages is a free data retrieval call binding the contract method 0x82e3702d.
//
// Solidity: function sentMessages(bytes32 ) view returns(bool)
func (_MessagePasser *MessagePasserCallerSession) SentMessages(arg0 [32]byte) (bool, error) {
	return _MessagePasser.Contract.SentMessages(&_MessagePasser.CallOpts, arg0)
}

// Burn is a paid mutator transaction binding the contract method 0x44df8e70.
//
// Solidity: function burn() returns()
func (_MessagePasser *MessagePasserTransactor) Burn(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MessagePasser.contract.Transact(opts, "burn")
}

// Burn is a paid mutator transaction binding the contract method 0x44df8e70.
//
// Solidity: function burn() returns()
func (_MessagePasser *MessagePasserSession) Burn() (*types.Transaction, error) {
	return _MessagePasser.Contract.Burn(&_MessagePasser.TransactOpts)
}

// Burn is a paid mutator transaction binding the contract method 0x44df8e70.
//
// Solidity: function burn() returns()
func (_MessagePasser *MessagePasserTransactorSession) Burn() (*types.Transaction, error) {
	return _MessagePasser.Contract.Burn(&_MessagePasser.TransactOpts)
}

// InitiateWithdrawal is a paid mutator transaction binding the contract method 0xc2b3e5ac.
//
// Solidity: function initiateWithdrawal(address _target, uint256 _gasLimit, bytes _data) payable returns()
func (_MessagePasser *MessagePasserTransactor) InitiateWithdrawal(opts *bind.TransactOpts, _target common.Address, _gasLimit *big.Int, _data []byte) (*types.Transaction, error) {
	return _MessagePasser.contract.Transact(opts, "initiateWithdrawal", _target, _gasLimit, _data)
}

// InitiateWithdrawal is a paid mutator transaction binding the contract method 0xc2b3e5ac.
//
// Solidity: function initiateWithdrawal(address _target, uint256 _gasLimit, bytes _data) payable returns()
func (_MessagePasser *MessagePasserSession) InitiateWithdrawal(_target common.Address, _gasLimit *big.Int, _data []byte) (*types.Transaction, error) {
	return _MessagePasser.Contract.InitiateWithdrawal(&_MessagePasser.TransactOpts, _target, _gasLimit, _data)
}

// InitiateWithdrawal is a paid mutator transaction binding the contract method 0xc2b3e5ac.
//
// Solidity: function initiateWithdrawal(address _target, uint256 _gasLimit, bytes _data) payable returns()
func (_MessagePasser *MessagePasserTransactorSession) InitiateWithdrawal(_target common.Address, _gasLimit *big.Int, _data []byte) (*types.Transaction, error) {
	return _MessagePasser.Contract.InitiateWithdrawal(&_MessagePasser.TransactOpts, _target, _gasLimit, _data)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_MessagePasser *MessagePasserTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MessagePasser.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_MessagePasser *MessagePasserSession) Receive() (*types.Transaction, error) {
	return _MessagePasser.Contract.Receive(&_MessagePasser.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_MessagePasser *MessagePasserTransactorSession) Receive() (*types.Transaction, error) {
	return _MessagePasser.Contract.Receive(&_MessagePasser.TransactOpts)
}

// MessagePasserMessagePasserBalanceBurntIterator is returned from FilterMessagePasserBalanceBurnt and is used to iterate over the raw logs and unpacked data for MessagePasserBalanceBurnt events raised by the MessagePasser contract.
type MessagePasserMessagePasserBalanceBurntIterator struct {
	Event *MessagePasserMessagePasserBalanceBurnt // Event containing the contract specifics and raw log

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
func (it *MessagePasserMessagePasserBalanceBurntIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MessagePasserMessagePasserBalanceBurnt)
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
		it.Event = new(MessagePasserMessagePasserBalanceBurnt)
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
func (it *MessagePasserMessagePasserBalanceBurntIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MessagePasserMessagePasserBalanceBurntIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MessagePasserMessagePasserBalanceBurnt represents a MessagePasserBalanceBurnt event raised by the MessagePasser contract.
type MessagePasserMessagePasserBalanceBurnt struct {
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterMessagePasserBalanceBurnt is a free log retrieval operation binding the contract event 0xa803ee038cd71f0e98c4aef6e78c05c7c44b5fae28e0acfb66190aea75565cff.
//
// Solidity: event MessagePasserBalanceBurnt(uint256 indexed amount)
func (_MessagePasser *MessagePasserFilterer) FilterMessagePasserBalanceBurnt(opts *bind.FilterOpts, amount []*big.Int) (*MessagePasserMessagePasserBalanceBurntIterator, error) {

	var amountRule []interface{}
	for _, amountItem := range amount {
		amountRule = append(amountRule, amountItem)
	}

	logs, sub, err := _MessagePasser.contract.FilterLogs(opts, "MessagePasserBalanceBurnt", amountRule)
	if err != nil {
		return nil, err
	}
	return &MessagePasserMessagePasserBalanceBurntIterator{contract: _MessagePasser.contract, event: "MessagePasserBalanceBurnt", logs: logs, sub: sub}, nil
}

// WatchMessagePasserBalanceBurnt is a free log subscription operation binding the contract event 0xa803ee038cd71f0e98c4aef6e78c05c7c44b5fae28e0acfb66190aea75565cff.
//
// Solidity: event MessagePasserBalanceBurnt(uint256 indexed amount)
func (_MessagePasser *MessagePasserFilterer) WatchMessagePasserBalanceBurnt(opts *bind.WatchOpts, sink chan<- *MessagePasserMessagePasserBalanceBurnt, amount []*big.Int) (event.Subscription, error) {

	var amountRule []interface{}
	for _, amountItem := range amount {
		amountRule = append(amountRule, amountItem)
	}

	logs, sub, err := _MessagePasser.contract.WatchLogs(opts, "MessagePasserBalanceBurnt", amountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MessagePasserMessagePasserBalanceBurnt)
				if err := _MessagePasser.contract.UnpackLog(event, "MessagePasserBalanceBurnt", log); err != nil {
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

// ParseMessagePasserBalanceBurnt is a log parse operation binding the contract event 0xa803ee038cd71f0e98c4aef6e78c05c7c44b5fae28e0acfb66190aea75565cff.
//
// Solidity: event MessagePasserBalanceBurnt(uint256 indexed amount)
func (_MessagePasser *MessagePasserFilterer) ParseMessagePasserBalanceBurnt(log types.Log) (*MessagePasserMessagePasserBalanceBurnt, error) {
	event := new(MessagePasserMessagePasserBalanceBurnt)
	if err := _MessagePasser.contract.UnpackLog(event, "MessagePasserBalanceBurnt", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MessagePasserWithdrawalInitiatedIterator is returned from FilterWithdrawalInitiated and is used to iterate over the raw logs and unpacked data for WithdrawalInitiated events raised by the MessagePasser contract.
type MessagePasserWithdrawalInitiatedIterator struct {
	Event *MessagePasserWithdrawalInitiated // Event containing the contract specifics and raw log

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
func (it *MessagePasserWithdrawalInitiatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MessagePasserWithdrawalInitiated)
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
		it.Event = new(MessagePasserWithdrawalInitiated)
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
func (it *MessagePasserWithdrawalInitiatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MessagePasserWithdrawalInitiatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MessagePasserWithdrawalInitiated represents a WithdrawalInitiated event raised by the MessagePasser contract.
type MessagePasserWithdrawalInitiated struct {
	Nonce    *big.Int
	Sender   common.Address
	Target   common.Address
	Value    *big.Int
	GasLimit *big.Int
	Data     []byte
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterWithdrawalInitiated is a free log retrieval operation binding the contract event 0x87bf7b546c8de873abb0db5b579ec131f8d0cf5b14f39933551cf9ced23a6136.
//
// Solidity: event WithdrawalInitiated(uint256 indexed nonce, address indexed sender, address indexed target, uint256 value, uint256 gasLimit, bytes data)
func (_MessagePasser *MessagePasserFilterer) FilterWithdrawalInitiated(opts *bind.FilterOpts, nonce []*big.Int, sender []common.Address, target []common.Address) (*MessagePasserWithdrawalInitiatedIterator, error) {

	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var targetRule []interface{}
	for _, targetItem := range target {
		targetRule = append(targetRule, targetItem)
	}

	logs, sub, err := _MessagePasser.contract.FilterLogs(opts, "WithdrawalInitiated", nonceRule, senderRule, targetRule)
	if err != nil {
		return nil, err
	}
	return &MessagePasserWithdrawalInitiatedIterator{contract: _MessagePasser.contract, event: "WithdrawalInitiated", logs: logs, sub: sub}, nil
}

// WatchWithdrawalInitiated is a free log subscription operation binding the contract event 0x87bf7b546c8de873abb0db5b579ec131f8d0cf5b14f39933551cf9ced23a6136.
//
// Solidity: event WithdrawalInitiated(uint256 indexed nonce, address indexed sender, address indexed target, uint256 value, uint256 gasLimit, bytes data)
func (_MessagePasser *MessagePasserFilterer) WatchWithdrawalInitiated(opts *bind.WatchOpts, sink chan<- *MessagePasserWithdrawalInitiated, nonce []*big.Int, sender []common.Address, target []common.Address) (event.Subscription, error) {

	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}
	var targetRule []interface{}
	for _, targetItem := range target {
		targetRule = append(targetRule, targetItem)
	}

	logs, sub, err := _MessagePasser.contract.WatchLogs(opts, "WithdrawalInitiated", nonceRule, senderRule, targetRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MessagePasserWithdrawalInitiated)
				if err := _MessagePasser.contract.UnpackLog(event, "WithdrawalInitiated", log); err != nil {
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

// ParseWithdrawalInitiated is a log parse operation binding the contract event 0x87bf7b546c8de873abb0db5b579ec131f8d0cf5b14f39933551cf9ced23a6136.
//
// Solidity: event WithdrawalInitiated(uint256 indexed nonce, address indexed sender, address indexed target, uint256 value, uint256 gasLimit, bytes data)
func (_MessagePasser *MessagePasserFilterer) ParseWithdrawalInitiated(log types.Log) (*MessagePasserWithdrawalInitiated, error) {
	event := new(MessagePasserWithdrawalInitiated)
	if err := _MessagePasser.contract.UnpackLog(event, "WithdrawalInitiated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
