package exec

import (
	"bytes"
)

//Alloc memory for base types, return the address in memory
func Malloc(proc *Process, size int32) int32 {
	vm := proc.GetVMInstance()
	pointer, err := vm.GetMemory().Malloc(int(size))
	if err != nil {
		return 0
	} else {
		return int32(pointer)
	}
}

// Copy data int memory, return
func Memcpy(proc *Process, dest, src, size int32) int32 {
	vm := proc.GetVMInstance()
	ret := bytes.Compare(vm.Mem.ByteMem[dest:dest+size], vm.Mem.ByteMem[src:src+size])
	copy(vm.Mem.ByteMem[dest:dest+size], vm.Mem.ByteMem[src:src+size])
	return int32(ret)
}
