// Copyright 2018 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package exec_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"reflect"

	"github.com/DSiSc/craft/types"
	"github.com/DSiSc/monkey"
	"github.com/DSiSc/repository"
	"github.com/DSiSc/wasm/exec"
	"github.com/DSiSc/wasm/util"
	"github.com/DSiSc/wasm/wasm"
	"github.com/stretchr/testify/assert"
	"testing"
)

func ExampleVM_add() {
	raw, err := compileWast2Wasm("testdata/add-ex-main.wast")
	if err != nil {
		log.Fatalf("could not compile wast file: %v", err)
	}

	m, err := wasm.ReadModule(bytes.NewReader(raw), func(name string) (*wasm.Module, error) {
		// ReadModule takes as a second argument an optional "importer" function
		// that is supposed to locate and import other modules when some module is
		// requested (by name.)
		// Theoretically, a general "importer" function not unlike the Python's 'import'
		// mechanism (that tries to locate and import modules from a $PYTHONPATH)
		// could be devised.
		switch name {
		case "add":
			raw, err := compileWast2Wasm("testdata/add-ex.wast")
			if err != nil {
				return nil, fmt.Errorf("could not compile wast file hosting %q: %v", name, err)
			}

			add, err := wasm.ReadModule(bytes.NewReader(raw), nil)
			if err != nil {
				return nil, fmt.Errorf("could not read wasm %q module: %v", name, err)
			}
			return add, nil
		case "go":
			// create a whole new module, called "go", from scratch.
			// this module will contain one exported function "print",
			// implemented itself in pure Go.
			print := func(proc *exec.Process, v int32) {
				fmt.Printf("result = %v\n", v)
			}

			m := wasm.NewModule()
			m.Types = &wasm.SectionTypes{
				Entries: []wasm.FunctionSig{
					{
						Form:       0, // value for the 'func' type constructor
						ParamTypes: []wasm.ValueType{wasm.ValueTypeI32},
					},
				},
			}
			m.FunctionIndexSpace = []wasm.Function{
				{
					Sig:  &m.Types.Entries[0],
					Host: reflect.ValueOf(print),
					Body: &wasm.FunctionBody{}, // create a dummy wasm body (the actual value will be taken from Host.)
				},
			}
			m.Export = &wasm.SectionExports{
				Entries: map[string]wasm.ExportEntry{
					"print": {
						FieldStr: "print",
						Kind:     wasm.ExternalFunction,
						Index:    0,
					},
				},
			}

			return m, nil
		}
		return nil, fmt.Errorf("module %q unknown", name)
	})
	if err != nil {
		log.Fatalf("could not read module: %v", err)
	}

	vm, err := exec.NewInterpreter(m)
	if err != nil {
		log.Fatalf("could not create wagon vm: %v", err)
	}

	const fct1 = 2 // index of function fct1
	out, err := vm.ExecCode(fct1)
	if err != nil {
		log.Fatalf("could not execute fct1(): %v", err)
	}
	fmt.Printf("fct1() -> %v\n", out)

	const fct2 = 3 // index of function fct2
	out, err = vm.ExecCode(fct2, 40, 6)
	if err != nil {
		log.Fatalf("could not execute fct2(40, 6): %v", err)
	}
	fmt.Printf("fct2() -> %v\n", out)

	const fct3 = 4 // index of function fct3
	out, err = vm.ExecCode(fct3, 42, 42)
	if err != nil {
		log.Fatalf("could not execute fct3(42, 42): %v", err)
	}
	fmt.Printf("fct3() -> %v\n", out)

	// Output:
	// fct1() -> 42
	// fct2() -> 46
	// result = 84
	// fct3() -> <nil>
}

func TestVM_Malloc(t *testing.T) {
	raw, err := compileWast2Wasm("testdata/malloc.wast")
	if err != nil {
		log.Fatalf("could not compile wast file: %v", err)
	}
	m, err := wasm.ReadModule(bytes.NewReader(raw), exec.NativeResolve)
	if err != nil {
		log.Fatalf("could not read module: %v", err)
	}

	vm, err := exec.NewInterpreter(m)
	if err != nil {
		log.Fatalf("could not create wagon vm: %v", err)
	}

	const fct1 = 1 // index of function fct1
	out, err := vm.ExecCode(fct1)
	assert.Nil(t, err)
	assert.Equal(t, "hello", string(vm.Mem.ByteMem[int(out.(uint32)):int(out.(uint32)+5)]))

	const fct2 = 2 // index of function fct2
	size := 5
	pointer, _ := vm.Mem.Malloc(size)
	copy(vm.Mem.ByteMem[pointer:pointer+size], []byte("dlrow"))
	out, err = vm.ExecCode(fct2, uint64(pointer), uint64(size))
	if err != nil {
		log.Fatalf("could not execute fct2(40, 6): %v", err)
	}
	assert.Nil(t, err)
	assert.Equal(t, "world", string(vm.Mem.ByteMem[pointer:pointer+size]))
}

func TestVM_UpdateState(t *testing.T) {
	raw, err := compileWast2Wasm("testdata/state.wast")
	if err != nil {
		log.Fatalf("could not compile wast file: %v", err)
	}
	m, err := wasm.ReadModule(bytes.NewReader(raw), exec.NativeResolve)
	if err != nil {
		log.Fatalf("could not read module: %v", err)
	}

	vm, err := exec.NewInterpreter(m)
	if err != nil {
		log.Fatalf("could not create wagon vm: %v", err)
	}
	vm.ChainContext = &exec.WasmChainContext{
		Origin: &types.Address{},
	}
	vm.StateDB = &repository.Repository{}

	defer monkey.UnpatchAll()
	db := make(map[string][]byte)
	monkey.PatchInstanceMethod(reflect.TypeOf(vm.StateDB), "SetState", func(self *repository.Repository, address types.Address, key types.Hash, value []byte) {
		db[string(util.HashToBytes(key))] = value
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(vm.StateDB), "GetState", func(self *repository.Repository, address types.Address, key types.Hash) []byte {
		return db[string(util.HashToBytes(key))]
	})

	key, val := []byte("Hello"), []byte("World")
	const fct1 = 2 // index of function fct1
	keyPtr, _ := vm.Mem.Malloc(len(key))
	copy(vm.Mem.ByteMem[keyPtr:keyPtr+len(key)], key)
	valPtr, _ := vm.Mem.Malloc(len(val))
	copy(vm.Mem.ByteMem[valPtr:valPtr+len(val)], val)

	fmt.Println(string(vm.Mem.ByteMem[:10]))
	_, err = vm.ExecCode(fct1, uint64(keyPtr), uint64(len(key)), uint64(valPtr), uint64(len(val)))
	fmt.Println(string(vm.Mem.ByteMem[:10]))
	assert.Nil(t, err)
	assert.Equal(t, val, db[string(util.Hex2Bytes("185f8db32271fe25f561a6fc938b2e264306ec304eda518007d1764826381969"))])

	const fct2 = 3 // index of function fct2
	out, err := vm.ExecCode(fct2, uint64(keyPtr), uint64(len(key)))
	assert.Nil(t, err)
	assert.Equal(t, val, vm.Mem.ByteMem[int(out.(uint32)):int(out.(uint32)+uint32(len(val)))])
}

// compileWast2Wasm fakes a compilation pass from WAST to WASM.
//
// When wagon gets a WAST parser, this function will be running an actual compilation.
// See: https://github.com/DSiSc/wasm/issues/34
func compileWast2Wasm(fname string) ([]byte, error) {
	switch fname {
	case "testdata/add-ex.wast":
		// obtained by running:
		//  $> wat2wasm -v -o add-ex.wasm add-ex.wast
		return ioutil.ReadFile("testdata/add-ex.wasm")
	case "testdata/add-ex-main.wast":
		// obtained by running:
		//  $> wat2wasm -v -o add-ex-main.wasm add-ex-main.wast
		return ioutil.ReadFile("testdata/add-ex-main.wasm")
	case "testdata/malloc.wast":
		return ioutil.ReadFile("testdata/malloc.wasm")
	case "testdata/state.wast":
		return ioutil.ReadFile("testdata/state.wasm")
	}
	return nil, fmt.Errorf("unknown wast test file %q", fname)
}
