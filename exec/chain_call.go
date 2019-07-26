package exec

import (
	"crypto/sha256"
	"github.com/DSiSc/craft/log"
	"github.com/DSiSc/wasm/util"
)

//SetState updates a value in account storage
func SetState(proc *Process, keyStart, keyLen, valStart, valLen int32) {
	vm := proc.GetVMInstance()

	hasher := sha256.New()
	wlen, err := hasher.Write(vm.Mem.ByteMem[keyStart : keyStart+keyLen])
	if err != nil || int32(wlen) < keyLen {
		return
	}
	hash := hasher.Sum(nil)
	vm.StateDB.SetState(*vm.ChainContext.Origin, util.BytesToHash(hash), vm.Mem.ByteMem[valStart:valStart+valLen])
}

//GetState get a value in account storage
func GetState(proc *Process, keyStart, keyLen int32) int32 {
	vm := proc.GetVMInstance()

	hasher := sha256.New()
	wlen, err := hasher.Write(vm.Mem.ByteMem[keyStart : keyStart+keyLen])
	if err != nil || int32(wlen) < keyLen {
		return -1
	}
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
