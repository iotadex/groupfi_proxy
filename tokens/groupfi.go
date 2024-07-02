// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package tokens

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

// GroupFiMetaData contains all meta data concerning the GroupFi contract.
var GroupFiMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"w\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"pubkey\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amountIn\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"amountOut\",\"type\":\"uint64\"}],\"name\":\"BuySmr\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"ed25519\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"amount\",\"type\":\"uint64\"}],\"name\":\"buySmr\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"adds\",\"type\":\"address[]\"},{\"internalType\":\"address[]\",\"name\":\"subs\",\"type\":\"address[]\"},{\"internalType\":\"contractIERC20\",\"name\":\"c\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"threshold\",\"type\":\"uint256\"}],\"name\":\"checkERC20Group\",\"outputs\":[{\"internalType\":\"int8\",\"name\":\"res\",\"type\":\"int8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"adds\",\"type\":\"address[]\"},{\"internalType\":\"address[]\",\"name\":\"subs\",\"type\":\"address[]\"},{\"internalType\":\"contractIERC721\",\"name\":\"c\",\"type\":\"address\"}],\"name\":\"checkERC721Group\",\"outputs\":[{\"internalType\":\"int8\",\"name\":\"\",\"type\":\"int8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"adds\",\"type\":\"address[]\"},{\"internalType\":\"address[]\",\"name\":\"subs\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"threshold\",\"type\":\"uint256\"}],\"name\":\"checkEthGroup\",\"outputs\":[{\"internalType\":\"int8\",\"name\":\"res\",\"type\":\"int8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"addrs\",\"type\":\"address[]\"},{\"internalType\":\"contractIERC20\",\"name\":\"c\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"threshold\",\"type\":\"uint256\"}],\"name\":\"filterERC20Addresses\",\"outputs\":[{\"internalType\":\"uint16[]\",\"name\":\"indexes\",\"type\":\"uint16[]\"},{\"internalType\":\"uint16\",\"name\":\"count\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"addrs\",\"type\":\"address[]\"},{\"internalType\":\"contractIERC721\",\"name\":\"c\",\"type\":\"address\"}],\"name\":\"filterERC721Addresses\",\"outputs\":[{\"internalType\":\"uint16[]\",\"name\":\"indexes\",\"type\":\"uint16[]\"},{\"internalType\":\"uint16\",\"name\":\"count\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"addrs\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"threshold\",\"type\":\"uint256\"}],\"name\":\"filterEthAddresses\",\"outputs\":[{\"internalType\":\"uint16[]\",\"name\":\"indexes\",\"type\":\"uint16[]\"},{\"internalType\":\"uint16\",\"name\":\"count\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"wallet\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// GroupFiABI is the input ABI used to generate the binding from.
// Deprecated: Use GroupFiMetaData.ABI instead.
var GroupFiABI = GroupFiMetaData.ABI

// GroupFi is an auto generated Go binding around an Ethereum contract.
type GroupFi struct {
	GroupFiCaller     // Read-only binding to the contract
	GroupFiTransactor // Write-only binding to the contract
	GroupFiFilterer   // Log filterer for contract events
}

// GroupFiCaller is an auto generated read-only Go binding around an Ethereum contract.
type GroupFiCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GroupFiTransactor is an auto generated write-only Go binding around an Ethereum contract.
type GroupFiTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GroupFiFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type GroupFiFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GroupFiSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type GroupFiSession struct {
	Contract     *GroupFi          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GroupFiCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type GroupFiCallerSession struct {
	Contract *GroupFiCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// GroupFiTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type GroupFiTransactorSession struct {
	Contract     *GroupFiTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// GroupFiRaw is an auto generated low-level Go binding around an Ethereum contract.
type GroupFiRaw struct {
	Contract *GroupFi // Generic contract binding to access the raw methods on
}

// GroupFiCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type GroupFiCallerRaw struct {
	Contract *GroupFiCaller // Generic read-only contract binding to access the raw methods on
}

// GroupFiTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type GroupFiTransactorRaw struct {
	Contract *GroupFiTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGroupFi creates a new instance of GroupFi, bound to a specific deployed contract.
func NewGroupFi(address common.Address, backend bind.ContractBackend) (*GroupFi, error) {
	contract, err := bindGroupFi(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &GroupFi{GroupFiCaller: GroupFiCaller{contract: contract}, GroupFiTransactor: GroupFiTransactor{contract: contract}, GroupFiFilterer: GroupFiFilterer{contract: contract}}, nil
}

// NewGroupFiCaller creates a new read-only instance of GroupFi, bound to a specific deployed contract.
func NewGroupFiCaller(address common.Address, caller bind.ContractCaller) (*GroupFiCaller, error) {
	contract, err := bindGroupFi(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GroupFiCaller{contract: contract}, nil
}

// NewGroupFiTransactor creates a new write-only instance of GroupFi, bound to a specific deployed contract.
func NewGroupFiTransactor(address common.Address, transactor bind.ContractTransactor) (*GroupFiTransactor, error) {
	contract, err := bindGroupFi(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GroupFiTransactor{contract: contract}, nil
}

// NewGroupFiFilterer creates a new log filterer instance of GroupFi, bound to a specific deployed contract.
func NewGroupFiFilterer(address common.Address, filterer bind.ContractFilterer) (*GroupFiFilterer, error) {
	contract, err := bindGroupFi(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GroupFiFilterer{contract: contract}, nil
}

// bindGroupFi binds a generic wrapper to an already deployed contract.
func bindGroupFi(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(GroupFiABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GroupFi *GroupFiRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GroupFi.Contract.GroupFiCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GroupFi *GroupFiRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GroupFi.Contract.GroupFiTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GroupFi *GroupFiRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GroupFi.Contract.GroupFiTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GroupFi *GroupFiCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GroupFi.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GroupFi *GroupFiTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GroupFi.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GroupFi *GroupFiTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GroupFi.Contract.contract.Transact(opts, method, params...)
}

// CheckERC20Group is a free data retrieval call binding the contract method 0xaf19b943.
//
// Solidity: function checkERC20Group(address[] adds, address[] subs, address c, uint256 threshold) view returns(int8 res)
func (_GroupFi *GroupFiCaller) CheckERC20Group(opts *bind.CallOpts, adds []common.Address, subs []common.Address, c common.Address, threshold *big.Int) (int8, error) {
	var out []interface{}
	err := _GroupFi.contract.Call(opts, &out, "checkERC20Group", adds, subs, c, threshold)

	if err != nil {
		return *new(int8), err
	}

	out0 := *abi.ConvertType(out[0], new(int8)).(*int8)

	return out0, err

}

// CheckERC20Group is a free data retrieval call binding the contract method 0xaf19b943.
//
// Solidity: function checkERC20Group(address[] adds, address[] subs, address c, uint256 threshold) view returns(int8 res)
func (_GroupFi *GroupFiSession) CheckERC20Group(adds []common.Address, subs []common.Address, c common.Address, threshold *big.Int) (int8, error) {
	return _GroupFi.Contract.CheckERC20Group(&_GroupFi.CallOpts, adds, subs, c, threshold)
}

// CheckERC20Group is a free data retrieval call binding the contract method 0xaf19b943.
//
// Solidity: function checkERC20Group(address[] adds, address[] subs, address c, uint256 threshold) view returns(int8 res)
func (_GroupFi *GroupFiCallerSession) CheckERC20Group(adds []common.Address, subs []common.Address, c common.Address, threshold *big.Int) (int8, error) {
	return _GroupFi.Contract.CheckERC20Group(&_GroupFi.CallOpts, adds, subs, c, threshold)
}

// CheckERC721Group is a free data retrieval call binding the contract method 0x0ebec685.
//
// Solidity: function checkERC721Group(address[] adds, address[] subs, address c) view returns(int8)
func (_GroupFi *GroupFiCaller) CheckERC721Group(opts *bind.CallOpts, adds []common.Address, subs []common.Address, c common.Address) (int8, error) {
	var out []interface{}
	err := _GroupFi.contract.Call(opts, &out, "checkERC721Group", adds, subs, c)

	if err != nil {
		return *new(int8), err
	}

	out0 := *abi.ConvertType(out[0], new(int8)).(*int8)

	return out0, err

}

// CheckERC721Group is a free data retrieval call binding the contract method 0x0ebec685.
//
// Solidity: function checkERC721Group(address[] adds, address[] subs, address c) view returns(int8)
func (_GroupFi *GroupFiSession) CheckERC721Group(adds []common.Address, subs []common.Address, c common.Address) (int8, error) {
	return _GroupFi.Contract.CheckERC721Group(&_GroupFi.CallOpts, adds, subs, c)
}

// CheckERC721Group is a free data retrieval call binding the contract method 0x0ebec685.
//
// Solidity: function checkERC721Group(address[] adds, address[] subs, address c) view returns(int8)
func (_GroupFi *GroupFiCallerSession) CheckERC721Group(adds []common.Address, subs []common.Address, c common.Address) (int8, error) {
	return _GroupFi.Contract.CheckERC721Group(&_GroupFi.CallOpts, adds, subs, c)
}

// CheckEthGroup is a free data retrieval call binding the contract method 0x9978c27d.
//
// Solidity: function checkEthGroup(address[] adds, address[] subs, uint256 threshold) view returns(int8 res)
func (_GroupFi *GroupFiCaller) CheckEthGroup(opts *bind.CallOpts, adds []common.Address, subs []common.Address, threshold *big.Int) (int8, error) {
	var out []interface{}
	err := _GroupFi.contract.Call(opts, &out, "checkEthGroup", adds, subs, threshold)

	if err != nil {
		return *new(int8), err
	}

	out0 := *abi.ConvertType(out[0], new(int8)).(*int8)

	return out0, err

}

// CheckEthGroup is a free data retrieval call binding the contract method 0x9978c27d.
//
// Solidity: function checkEthGroup(address[] adds, address[] subs, uint256 threshold) view returns(int8 res)
func (_GroupFi *GroupFiSession) CheckEthGroup(adds []common.Address, subs []common.Address, threshold *big.Int) (int8, error) {
	return _GroupFi.Contract.CheckEthGroup(&_GroupFi.CallOpts, adds, subs, threshold)
}

// CheckEthGroup is a free data retrieval call binding the contract method 0x9978c27d.
//
// Solidity: function checkEthGroup(address[] adds, address[] subs, uint256 threshold) view returns(int8 res)
func (_GroupFi *GroupFiCallerSession) CheckEthGroup(adds []common.Address, subs []common.Address, threshold *big.Int) (int8, error) {
	return _GroupFi.Contract.CheckEthGroup(&_GroupFi.CallOpts, adds, subs, threshold)
}

// FilterERC20Addresses is a free data retrieval call binding the contract method 0xb88d5ea4.
//
// Solidity: function filterERC20Addresses(address[] addrs, address c, uint256 threshold) view returns(uint16[] indexes, uint16 count)
func (_GroupFi *GroupFiCaller) FilterERC20Addresses(opts *bind.CallOpts, addrs []common.Address, c common.Address, threshold *big.Int) (struct {
	Indexes []uint16
	Count   uint16
}, error) {
	var out []interface{}
	err := _GroupFi.contract.Call(opts, &out, "filterERC20Addresses", addrs, c, threshold)

	outstruct := new(struct {
		Indexes []uint16
		Count   uint16
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Indexes = *abi.ConvertType(out[0], new([]uint16)).(*[]uint16)
	outstruct.Count = *abi.ConvertType(out[1], new(uint16)).(*uint16)

	return *outstruct, err

}

// FilterERC20Addresses is a free data retrieval call binding the contract method 0xb88d5ea4.
//
// Solidity: function filterERC20Addresses(address[] addrs, address c, uint256 threshold) view returns(uint16[] indexes, uint16 count)
func (_GroupFi *GroupFiSession) FilterERC20Addresses(addrs []common.Address, c common.Address, threshold *big.Int) (struct {
	Indexes []uint16
	Count   uint16
}, error) {
	return _GroupFi.Contract.FilterERC20Addresses(&_GroupFi.CallOpts, addrs, c, threshold)
}

// FilterERC20Addresses is a free data retrieval call binding the contract method 0xb88d5ea4.
//
// Solidity: function filterERC20Addresses(address[] addrs, address c, uint256 threshold) view returns(uint16[] indexes, uint16 count)
func (_GroupFi *GroupFiCallerSession) FilterERC20Addresses(addrs []common.Address, c common.Address, threshold *big.Int) (struct {
	Indexes []uint16
	Count   uint16
}, error) {
	return _GroupFi.Contract.FilterERC20Addresses(&_GroupFi.CallOpts, addrs, c, threshold)
}

// FilterERC721Addresses is a free data retrieval call binding the contract method 0x66d61066.
//
// Solidity: function filterERC721Addresses(address[] addrs, address c) view returns(uint16[] indexes, uint16 count)
func (_GroupFi *GroupFiCaller) FilterERC721Addresses(opts *bind.CallOpts, addrs []common.Address, c common.Address) (struct {
	Indexes []uint16
	Count   uint16
}, error) {
	var out []interface{}
	err := _GroupFi.contract.Call(opts, &out, "filterERC721Addresses", addrs, c)

	outstruct := new(struct {
		Indexes []uint16
		Count   uint16
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Indexes = *abi.ConvertType(out[0], new([]uint16)).(*[]uint16)
	outstruct.Count = *abi.ConvertType(out[1], new(uint16)).(*uint16)

	return *outstruct, err

}

// FilterERC721Addresses is a free data retrieval call binding the contract method 0x66d61066.
//
// Solidity: function filterERC721Addresses(address[] addrs, address c) view returns(uint16[] indexes, uint16 count)
func (_GroupFi *GroupFiSession) FilterERC721Addresses(addrs []common.Address, c common.Address) (struct {
	Indexes []uint16
	Count   uint16
}, error) {
	return _GroupFi.Contract.FilterERC721Addresses(&_GroupFi.CallOpts, addrs, c)
}

// FilterERC721Addresses is a free data retrieval call binding the contract method 0x66d61066.
//
// Solidity: function filterERC721Addresses(address[] addrs, address c) view returns(uint16[] indexes, uint16 count)
func (_GroupFi *GroupFiCallerSession) FilterERC721Addresses(addrs []common.Address, c common.Address) (struct {
	Indexes []uint16
	Count   uint16
}, error) {
	return _GroupFi.Contract.FilterERC721Addresses(&_GroupFi.CallOpts, addrs, c)
}

// FilterEthAddresses is a free data retrieval call binding the contract method 0xd9524e5a.
//
// Solidity: function filterEthAddresses(address[] addrs, uint256 threshold) view returns(uint16[] indexes, uint16 count)
func (_GroupFi *GroupFiCaller) FilterEthAddresses(opts *bind.CallOpts, addrs []common.Address, threshold *big.Int) (struct {
	Indexes []uint16
	Count   uint16
}, error) {
	var out []interface{}
	err := _GroupFi.contract.Call(opts, &out, "filterEthAddresses", addrs, threshold)

	outstruct := new(struct {
		Indexes []uint16
		Count   uint16
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Indexes = *abi.ConvertType(out[0], new([]uint16)).(*[]uint16)
	outstruct.Count = *abi.ConvertType(out[1], new(uint16)).(*uint16)

	return *outstruct, err

}

// FilterEthAddresses is a free data retrieval call binding the contract method 0xd9524e5a.
//
// Solidity: function filterEthAddresses(address[] addrs, uint256 threshold) view returns(uint16[] indexes, uint16 count)
func (_GroupFi *GroupFiSession) FilterEthAddresses(addrs []common.Address, threshold *big.Int) (struct {
	Indexes []uint16
	Count   uint16
}, error) {
	return _GroupFi.Contract.FilterEthAddresses(&_GroupFi.CallOpts, addrs, threshold)
}

// FilterEthAddresses is a free data retrieval call binding the contract method 0xd9524e5a.
//
// Solidity: function filterEthAddresses(address[] addrs, uint256 threshold) view returns(uint16[] indexes, uint16 count)
func (_GroupFi *GroupFiCallerSession) FilterEthAddresses(addrs []common.Address, threshold *big.Int) (struct {
	Indexes []uint16
	Count   uint16
}, error) {
	return _GroupFi.Contract.FilterEthAddresses(&_GroupFi.CallOpts, addrs, threshold)
}

// Wallet is a free data retrieval call binding the contract method 0x521eb273.
//
// Solidity: function wallet() view returns(address)
func (_GroupFi *GroupFiCaller) Wallet(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _GroupFi.contract.Call(opts, &out, "wallet")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Wallet is a free data retrieval call binding the contract method 0x521eb273.
//
// Solidity: function wallet() view returns(address)
func (_GroupFi *GroupFiSession) Wallet() (common.Address, error) {
	return _GroupFi.Contract.Wallet(&_GroupFi.CallOpts)
}

// Wallet is a free data retrieval call binding the contract method 0x521eb273.
//
// Solidity: function wallet() view returns(address)
func (_GroupFi *GroupFiCallerSession) Wallet() (common.Address, error) {
	return _GroupFi.Contract.Wallet(&_GroupFi.CallOpts)
}

// BuySmr is a paid mutator transaction binding the contract method 0x300ff8fb.
//
// Solidity: function buySmr(bytes32 ed25519, uint64 amount) payable returns()
func (_GroupFi *GroupFiTransactor) BuySmr(opts *bind.TransactOpts, ed25519 [32]byte, amount uint64) (*types.Transaction, error) {
	return _GroupFi.contract.Transact(opts, "buySmr", ed25519, amount)
}

// BuySmr is a paid mutator transaction binding the contract method 0x300ff8fb.
//
// Solidity: function buySmr(bytes32 ed25519, uint64 amount) payable returns()
func (_GroupFi *GroupFiSession) BuySmr(ed25519 [32]byte, amount uint64) (*types.Transaction, error) {
	return _GroupFi.Contract.BuySmr(&_GroupFi.TransactOpts, ed25519, amount)
}

// BuySmr is a paid mutator transaction binding the contract method 0x300ff8fb.
//
// Solidity: function buySmr(bytes32 ed25519, uint64 amount) payable returns()
func (_GroupFi *GroupFiTransactorSession) BuySmr(ed25519 [32]byte, amount uint64) (*types.Transaction, error) {
	return _GroupFi.Contract.BuySmr(&_GroupFi.TransactOpts, ed25519, amount)
}

// GroupFiBuySmrIterator is returned from FilterBuySmr and is used to iterate over the raw logs and unpacked data for BuySmr events raised by the GroupFi contract.
type GroupFiBuySmrIterator struct {
	Event *GroupFiBuySmr // Event containing the contract specifics and raw log

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
func (it *GroupFiBuySmrIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GroupFiBuySmr)
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
		it.Event = new(GroupFiBuySmr)
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
func (it *GroupFiBuySmrIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GroupFiBuySmrIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GroupFiBuySmr represents a BuySmr event raised by the GroupFi contract.
type GroupFiBuySmr struct {
	User      common.Address
	Pubkey    [32]byte
	AmountIn  *big.Int
	AmountOut uint64
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterBuySmr is a free log retrieval operation binding the contract event 0x80b13b0b008cc1a1ca4fe2bf60e077b51203e806895da5e9eef74f477b8c5996.
//
// Solidity: event BuySmr(address indexed user, bytes32 indexed pubkey, uint256 amountIn, uint64 amountOut)
func (_GroupFi *GroupFiFilterer) FilterBuySmr(opts *bind.FilterOpts, user []common.Address, pubkey [][32]byte) (*GroupFiBuySmrIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var pubkeyRule []interface{}
	for _, pubkeyItem := range pubkey {
		pubkeyRule = append(pubkeyRule, pubkeyItem)
	}

	logs, sub, err := _GroupFi.contract.FilterLogs(opts, "BuySmr", userRule, pubkeyRule)
	if err != nil {
		return nil, err
	}
	return &GroupFiBuySmrIterator{contract: _GroupFi.contract, event: "BuySmr", logs: logs, sub: sub}, nil
}

// WatchBuySmr is a free log subscription operation binding the contract event 0x80b13b0b008cc1a1ca4fe2bf60e077b51203e806895da5e9eef74f477b8c5996.
//
// Solidity: event BuySmr(address indexed user, bytes32 indexed pubkey, uint256 amountIn, uint64 amountOut)
func (_GroupFi *GroupFiFilterer) WatchBuySmr(opts *bind.WatchOpts, sink chan<- *GroupFiBuySmr, user []common.Address, pubkey [][32]byte) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var pubkeyRule []interface{}
	for _, pubkeyItem := range pubkey {
		pubkeyRule = append(pubkeyRule, pubkeyItem)
	}

	logs, sub, err := _GroupFi.contract.WatchLogs(opts, "BuySmr", userRule, pubkeyRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GroupFiBuySmr)
				if err := _GroupFi.contract.UnpackLog(event, "BuySmr", log); err != nil {
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

// ParseBuySmr is a log parse operation binding the contract event 0x80b13b0b008cc1a1ca4fe2bf60e077b51203e806895da5e9eef74f477b8c5996.
//
// Solidity: event BuySmr(address indexed user, bytes32 indexed pubkey, uint256 amountIn, uint64 amountOut)
func (_GroupFi *GroupFiFilterer) ParseBuySmr(log types.Log) (*GroupFiBuySmr, error) {
	event := new(GroupFiBuySmr)
	if err := _GroupFi.contract.UnpackLog(event, "BuySmr", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
