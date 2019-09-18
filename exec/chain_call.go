package exec

import (
	"crypto/sha256"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/wasm/util"
	"math/big"
)

//SetState updates a value in account storage
func SetState(proc *Process, keyPtr, valPtr int32) {
	vm := proc.GetVMInstance()

	keyLen := vm.Mem.MemPoints[uint64(keyPtr)].Length
	hasher := sha256.New()
	hasher.Write(vm.Mem.ByteMem[keyPtr : keyPtr+int32(keyLen)])
	hash := hasher.Sum(nil)

	valLen := vm.Mem.MemPoints[uint64(valPtr)].Length
	vm.StateDB.SetState(*vm.ChainContext.Origin, util.BytesToHash(hash), vm.Mem.ByteMem[valPtr:valPtr+int32(valLen)])
}

//GetState get a value in account storage
func GetState(proc *Process, keyPtr int32) int32 {
	vm := proc.GetVMInstance()

	keyLen := vm.Mem.MemPoints[uint64(keyPtr)].Length
	hasher := sha256.New()
	hasher.Write(vm.Mem.ByteMem[keyPtr : keyPtr+int32(keyLen)])
	hash := hasher.Sum(nil)
	val := vm.StateDB.GetState(*vm.ChainContext.Origin, util.BytesToHash(hash))
	pointer, err := vm.Mem.Malloc(len(val))
	if err != nil {
		log.Error("failed to malloc memory, as:%v", err)
		return -1
	}
	copy(vm.Mem.ByteMem[pointer:pointer+len(val)], val)
	return int32(pointer)
}

//BlockHeight get current block height
func BlockHeight(proc *Process) int64 {
	vm := proc.GetVMInstance()
	return vm.ChainContext.BlockNumber.Int64()
}

//BlockTimeStamp get current block's timestamp
func BlockTimeStamp(proc *Process) int64 {
	vm := proc.GetVMInstance()
	return vm.ChainContext.Time.Int64()
}

//SelfAddress return contract address
func SelfAddress(proc *Process) int32 {
	vm := proc.GetVMInstance()
	return writeContentToMemory(vm, vm.ChainContext.ContractAddr[:])
}

//Sha256 compute the content's sha256 hash
func Sha256(proc *Process, contentPtr int32) int32 {
	vm := proc.GetVMInstance()

	val := getContentFromMemory(vm, contentPtr)
	hasher := sha256.New()
	wlen, err := hasher.Write(val)
	if err != nil || wlen < len(val) {
		return -1
	}
	hash := hasher.Sum(nil)

	return writeContentToMemory(vm, hash)
}

//Call call another contract
func Call(proc *Process, contractAddrPtr int32, paramPtr int32, value int64) int32 {
	vmInterpreter := proc.GetVMInstance()
	return callWithCaller(vmInterpreter, vmInterpreter.ChainContext.Caller, contractAddrPtr, paramPtr, value)
}

//Call call another contract
func StaticCall(proc *Process, contractAddrPtr int32, paramPtr int32, value int64) int32 {
	vmInterpreter := proc.GetVMInstance()
	return callWithCaller(vmInterpreter, vmInterpreter.ChainContext.ContractAddr, contractAddrPtr, paramPtr, value)
}

func callWithCaller(vmInterpreter *VMInterpreter, caller types.Address, contractAddrPtr int32, paramPtr int32, value int64) int32 {
	contractAddrByte := getContentFromMemory(vmInterpreter, contractAddrPtr)
	contractAddr := util.BytesToAddress(contractAddrByte)
	param := getContentFromMemory(vmInterpreter, paramPtr)

	chainContex := &WasmChainContext{
		Origin:       &caller,
		GasPrice:     vmInterpreter.ChainContext.GasPrice,
		Coinbase:     vmInterpreter.ChainContext.Coinbase,
		GasLimit:     vmInterpreter.ChainContext.GasLimit - vmInterpreter.UsedGas,
		BlockNumber:  vmInterpreter.ChainContext.BlockNumber,
		Time:         vmInterpreter.ChainContext.Time,
		Caller:       vmInterpreter.ChainContext.Caller,
		ContractAddr: vmInterpreter.ChainContext.ContractAddr,
	}
	vm := NewVM(chainContex, vmInterpreter.StateDB)
	ret, leftGas, err := vm.Call(caller, contractAddr, param, 0, big.NewInt(value))
	vmInterpreter.UsedGas += chainContex.GasLimit - leftGas
	if err != nil {
		return -1
	}
	return writeContentToMemory(vmInterpreter, ret)
}

func getContentFromMemory(vm *VMInterpreter, pointer int32) []byte {
	cLen := vm.Mem.MemPoints[uint64(pointer)].Length
	content := vm.Mem.ByteMem[pointer : pointer+int32(cLen)]
	return content
}

func writeContentToMemory(vm *VMInterpreter, val []byte) int32 {
	pointer, err := vm.Mem.Malloc(len(val))
	if err != nil {
		return -1
	}
	copy(vm.Mem.ByteMem[pointer:pointer+len(val)], val)
	return int32(pointer)
}
