// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package exec

import (
	"math"
)

func (vm *VMInterpreter) i32Wrapi64() {
	vm.pushUint32(uint32(vm.popUint64()))
}

func (vm *VMInterpreter) i32TruncSF32() {
	vm.pushInt32(int32(math.Trunc(float64(vm.popFloat32()))))
}

func (vm *VMInterpreter) i32TruncUF32() {
	vm.pushUint32(uint32(math.Trunc(float64(vm.popFloat32()))))
}

func (vm *VMInterpreter) i32TruncSF64() {
	vm.pushInt32(int32(math.Trunc(vm.popFloat64())))
}

func (vm *VMInterpreter) i32TruncUF64() {
	vm.pushUint32(uint32(math.Trunc(vm.popFloat64())))
}

func (vm *VMInterpreter) i64ExtendSI32() {
	vm.pushInt64(int64(vm.popInt32()))
}

func (vm *VMInterpreter) i64ExtendUI32() {
	vm.pushUint64(uint64(vm.popUint32()))
}

func (vm *VMInterpreter) i64TruncSF32() {
	vm.pushInt64(int64(math.Trunc(float64(vm.popFloat32()))))
}

func (vm *VMInterpreter) i64TruncUF32() {
	vm.pushUint64(uint64(math.Trunc(float64(vm.popFloat32()))))
}

func (vm *VMInterpreter) i64TruncSF64() {
	vm.pushInt64(int64(math.Trunc(vm.popFloat64())))
}

func (vm *VMInterpreter) i64TruncUF64() {
	vm.pushUint64(uint64(math.Trunc(vm.popFloat64())))
}

func (vm *VMInterpreter) f32ConvertSI32() {
	vm.pushFloat32(float32(vm.popInt32()))
}

func (vm *VMInterpreter) f32ConvertUI32() {
	vm.pushFloat32(float32(vm.popUint32()))
}

func (vm *VMInterpreter) f32ConvertSI64() {
	vm.pushFloat32(float32(vm.popInt64()))
}

func (vm *VMInterpreter) f32ConvertUI64() {
	vm.pushFloat32(float32(vm.popUint64()))
}

func (vm *VMInterpreter) f32DemoteF64() {
	vm.pushFloat32(float32(vm.popFloat64()))
}

func (vm *VMInterpreter) f64ConvertSI32() {
	vm.pushFloat64(float64(vm.popInt32()))
}

func (vm *VMInterpreter) f64ConvertUI32() {
	vm.pushFloat64(float64(vm.popUint32()))
}

func (vm *VMInterpreter) f64ConvertSI64() {
	vm.pushFloat64(float64(vm.popInt64()))
}

func (vm *VMInterpreter) f64ConvertUI64() {
	vm.pushFloat64(float64(vm.popUint64()))
}

func (vm *VMInterpreter) f64PromoteF32() {
	vm.pushFloat64(float64(vm.popFloat32()))
}
