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

//alloc memory bytes number is num*size, difference with Malloc is, calloc will reset the memory as zero
func Calloc(proc *Process, num, size int32) int32{
	vm := proc.GetVMInstance()
	pointer, err := vm.GetMemory().Malloc(int(num*size))
	if err != nil {
		return 0
	}

	//clear the memory bytes
	for i := 0; i < int(num*size); i++ {
		vm.Mem.ByteMem[pointer+i] = 0
	}
	return int32(pointer)
}

//copy the charactor c to the str forward n elements
func Memset(proc *Process, str, c, n int32) int32 {
	vm := proc.GetVMInstance()
	for i := 0; i < int(n); i++ {
		vm.Mem.ByteMem[int(str)+i] = byte(c)
	}

	return str
}
