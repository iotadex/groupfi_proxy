// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package middleware

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

// ILuksoMetaData contains all meta data concerning the ILukso contract.
var ILuksoMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"dataKey\",\"type\":\"bytes32\"}],\"name\":\"getData\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// ILuksoABI is the input ABI used to generate the binding from.
// Deprecated: Use ILuksoMetaData.ABI instead.
var ILuksoABI = ILuksoMetaData.ABI

// ILukso is an auto generated Go binding around an Ethereum contract.
type ILukso struct {
	ILuksoCaller     // Read-only binding to the contract
	ILuksoTransactor // Write-only binding to the contract
	ILuksoFilterer   // Log filterer for contract events
}

// ILuksoCaller is an auto generated read-only Go binding around an Ethereum contract.
type ILuksoCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ILuksoTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ILuksoTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ILuksoFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ILuksoFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ILuksoSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ILuksoSession struct {
	Contract     *ILukso           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ILuksoCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ILuksoCallerSession struct {
	Contract *ILuksoCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// ILuksoTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ILuksoTransactorSession struct {
	Contract     *ILuksoTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ILuksoRaw is an auto generated low-level Go binding around an Ethereum contract.
type ILuksoRaw struct {
	Contract *ILukso // Generic contract binding to access the raw methods on
}

// ILuksoCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ILuksoCallerRaw struct {
	Contract *ILuksoCaller // Generic read-only contract binding to access the raw methods on
}

// ILuksoTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ILuksoTransactorRaw struct {
	Contract *ILuksoTransactor // Generic write-only contract binding to access the raw methods on
}

// NewILukso creates a new instance of ILukso, bound to a specific deployed contract.
func NewILukso(address common.Address, backend bind.ContractBackend) (*ILukso, error) {
	contract, err := bindILukso(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ILukso{ILuksoCaller: ILuksoCaller{contract: contract}, ILuksoTransactor: ILuksoTransactor{contract: contract}, ILuksoFilterer: ILuksoFilterer{contract: contract}}, nil
}

// NewILuksoCaller creates a new read-only instance of ILukso, bound to a specific deployed contract.
func NewILuksoCaller(address common.Address, caller bind.ContractCaller) (*ILuksoCaller, error) {
	contract, err := bindILukso(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ILuksoCaller{contract: contract}, nil
}

// NewILuksoTransactor creates a new write-only instance of ILukso, bound to a specific deployed contract.
func NewILuksoTransactor(address common.Address, transactor bind.ContractTransactor) (*ILuksoTransactor, error) {
	contract, err := bindILukso(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ILuksoTransactor{contract: contract}, nil
}

// NewILuksoFilterer creates a new log filterer instance of ILukso, bound to a specific deployed contract.
func NewILuksoFilterer(address common.Address, filterer bind.ContractFilterer) (*ILuksoFilterer, error) {
	contract, err := bindILukso(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ILuksoFilterer{contract: contract}, nil
}

// bindILukso binds a generic wrapper to an already deployed contract.
func bindILukso(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ILuksoABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ILukso *ILuksoRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ILukso.Contract.ILuksoCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ILukso *ILuksoRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ILukso.Contract.ILuksoTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ILukso *ILuksoRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ILukso.Contract.ILuksoTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ILukso *ILuksoCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ILukso.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ILukso *ILuksoTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ILukso.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ILukso *ILuksoTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ILukso.Contract.contract.Transact(opts, method, params...)
}

// GetData is a free data retrieval call binding the contract method 0x54f6127f.
//
// Solidity: function getData(bytes32 dataKey) view returns(bytes)
func (_ILukso *ILuksoCaller) GetData(opts *bind.CallOpts, dataKey [32]byte) ([]byte, error) {
	var out []interface{}
	err := _ILukso.contract.Call(opts, &out, "getData", dataKey)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetData is a free data retrieval call binding the contract method 0x54f6127f.
//
// Solidity: function getData(bytes32 dataKey) view returns(bytes)
func (_ILukso *ILuksoSession) GetData(dataKey [32]byte) ([]byte, error) {
	return _ILukso.Contract.GetData(&_ILukso.CallOpts, dataKey)
}

// GetData is a free data retrieval call binding the contract method 0x54f6127f.
//
// Solidity: function getData(bytes32 dataKey) view returns(bytes)
func (_ILukso *ILuksoCallerSession) GetData(dataKey [32]byte) ([]byte, error) {
	return _ILukso.Contract.GetData(&_ILukso.CallOpts, dataKey)
}
