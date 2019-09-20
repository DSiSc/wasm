package memory

import (
	"container/list"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVMmemory_Malloc(t *testing.T) {
	vmMem := &VMmemory{
		ByteMem:   make([]byte, 20),
		MemPoints: make(map[uint64]*TypeLength),
	}
	vmMem.PointedMemIndex = len(vmMem.ByteMem) / 2 //the second half memory is reserved for the pointed objects,string,array,structs
	vmMem.AvailableMem = list.New()
	vmMem.AvailableMem.PushFront(&AvailableMemFragement{
		Start: 1,
		Size:  vmMem.PointedMemIndex - 1,
	})

	ptr, err := vmMem.Malloc(3)
	assert.Nil(t, err)
	assert.Equal(t, 1, ptr)
	assert.Equal(t, 1, vmMem.AvailableMem.Len())
	assert.Equal(t, 4, vmMem.AvailableMem.Front().Value.(*AvailableMemFragement).Start)
	assert.Equal(t, 6, vmMem.AvailableMem.Front().Value.(*AvailableMemFragement).Size)
}

func TestVMmemory_Free(t *testing.T) {
	vmMem := &VMmemory{
		ByteMem:   make([]byte, 20),
		MemPoints: make(map[uint64]*TypeLength),
	}
	vmMem.PointedMemIndex = len(vmMem.ByteMem) / 2 //the second half memory is reserved for the pointed objects,string,array,structs
	vmMem.AvailableMem = list.New()
	vmMem.AvailableMem.PushFront(&AvailableMemFragement{
		Start: 1,
		Size:  vmMem.PointedMemIndex - 1,
	})

	ptr, err := vmMem.Malloc(3)
	assert.Nil(t, err)
	assert.Equal(t, 1, ptr)

	vmMem.Free(ptr)
	assert.Nil(t, vmMem.MemPoints[uint64(ptr)])
	assert.Equal(t, 1, vmMem.AvailableMem.Len())
	assert.Equal(t, 1, vmMem.AvailableMem.Front().Value.(*AvailableMemFragement).Start)
	assert.Equal(t, 9, vmMem.AvailableMem.Front().Value.(*AvailableMemFragement).Size)
}

func TestVMmemory_Free1(t *testing.T) {
	vmMem := &VMmemory{
		ByteMem:   make([]byte, 20),
		MemPoints: make(map[uint64]*TypeLength),
	}
	vmMem.PointedMemIndex = len(vmMem.ByteMem) / 2 //the second half memory is reserved for the pointed objects,string,array,structs
	vmMem.AvailableMem = list.New()
	vmMem.AvailableMem.PushFront(&AvailableMemFragement{
		Start: 1,
		Size:  vmMem.PointedMemIndex - 1,
	})

	ptr1, err := vmMem.Malloc(3)
	assert.Nil(t, err)
	assert.Equal(t, 1, ptr1)

	ptr2, err := vmMem.Malloc(2)
	assert.Nil(t, err)
	assert.Equal(t, 4, ptr2)

	vmMem.Free(ptr1)
	assert.Nil(t, vmMem.MemPoints[uint64(ptr1)])
	assert.Equal(t, 2, vmMem.AvailableMem.Len())

	vmMem.Free(ptr2)
	assert.Nil(t, vmMem.MemPoints[uint64(ptr1)])
	assert.Equal(t, 1, vmMem.AvailableMem.Len())
	assert.Equal(t, 1, vmMem.AvailableMem.Front().Value.(*AvailableMemFragement).Start)
	assert.Equal(t, 9, vmMem.AvailableMem.Front().Value.(*AvailableMemFragement).Size)
}

func TestVMmemory_Free2(t *testing.T) {
	vmMem := &VMmemory{
		ByteMem:   make([]byte, 20),
		MemPoints: make(map[uint64]*TypeLength),
	}
	vmMem.PointedMemIndex = len(vmMem.ByteMem) / 2 //the second half memory is reserved for the pointed objects,string,array,structs
	vmMem.AvailableMem = list.New()
	vmMem.AvailableMem.PushFront(&AvailableMemFragement{
		Start: 1,
		Size:  vmMem.PointedMemIndex - 1,
	})

	ptr1, err := vmMem.Malloc(3)
	assert.Nil(t, err)
	assert.Equal(t, 1, ptr1)

	ptr2, err := vmMem.Malloc(2)
	assert.Nil(t, err)
	assert.Equal(t, 4, ptr2)
	ptr3, err := vmMem.Malloc(2)
	assert.Nil(t, err)
	assert.Equal(t, 6, ptr3)

	vmMem.Free(ptr1)
	assert.Nil(t, vmMem.MemPoints[uint64(ptr1)])
	assert.Equal(t, 2, vmMem.AvailableMem.Len())

	vmMem.Free(ptr3)
	assert.Nil(t, vmMem.MemPoints[uint64(ptr1)])
	assert.Equal(t, 2, vmMem.AvailableMem.Len())

	vmMem.Free(ptr2)
	assert.Nil(t, vmMem.MemPoints[uint64(ptr1)])
	assert.Equal(t, 1, vmMem.AvailableMem.Len())
	assert.Equal(t, 1, vmMem.AvailableMem.Front().Value.(*AvailableMemFragement).Start)
	assert.Equal(t, 9, vmMem.AvailableMem.Front().Value.(*AvailableMemFragement).Size)
}
