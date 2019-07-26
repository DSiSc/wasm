package exec

import (
	"github.com/DSiSc/wasm/exec/memory"
	"github.com/magiconair/properties/assert"
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
