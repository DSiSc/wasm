// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package exec

func (vm *VMInterpreter) i32Const() {
	vm.pushUint32(vm.fetchUint32())
}

func (vm *VMInterpreter) i64Const() {
	vm.pushUint64(vm.fetchUint64())
}

func (vm *VMInterpreter) f32Const() {
	vm.pushFloat32(vm.fetchFloat32())
}

func (vm *VMInterpreter) f64Const() {
	vm.pushFloat64(vm.fetchFloat64())
}
