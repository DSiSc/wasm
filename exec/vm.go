// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package exec provides functions for executing WebAssembly bytecode.
package exec

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"bytes"
	"encoding/json"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/crypto-suite/crypto"
	"github.com/DSiSc/repository"
	"github.com/DSiSc/wasm/disasm"
	"github.com/DSiSc/wasm/exec/internal/compile"
	"github.com/DSiSc/wasm/exec/memory"
	"github.com/DSiSc/wasm/wasm"
	ops "github.com/DSiSc/wasm/wasm/operators"
	"math/big"
)

var (
	// ErrMultipleLinearMemories is returned by (*VMInterpreter).NewInterpreter when the module
	// has more then one entries in the linear Mem space.
	ErrMultipleLinearMemories = errors.New("exec: more than one linear memories in module")
	// ErrInvalidArgumentCount is returned by (*VMInterpreter).ExecCode when an invalid
	// number of arguments to the WebAssembly function are passed to it.
	ErrInvalidArgumentCount = errors.New("exec: invalid number of arguments to function")

	ErrOutOfGas                 = errors.New("out of gas")
	ErrCodeSizeExceedLimit      = errors.New("code size exceed the contract limit")
	ErrDepth                    = errors.New("max call depth exceeded")
	ErrTraceLimitReached        = errors.New("the number of logs reached the specified limit")
	ErrInsufficientBalance      = errors.New("insufficient balance for transfer")
	ErrContractAddressCollision = errors.New("contract address collision")
	ErrNoCompatibleInterpreter  = errors.New("no compatible interpreter")
)

// emptyCodeHash is used by create to ensure deployment is disallowed to already
// deployed contract addresses (relevant after the account abstraction).
var emptyCodeHash = crypto.Keccak256Hash(nil)

// InvalidReturnTypeError is returned by (*VMInterpreter).ExecCode when the module
// specifies an invalid return type value for the executed function.
type InvalidReturnTypeError int8

func (e InvalidReturnTypeError) Error() string {
	return fmt.Sprintf("Function has invalid return value_type: %d", int8(e))
}

// InvalidFunctionIndexError is returned by (*VMInterpreter).ExecCode when the function
// index provided is invalid.
type InvalidFunctionIndexError int64

func (e InvalidFunctionIndexError) Error() string {
	return fmt.Sprintf("Invalid index to function index space: %d", int64(e))
}

type context struct {
	stack   []uint64
	locals  []uint64
	code    []byte
	asm     []asmBlock
	pc      int64
	curFunc int64
}

//VM WebAssembly embedder.
type VM struct {
	ChainContext *WasmChainContext
	StateDB      *repository.Repository
}

// NewVM creates a new VM instance.
func NewVM(chainContext *WasmChainContext, state *repository.Repository) *VM {
	return &VM{
		ChainContext: chainContext,
		StateDB:      state,
	}
}

// Create creates a new contract using code as deployment code.
func (self *VM) Create(caller types.Address, code []byte, gas uint64, value *big.Int) (ret []byte, contractAddr types.Address, leftOverGas uint64, err error) {
	_, err = wasm.ReadModule(bytes.NewReader(code), NativeResolve)
	if err != nil {
		log.Error("failed to read the module from code, as: %v", err)
		return nil, types.Address{}, gas, err
	}
	//generate contract address
	nonce := self.StateDB.GetNonce(caller)
	contractAddr = crypto.CreateAddress(caller, nonce)

	//check balance
	if !CanTransfer(self.StateDB, caller, value) {
		return nil, types.Address{}, gas, ErrInsufficientBalance
	}

	//update caller nonce
	self.StateDB.SetNonce(caller, nonce+1)

	// Ensure there's no existing contract already at the designated address
	contractHash := self.StateDB.GetCodeHash(contractAddr)
	if self.StateDB.GetNonce(contractAddr) != 0 || (contractHash != (types.Hash{}) && contractHash != emptyCodeHash) {
		return nil, types.Address{}, 0, ErrContractAddressCollision
	}

	// check whether the max code size has been exceeded
	if len(code) > MaxCodeSize {
		return nil, types.Address{}, 0, ErrCodeSizeExceedLimit
	}

	// Create a new account on the state
	self.StateDB.CreateAccount(contractAddr)
	self.StateDB.SetCode(contractAddr, code)
	Transfer(self.StateDB, caller, contractAddr, value)
	return nil, contractAddr, gas, nil
}

const entryPointMethod = "invoke"

// Call executes the contract associated with the addr with the given input as
// parameters. It also handles any necessary value transfer required and takes
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.
func (self *VM) Call(caller, addr types.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
	// Fail if we're trying to transfer more than the available balance
	if !CanTransfer(self.StateDB, caller, value) {
		return nil, gas, ErrInsufficientBalance
	}
	Transfer(self.StateDB, caller, addr, value)

	code := self.StateDB.GetCode(addr)
	m, err := wasm.ReadModule(bytes.NewReader(code), NativeResolve)
	if err != nil {
		log.Error("failed to read the module from code, as: %v", err)
		return nil, gas, err
	}

	interpreter, err := NewVMIntepreter(self, m)
	if err != nil {
		log.Error("failed to create wasm interpreter, as: %v", err)
		return nil, gas, err
	}

	var fnIndex int64
	if fnEntry, ok := m.Export.Entries[entryPointMethod]; ok {
		fnIndex = int64(fnEntry.Index)
	} else {
		return nil, gas, errors.New("failed to find the contract entry method.")
	}

	args, err := extractParams(input)
	if err != nil {
		return nil, gas, fmt.Errorf("failed to extract input params, as:%v.", err)
	}

	argPs := make([]int, len(args))
	for index, arg := range args {
		argP, err := interpreter.GetMemory().SetPointerMemory(arg)
		if err != nil {
			return nil, gas, fmt.Errorf("failed to pass input params to contract, as:%v.", err)
		}
		argPs[index] = argP
	}
	argPPs, err := interpreter.GetMemory().SetPointerMemory(argPs)
	if err != nil {
		return nil, gas, fmt.Errorf("failed to pass input params to contract, as:%v.", err)
	}

	retP, err := interpreter.ExecCode(fnIndex, uint64(len(args)), uint64(argPPs))
	if err != nil {
		return nil, gas, fmt.Errorf("failed to execute contract, as:%v.", err)
	}

	ret, err = interpreter.GetMemory().GetMemory(uint64(retP.(uint32)))
	if err != nil {
		return nil, gas, fmt.Errorf("failed to parse contract retrun value, as:%v.", err)
	}
	return ret, gas, err
}

// extract call input params
func extractParams(input []byte) ([]string, error) {
	var params []string
	err := json.Unmarshal(input, &params)
	if err != nil {
		return nil, err
	}
	return params, nil
}

// VMInterpreter is the execution context for executing WebAssembly bytecode.
type VMInterpreter struct {
	ChainContext *WasmChainContext
	StateDB      *repository.Repository

	ctx context

	module  *wasm.Module
	globals []uint64
	Mem     *memory.VMmemory
	funcs   []function

	funcTable [256]func()

	// RecoverPanic controls whether the `ExecCode` method
	// recovers from a panic and returns it as an error
	// instead.
	// A panic can occur either when executing an invalid VMInterpreter
	// or encountering an invalid instruction, e.g. `unreachable`.
	RecoverPanic bool

	abort bool // Flag for host functions to terminate execution

	nativeBackend *nativeCompiler
}

// As per the WebAssembly spec: https://github.com/WebAssembly/design/blob/27ac254c854994103c24834a994be16f74f54186/Semantics.md#linear-memory
const wasmPageSize = 65536 // (64 KB)

var endianess = binary.LittleEndian

type config struct {
	EnableAOT bool
}

// VMOption describes a customization that can be applied to the VMInterpreter.
type VMOption func(c *config)

// EnableAOT enables ahead-of-time compilation of supported opcodes
// into runs of native instructions, if wagon supports native compilation
// for the current architecture.
func EnableAOT(v bool) VMOption {
	return func(c *config) {
		c.EnableAOT = v
	}
}

// NewVMIntepreter create interpreter instance running in specified vm env.
func NewVMIntepreter(vm *VM, module *wasm.Module, opts ...VMOption) (*VMInterpreter, error) {
	interpreter, err := NewInterpreter(module, opts...)
	if err != nil {
		return nil, err
	}
	interpreter.StateDB = vm.StateDB
	interpreter.ChainContext = vm.ChainContext
	return interpreter, nil
}

// NewInterpreter creates a new interpreter from a given module and options. If the module defines
// a start function, it will be executed.
func NewInterpreter(module *wasm.Module, opts ...VMOption) (*VMInterpreter, error) {
	var vm VMInterpreter
	var options config
	for _, opt := range opts {
		opt(&options)
	}

	// init module memory
	vmMem, err := initVMMemory(module)
	if err != nil {
		return nil, err
	}
	vm.Mem = vmMem

	vm.funcs = make([]function, len(module.FunctionIndexSpace))
	vm.globals = make([]uint64, len(module.GlobalIndexSpace))
	vm.newFuncTable()
	vm.module = module

	nNatives := 0
	for i, fn := range module.FunctionIndexSpace {
		// Skip native methods as they need not be
		// disassembled; simply add them at the end
		// of the `funcs` array as is, as specified
		// in the spec. See the "host functions"
		// section of:
		// https://webassembly.github.io/spec/core/exec/modules.html#allocation
		if fn.IsHost() {
			vm.funcs[i] = goFunction{
				typ: fn.Host.Type(),
				val: fn.Host,
			}
			nNatives++
			continue
		}

		disassembly, err := disasm.NewDisassembly(fn, module)
		if err != nil {
			return nil, err
		}

		totalLocalVars := 0
		totalLocalVars += len(fn.Sig.ParamTypes)
		for _, entry := range fn.Body.Locals {
			totalLocalVars += int(entry.Count)
		}
		code, meta := compile.Compile(disassembly.Code)
		vm.funcs[i] = compiledFunction{
			codeMeta:       meta,
			code:           code,
			branchTables:   meta.BranchTables,
			maxDepth:       disassembly.MaxDepth,
			totalLocalVars: totalLocalVars,
			args:           len(fn.Sig.ParamTypes),
			returns:        len(fn.Sig.ReturnTypes) != 0,
		}
	}

	if err := vm.resetGlobals(); err != nil {
		return nil, err
	}

	if module.Start != nil {
		_, err := vm.ExecCode(int64(module.Start.Index))
		if err != nil {
			return nil, err
		}
	}

	if options.EnableAOT {
		supportedBackend, backend := nativeBackend()
		if supportedBackend {
			vm.nativeBackend = backend
			if err := vm.tryNativeCompile(); err != nil {
				return nil, err
			}
		}
	}

	return &vm, nil
}

// init wasm runtime memory
func initVMMemory(module *wasm.Module) (*memory.VMmemory, error) {

	vmMem := &memory.VMmemory{}
	if module.Memory != nil && len(module.Memory.Entries) != 0 {
		if len(module.Memory.Entries) > 1 {
			return nil, ErrMultipleLinearMemories
		}
		vmMem.ByteMem = make([]byte, uint(module.Memory.Entries[0].Limits.Initial)*wasmPageSize)
		copy(vmMem.ByteMem, module.LinearMemoryIndexSpace[0])
	}

	//give a default memory even if no memory section exist in wasm file
	if vmMem.ByteMem == nil {
		vmMem.ByteMem = make([]byte, 1*wasmPageSize)
	}

	vmMem.MemPoints = make(map[uint64]*memory.TypeLength) //init the pointer map

	//solve the Data section
	//this section is for some const strings, just like heap
	if module.Data != nil {
		var tmpIdx int
		for _, entry := range module.Data.Entries {
			if entry.Index != 0 {
				return nil, errors.New("invalid data index")
			}
			val, err := module.ExecInitExpr(entry.Offset)
			if err != nil {
				return nil, err
			}
			offset, ok := val.(int32)
			tmpIdx += int(offset) + len(entry.Data)
			if !ok {
				return nil, errors.New("invalid data index")
			}
			// for the case of " (data (get_global 0) "init\00init success!\00add\00int"))"
			if bytes.Contains(entry.Data, []byte{byte(0)}) {
				splited := bytes.Split(entry.Data, []byte{byte(0)})
				var tmpoffset = int(offset)
				for _, tmp := range splited {
					vmMem.MemPoints[uint64(tmpoffset)] = &memory.TypeLength{Ptype: memory.PString, Length: len(tmp) + 1}
					tmpoffset += len(tmp) + 1
				}
			} else {
				vmMem.MemPoints[uint64(offset)] = &memory.TypeLength{Ptype: memory.PString, Length: len(entry.Data)}
			}
		}
		//
		vmMem.AllocedMemIdex = tmpIdx
		vmMem.PointedMemIndex = (len(vmMem.ByteMem) + tmpIdx) / 2
	} else {
		//default pointed memory
		vmMem.AllocedMemIdex = -1
		vmMem.PointedMemIndex = len(vmMem.ByteMem) / 2 //the second half memory is reserved for the pointed objects,string,array,structs
	}
	return vmMem, nil
}

func (vm *VMInterpreter) resetGlobals() error {
	for i, global := range vm.module.GlobalIndexSpace {
		val, err := vm.module.ExecInitExpr(global.Init)
		if err != nil {
			return err
		}
		switch v := val.(type) {
		case int32:
			vm.globals[i] = uint64(v)
		case int64:
			vm.globals[i] = uint64(v)
		case float32:
			vm.globals[i] = uint64(math.Float32bits(v))
		case float64:
			vm.globals[i] = uint64(math.Float64bits(v))
		}
	}

	return nil
}

// ByteMem returns the linear Mem space for the VMInterpreter.
func (vm *VMInterpreter) GetMemory() *memory.VMmemory {
	return vm.Mem
}

func (vm *VMInterpreter) pushBool(v bool) {
	if v {
		vm.pushUint64(1)
	} else {
		vm.pushUint64(0)
	}
}

func (vm *VMInterpreter) fetchBool() bool {
	return vm.fetchInt8() != 0
}

func (vm *VMInterpreter) fetchInt8() int8 {
	i := int8(vm.ctx.code[vm.ctx.pc])
	vm.ctx.pc++
	return i
}

func (vm *VMInterpreter) fetchUint32() uint32 {
	v := endianess.Uint32(vm.ctx.code[vm.ctx.pc:])
	vm.ctx.pc += 4
	return v
}

func (vm *VMInterpreter) fetchInt32() int32 {
	return int32(vm.fetchUint32())
}

func (vm *VMInterpreter) fetchFloat32() float32 {
	return math.Float32frombits(vm.fetchUint32())
}

func (vm *VMInterpreter) fetchUint64() uint64 {
	v := endianess.Uint64(vm.ctx.code[vm.ctx.pc:])
	vm.ctx.pc += 8
	return v
}

func (vm *VMInterpreter) fetchInt64() int64 {
	return int64(vm.fetchUint64())
}

func (vm *VMInterpreter) fetchFloat64() float64 {
	return math.Float64frombits(vm.fetchUint64())
}

func (vm *VMInterpreter) popUint64() uint64 {
	i := vm.ctx.stack[len(vm.ctx.stack)-1]
	vm.ctx.stack = vm.ctx.stack[:len(vm.ctx.stack)-1]
	return i
}

func (vm *VMInterpreter) popInt64() int64 {
	return int64(vm.popUint64())
}

func (vm *VMInterpreter) popFloat64() float64 {
	return math.Float64frombits(vm.popUint64())
}

func (vm *VMInterpreter) popUint32() uint32 {
	return uint32(vm.popUint64())
}

func (vm *VMInterpreter) popInt32() int32 {
	return int32(vm.popUint32())
}

func (vm *VMInterpreter) popFloat32() float32 {
	return math.Float32frombits(vm.popUint32())
}

func (vm *VMInterpreter) pushUint64(i uint64) {
	if debugStackDepth {
		if len(vm.ctx.stack) >= cap(vm.ctx.stack) {
			panic("stack exceeding max depth: " + fmt.Sprintf("len=%d,cap=%d", len(vm.ctx.stack), cap(vm.ctx.stack)))
		}
	}
	vm.ctx.stack = append(vm.ctx.stack, i)
}

func (vm *VMInterpreter) pushInt64(i int64) {
	vm.pushUint64(uint64(i))
}

func (vm *VMInterpreter) pushFloat64(f float64) {
	vm.pushUint64(math.Float64bits(f))
}

func (vm *VMInterpreter) pushUint32(i uint32) {
	vm.pushUint64(uint64(i))
}

func (vm *VMInterpreter) pushInt32(i int32) {
	vm.pushUint64(uint64(i))
}

func (vm *VMInterpreter) pushFloat32(f float32) {
	vm.pushUint32(math.Float32bits(f))
}

// ExecCode calls the function with the given index and arguments.
// fnIndex should be a valid index into the function index space of
// the VMInterpreter's module.
func (vm *VMInterpreter) ExecCode(fnIndex int64, args ...uint64) (rtrn interface{}, err error) {
	// If used as a library, client code should set vm.RecoverPanic to true
	// in order to have an error returned.
	if vm.RecoverPanic {
		defer func() {
			if r := recover(); r != nil {
				switch e := r.(type) {
				case error:
					err = e
				default:
					err = fmt.Errorf("exec: %v", e)
				}
			}
		}()
	}
	if int(fnIndex) > len(vm.funcs) {
		return nil, InvalidFunctionIndexError(fnIndex)
	}
	if len(vm.module.GetFunction(int(fnIndex)).Sig.ParamTypes) != len(args) {
		return nil, ErrInvalidArgumentCount
	}
	compiled, ok := vm.funcs[fnIndex].(compiledFunction)
	if !ok {
		panic(fmt.Sprintf("exec: function at index %d is not a compiled function", fnIndex))
	}

	depth := compiled.maxDepth + 1
	if cap(vm.ctx.stack) < depth {
		vm.ctx.stack = make([]uint64, 0, depth)
	} else {
		vm.ctx.stack = vm.ctx.stack[:0]
	}

	vm.ctx.locals = make([]uint64, compiled.totalLocalVars)
	vm.ctx.pc = 0
	vm.ctx.code = compiled.code
	vm.ctx.asm = compiled.asm
	vm.ctx.curFunc = fnIndex

	for i, arg := range args {
		vm.ctx.locals[i] = arg
	}

	res := vm.execCode(compiled)
	if compiled.returns {
		rtrnType := vm.module.GetFunction(int(fnIndex)).Sig.ReturnTypes[0]
		switch rtrnType {
		case wasm.ValueTypeI32:
			rtrn = uint32(res)
		case wasm.ValueTypeI64:
			rtrn = uint64(res)
		case wasm.ValueTypeF32:
			rtrn = math.Float32frombits(uint32(res))
		case wasm.ValueTypeF64:
			rtrn = math.Float64frombits(res)
		default:
			return nil, InvalidReturnTypeError(rtrnType)
		}
	}

	return rtrn, nil
}

func (vm *VMInterpreter) execCode(compiled compiledFunction) uint64 {
outer:
	for int(vm.ctx.pc) < len(vm.ctx.code) && !vm.abort {
		op := vm.ctx.code[vm.ctx.pc]
		vm.ctx.pc++
		switch op {
		case ops.Return:
			break outer
		case compile.OpJmp:
			vm.ctx.pc = vm.fetchInt64()
			continue
		case compile.OpJmpZ:
			target := vm.fetchInt64()
			if vm.popUint32() == 0 {
				vm.ctx.pc = target
				continue
			}
		case compile.OpJmpNz:
			target := vm.fetchInt64()
			preserveTop := vm.fetchBool()
			discard := vm.fetchInt64()
			if vm.popUint32() != 0 {
				vm.ctx.pc = target
				var top uint64
				if preserveTop {
					top = vm.ctx.stack[len(vm.ctx.stack)-1]
				}
				vm.ctx.stack = vm.ctx.stack[:len(vm.ctx.stack)-int(discard)]
				if preserveTop {
					vm.pushUint64(top)
				}
				continue
			}
		case ops.BrTable:
			index := vm.fetchInt64()
			label := vm.popInt32()
			cf, ok := vm.funcs[vm.ctx.curFunc].(compiledFunction)
			if !ok {
				panic(fmt.Sprintf("exec: function at index %d is not a compiled function", vm.ctx.curFunc))
			}
			table := cf.branchTables[index]
			var target compile.Target
			if label >= 0 && label < int32(len(table.Targets)) {
				target = table.Targets[int32(label)]
			} else {
				target = table.DefaultTarget
			}

			if target.Return {
				break outer
			}
			vm.ctx.pc = target.Addr
			var top uint64
			if target.PreserveTop {
				top = vm.ctx.stack[len(vm.ctx.stack)-1]
			}
			vm.ctx.stack = vm.ctx.stack[:len(vm.ctx.stack)-int(target.Discard)]
			if target.PreserveTop {
				vm.pushUint64(top)
			}
			continue
		case compile.OpDiscard:
			place := vm.fetchInt64()
			vm.ctx.stack = vm.ctx.stack[:len(vm.ctx.stack)-int(place)]
		case compile.OpDiscardPreserveTop:
			top := vm.ctx.stack[len(vm.ctx.stack)-1]
			place := vm.fetchInt64()
			vm.ctx.stack = vm.ctx.stack[:len(vm.ctx.stack)-int(place)]
			vm.pushUint64(top)

		case ops.WagonNativeExec:
			i := vm.fetchUint32()
			vm.nativeCodeInvocation(i)
		default:
			vm.funcTable[op]()
		}
	}

	if compiled.returns {
		return vm.ctx.stack[len(vm.ctx.stack)-1]
	}
	return 0
}

// Restart readies the VMInterpreter for another run.
func (vm *VMInterpreter) Restart() {
	vm.resetGlobals()
	vm.ctx.locals = make([]uint64, 0)
	vm.abort = false
}

// Close frees any resources managed by the VMInterpreter.
func (vm *VMInterpreter) Close() error {
	vm.abort = true // prevents further use.
	if vm.nativeBackend != nil {
		if err := vm.nativeBackend.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Process is a proxy passed to host functions in order to access
// things such as Mem and control.
type Process struct {
	vm *VMInterpreter
}

// NewProcess creates a VMInterpreter interface object for host functions
func NewProcess(vm *VMInterpreter) *Process {
	return &Process{vm: vm}
}

// ReadAt implements the ReaderAt interface: it copies into p
// the content of Mem at offset off.
func (proc *Process) ReadAt(p []byte, off int64) (int, error) {
	mem := proc.vm.GetMemory().ByteMem

	var length int
	if len(mem) < len(p)+int(off) {
		length = len(mem) - int(off)
	} else {
		length = len(p)
	}

	copy(p, mem[off:off+int64(length)])

	var err error
	if length < len(p) {
		err = io.ErrShortBuffer
	}

	return length, err
}

// WriteAt implements the WriterAt interface: it writes the content of p
// into the vmMem at offset off.
func (proc *Process) WriteAt(p []byte, off int64) (int, error) {
	mem := proc.vm.GetMemory().ByteMem

	var length int
	if len(mem) < len(p)+int(off) {
		length = len(mem) - int(off)
	} else {
		length = len(p)
	}

	copy(mem[off:], p[:length])

	var err error
	if length < len(p) {
		err = io.ErrShortWrite
	}

	return length, err
}

// Terminate stops the execution of the current module.
func (proc *Process) Terminate() {
	proc.vm.abort = true
}

// GetVMInstance get the current vm instance
func (proc *Process) GetVMInstance() *VMInterpreter {
	return proc.vm
}
