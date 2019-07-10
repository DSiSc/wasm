// Copyright 2018 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package exec

import (
	"encoding/json"
	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/evm-NG/common/math"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/repository"
	"github.com/DSiSc/wasm/exec/memory"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/big"
	"reflect"
	"testing"
)

var (
	smallMemoryVM = &VMInterpreter{Mem: &memory.VMmemory{
		ByteMem: []byte{1, 2, 3}}}
	emptyMemoryVM = &VMInterpreter{Mem: &memory.VMmemory{
		ByteMem: []byte{}}}
	smallMemoryProcess = &Process{vm: smallMemoryVM}
	emptyMemoryProcess = &Process{vm: emptyMemoryVM}
	tooBigABuffer      = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
)

func TestNormalWrite(t *testing.T) {
	vm := &VMInterpreter{Mem: &memory.VMmemory{
		ByteMem: make([]byte, 300)}}
	proc := &Process{vm: vm}
	n, err := proc.WriteAt(tooBigABuffer, 0)
	if err != nil {
		t.Fatalf("Found an error when writing: %v", err)
	}
	if n != len(tooBigABuffer) {
		t.Fatalf("Number of written bytes was %d, should have been %d", n, len(tooBigABuffer))
	}
}

func TestWriteBoundary(t *testing.T) {
	n, err := smallMemoryProcess.WriteAt(tooBigABuffer, 0)
	if err == nil {
		t.Fatal("Should have reported an error and didn't")
	}
	if n != len(smallMemoryVM.Mem.ByteMem) {
		t.Fatalf("Number of written bytes was %d, should have been 0", n)
	}
}

func TestReadBoundary(t *testing.T) {
	buf := make([]byte, 300)
	n, err := smallMemoryProcess.ReadAt(buf, 0)
	if err == nil {
		t.Fatal("Should have reported an error and didn't")
	}
	if n != len(smallMemoryVM.Mem.ByteMem) {
		t.Fatalf("Number of written bytes was %d, should have been 0", n)
	}
}

func TestReadEmpty(t *testing.T) {
	buf := make([]byte, 300)
	n, err := emptyMemoryProcess.ReadAt(buf, 0)
	if err == nil {
		t.Fatal("Should have reported an error and didn't")
	}
	if n != 0 {
		t.Fatalf("Number of written bytes was %d, should have been 0", n)
	}
}

func TestReadOffset(t *testing.T) {
	buf0 := make([]byte, 2)
	n0, err := smallMemoryProcess.ReadAt(buf0, 0)
	if err != nil {
		t.Fatalf("Error reading 1-byte buffer: %v", err)
	}
	if n0 != 2 {
		t.Fatalf("Read %d bytes, expected 2", n0)
	}

	buf1 := make([]byte, 1)
	n1, err := smallMemoryProcess.ReadAt(buf1, 1)
	if err != nil {
		t.Fatalf("Error reading 1-byte buffer: %v", err)
	}
	if n1 != 1 {
		t.Fatalf("Read %d bytes, expected 1.", n0)
	}

	if buf0[1] != buf1[0] {
		t.Fatal("Read two different bytes from what should be the same location")
	}
}

func TestWriteEmpty(t *testing.T) {
	n, err := emptyMemoryProcess.WriteAt(tooBigABuffer, 0)
	if err == nil {
		t.Fatal("Should have reported an error and didn't")
	}
	if n != 0 {
		t.Fatalf("Number of written bytes was %d, should have been 0", n)
	}
}

func TestWriteOffset(t *testing.T) {
	vm := &VMInterpreter{Mem: &memory.VMmemory{
		ByteMem: make([]byte, 300)}}
	proc := &Process{vm: vm}

	n, err := proc.WriteAt(tooBigABuffer, 2)
	if err != nil {
		t.Fatalf("error writing to buffer: %v", err)
	}
	if n != len(tooBigABuffer) {
		t.Fatalf("Number of written bytes was %d, should have been %d", n, len(tooBigABuffer))
	}

	if vm.Mem.ByteMem[0] != 0 || vm.Mem.ByteMem[1] != 0 || vm.Mem.ByteMem[2] != tooBigABuffer[0] {
		t.Fatal("Writing at offset didn't work")
	}
}

func TestNewVM(t *testing.T) {
	context := &WasmChainContext{}
	state := &repository.Repository{}
	vm := NewVM(context, state)
	assert.NotNil(t, vm)
}

func TestVM_Create(t *testing.T) {
	context := &WasmChainContext{}
	state := &repository.Repository{}
	vm := NewVM(context, state)
	assert.NotNil(t, vm)

	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(vm.StateDB), "GetNonce", func(self *repository.Repository, address types.Address) uint64 {
		return 0
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(vm.StateDB), "SetNonce", func(self *repository.Repository, address types.Address, nonce uint64) {
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(vm.StateDB), "GetBalance", func(self *repository.Repository, address types.Address) *big.Int {
		return big.NewInt(100)
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(vm.StateDB), "GetCodeHash", func(self *repository.Repository, address types.Address) types.Hash {
		return types.Hash{}
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(vm.StateDB), "CreateAccount", func(self *repository.Repository, address types.Address) {
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(vm.StateDB), "SetCode", func(self *repository.Repository, address types.Address, code []byte) {
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(vm.StateDB), "SubBalance", func(self *repository.Repository, address types.Address, value *big.Int) {
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(vm.StateDB), "AddBalance", func(self *repository.Repository, address types.Address, value *big.Int) {
	})
	caller := types.Address{}
	code, _ := ioutil.ReadFile("testdata/add-ex.wasm")
	gas := uint64(math.MaxUint64)
	_, caddr, leftOverGas, err := vm.Create(caller, code, uint64(gas), big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, gas, leftOverGas)
	assert.Equal(t, types.Address{0xbd, 0x77, 0x4, 0x16, 0xa3, 0x34, 0x5f, 0x91, 0xe4, 0xb3, 0x45, 0x76, 0xcb, 0x80, 0x4a, 0x57, 0x6f, 0xa4, 0x8e, 0xb1}, caddr)
}

func TestVM_Call(t *testing.T) {
	context := &WasmChainContext{}
	state := &repository.Repository{}
	vm := NewVM(context, state)
	assert.NotNil(t, vm)
	caller := types.Address{}
	contractAddr := types.Address{0xbd, 0x77, 0x4, 0x16, 0xa3, 0x34, 0x5f, 0x91, 0xe4, 0xb3, 0x45, 0x76, 0xcb, 0x80, 0x4a, 0x57, 0x6f, 0xa4, 0x8e, 0xb1}
	code, _ := ioutil.ReadFile("testdata/invoke.wasm")

	defer monkey.UnpatchAll()
	monkey.PatchInstanceMethod(reflect.TypeOf(vm.StateDB), "GetCode", func(self *repository.Repository, address types.Address) []byte {
		return code
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(vm.StateDB), "SubBalance", func(self *repository.Repository, address types.Address, value *big.Int) {
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(vm.StateDB), "AddBalance", func(self *repository.Repository, address types.Address, value *big.Int) {
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(vm.StateDB), "GetBalance", func(self *repository.Repository, address types.Address) *big.Int {
		return big.NewInt(100)
	})

	gas := uint64(math.MaxUint64)
	input, _ := json.Marshal([]string{"method1", "argv1"})
	ret, leftOverGas, err := vm.Call(caller, contractAddr, input, gas, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, gas, leftOverGas)
	assert.Equal(t, "method1", string(ret))
}
