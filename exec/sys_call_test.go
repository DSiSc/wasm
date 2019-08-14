package exec

import (
	"github.com/DSiSc/wasm/exec/memory"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMalloc(t *testing.T) {
	vm := &VMInterpreter{
		Mem: &memory.VMmemory{
			ByteMem:         make([]byte, 100),
			PointedMemIndex: 50,
			MemPoints:       make(map[uint64]*memory.TypeLength),
		},
	}
	proc := NewProcess(vm)
	pointer1 := Malloc(proc, 2)
	pointer2 := Malloc(proc, 2)
	assert.Equal(t, pointer1+2, pointer2)
}

func TestCalloc(t *testing.T) {
	vm := &VMInterpreter{
		Mem: &memory.VMmemory{
			ByteMem:		make([]byte, 100),
			PointedMemIndex:50,
			MemPoints:		make(map[uint64]*memory.TypeLength),
		},
	}
	proc := NewProcess(vm)
	pointer := Calloc(proc, 10, 4)
	var b byte
	var i int32
	b = 0
	for i = 0; i < 10*4; i++ {
		assert.Equal(t, vm.Mem.ByteMem[pointer+i], b)
	}
}

func TestMemcpy(t *testing.T) {
	vm := &VMInterpreter{
		Mem: &memory.VMmemory{
			ByteMem:		make([]byte, 100),
			PointedMemIndex:50,
			MemPoints:		make(map[uint64]*memory.TypeLength),
		},
	}
	proc := NewProcess(vm)

	pointer := Malloc(proc, 3)
	vm.Mem.ByteMem[pointer] = 'c'
	vm.Mem.ByteMem[pointer+1] = 'h'
	vm.Mem.ByteMem[pointer+2] = 't'
	res := Memcpy(proc, 50, pointer, 3)

	assert.Equal(t, res, int32(-1))
	assert.Equal(t, vm.Mem.ByteMem[50], uint8('c'))
	assert.Equal(t, vm.Mem.ByteMem[50+1], uint8('h'))
	assert.Equal(t, vm.Mem.ByteMem[50+2], uint8('t'))
}

func TestMemset(t *testing.T) {
	vm := &VMInterpreter{
		Mem: &memory.VMmemory{
			ByteMem:		make([]byte, 100),
			PointedMemIndex:50,
			MemPoints:		make(map[uint64]*memory.TypeLength),
		},
	}
	proc := NewProcess(vm)
	pointer := Malloc(proc, 3)
	vm.Mem.ByteMem[pointer] = 'c'
	vm.Mem.ByteMem[pointer+1] = 'h'
	vm.Mem.ByteMem[pointer+2] = 't'

	pointer1 := Memset(proc, pointer, 'h', 10)

	assert.Equal(t, pointer, pointer1)
	for i := 0; i < 10; i++ {
		assert.Equal(t, vm.Mem.ByteMem[int(pointer)+i], uint8('h'))
	}
}