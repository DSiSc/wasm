// Copyright 2017 The go-interpreter Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package exec

import (
	"math"
	"math/bits"
)

// int32 operators

func (vm *VMInterpreter) i32Clz() {
	vm.pushUint64(uint64(bits.LeadingZeros32(vm.popUint32())))
}

func (vm *VMInterpreter) i32Ctz() {
	vm.pushUint64(uint64(bits.TrailingZeros32(vm.popUint32())))
}

func (vm *VMInterpreter) i32Popcnt() {
	vm.pushUint64(uint64(bits.OnesCount32(vm.popUint32())))
}

func (vm *VMInterpreter) i32Add() {
	vm.pushUint32(vm.popUint32() + vm.popUint32())
}

func (vm *VMInterpreter) i32Mul() {
	vm.pushUint32(vm.popUint32() * vm.popUint32())
}

func (vm *VMInterpreter) i32DivS() {
	v2 := vm.popInt32()
	v1 := vm.popInt32()
	vm.pushInt32(v1 / v2)
}

func (vm *VMInterpreter) i32DivU() {
	v2 := vm.popUint32()
	v1 := vm.popUint32()
	vm.pushUint32(v1 / v2)
}

func (vm *VMInterpreter) i32RemS() {
	v2 := vm.popInt32()
	v1 := vm.popInt32()
	vm.pushInt32(v1 % v2)
}

func (vm *VMInterpreter) i32RemU() {
	v2 := vm.popUint32()
	v1 := vm.popUint32()
	vm.pushUint32(v1 % v2)
}

func (vm *VMInterpreter) i32Sub() {
	v2 := vm.popUint32()
	v1 := vm.popUint32()
	vm.pushUint32(v1 - v2)
}

func (vm *VMInterpreter) i32And() {
	vm.pushUint32(vm.popUint32() & vm.popUint32())
}

func (vm *VMInterpreter) i32Or() {
	vm.pushUint32(vm.popUint32() | vm.popUint32())
}

func (vm *VMInterpreter) i32Xor() {
	vm.pushUint32(vm.popUint32() ^ vm.popUint32())
}

func (vm *VMInterpreter) i32Shl() {
	v2 := vm.popUint32()
	v1 := vm.popUint32()
	vm.pushUint32(v1 << v2)
}

func (vm *VMInterpreter) i32ShrU() {
	v2 := vm.popUint32()
	v1 := vm.popUint32()
	vm.pushUint32(v1 >> v2)
}

func (vm *VMInterpreter) i32ShrS() {
	v2 := vm.popUint32()
	v1 := vm.popInt32()
	vm.pushInt32(v1 >> v2)
}

func (vm *VMInterpreter) i32Rotl() {
	v2 := vm.popUint32()
	v1 := vm.popUint32()
	vm.pushUint32(bits.RotateLeft32(v1, int(v2)))
}

func (vm *VMInterpreter) i32Rotr() {
	v2 := vm.popUint32()
	v1 := vm.popUint32()
	vm.pushUint32(bits.RotateLeft32(v1, -int(v2)))
}

func (vm *VMInterpreter) i32LeS() {
	v2 := vm.popInt32()
	v1 := vm.popInt32()
	vm.pushBool(v1 <= v2)
}

func (vm *VMInterpreter) i32LeU() {
	v2 := vm.popUint32()
	v1 := vm.popUint32()
	vm.pushBool(v1 <= v2)
}

func (vm *VMInterpreter) i32LtS() {
	v2 := vm.popInt32()
	v1 := vm.popInt32()
	vm.pushBool(v1 < v2)
}

func (vm *VMInterpreter) i32LtU() {
	v2 := vm.popUint32()
	v1 := vm.popUint32()
	vm.pushBool(v1 < v2)
}

func (vm *VMInterpreter) i32GtS() {
	v2 := vm.popInt32()
	v1 := vm.popInt32()
	vm.pushBool(v1 > v2)
}

func (vm *VMInterpreter) i32GeS() {
	v2 := vm.popInt32()
	v1 := vm.popInt32()
	vm.pushBool(v1 >= v2)
}

func (vm *VMInterpreter) i32GtU() {
	v2 := vm.popUint32()
	v1 := vm.popUint32()
	vm.pushBool(v1 > v2)
}

func (vm *VMInterpreter) i32GeU() {
	v2 := vm.popUint32()
	v1 := vm.popUint32()
	vm.pushBool(v1 >= v2)
}

func (vm *VMInterpreter) i32Eqz() {
	vm.pushBool(vm.popUint32() == 0)
}

func (vm *VMInterpreter) i32Eq() {
	vm.pushBool(vm.popUint32() == vm.popUint32())
}

func (vm *VMInterpreter) i32Ne() {
	vm.pushBool(vm.popUint32() != vm.popUint32())
}

// int64 operators

func (vm *VMInterpreter) i64Clz() {
	vm.pushUint64(uint64(bits.LeadingZeros64(vm.popUint64())))
}

func (vm *VMInterpreter) i64Ctz() {
	vm.pushUint64(uint64(bits.TrailingZeros64(vm.popUint64())))
}

func (vm *VMInterpreter) i64Popcnt() {
	vm.pushUint64(uint64(bits.OnesCount64(vm.popUint64())))
}

func (vm *VMInterpreter) i64Add() {
	vm.pushUint64(vm.popUint64() + vm.popUint64())
}

func (vm *VMInterpreter) i64Sub() {
	v2 := vm.popUint64()
	v1 := vm.popUint64()
	vm.pushUint64(v1 - v2)
}

func (vm *VMInterpreter) i64Mul() {
	vm.pushUint64(vm.popUint64() * vm.popUint64())
}

func (vm *VMInterpreter) i64DivS() {
	v2 := vm.popInt64()
	v1 := vm.popInt64()
	vm.pushInt64(v1 / v2)
}

func (vm *VMInterpreter) i64DivU() {
	v2 := vm.popUint64()
	v1 := vm.popUint64()
	vm.pushUint64(v1 / v2)
}

func (vm *VMInterpreter) i64RemS() {
	v2 := vm.popInt64()
	v1 := vm.popInt64()
	vm.pushInt64(v1 % v2)
}

func (vm *VMInterpreter) i64RemU() {
	v2 := vm.popUint64()
	v1 := vm.popUint64()
	vm.pushUint64(v1 % v2)
}

func (vm *VMInterpreter) i64And() {
	vm.pushUint64(vm.popUint64() & vm.popUint64())
}

func (vm *VMInterpreter) i64Or() {
	vm.pushUint64(vm.popUint64() | vm.popUint64())
}

func (vm *VMInterpreter) i64Xor() {
	vm.pushUint64(vm.popUint64() ^ vm.popUint64())
}

func (vm *VMInterpreter) i64Shl() {
	v2 := vm.popUint64()
	v1 := vm.popUint64()
	vm.pushUint64(v1 << v2)
}

func (vm *VMInterpreter) i64ShrS() {
	v2 := vm.popUint64()
	v1 := vm.popInt64()
	vm.pushInt64(v1 >> v2)
}

func (vm *VMInterpreter) i64ShrU() {
	v2 := vm.popUint64()
	v1 := vm.popUint64()
	vm.pushUint64(v1 >> v2)
}

func (vm *VMInterpreter) i64Rotl() {
	v2 := vm.popInt64()
	v1 := vm.popUint64()
	vm.pushUint64(bits.RotateLeft64(v1, int(v2)))
}

func (vm *VMInterpreter) i64Rotr() {
	v2 := vm.popInt64()
	v1 := vm.popUint64()
	vm.pushUint64(bits.RotateLeft64(v1, -int(v2)))
}

func (vm *VMInterpreter) i64Eq() {
	vm.pushBool(vm.popUint64() == vm.popUint64())
}

func (vm *VMInterpreter) i64Eqz() {
	vm.pushBool(vm.popUint64() == 0)
}

func (vm *VMInterpreter) i64Ne() {
	vm.pushBool(vm.popUint64() != vm.popUint64())
}

func (vm *VMInterpreter) i64LtS() {
	v2 := vm.popInt64()
	v1 := vm.popInt64()
	vm.pushBool(v1 < v2)
}

func (vm *VMInterpreter) i64LtU() {
	v2 := vm.popUint64()
	v1 := vm.popUint64()
	vm.pushBool(v1 < v2)
}

func (vm *VMInterpreter) i64GtS() {
	v2 := vm.popInt64()
	v1 := vm.popInt64()
	vm.pushBool(v1 > v2)
}

func (vm *VMInterpreter) i64GtU() {
	v2 := vm.popUint64()
	v1 := vm.popUint64()
	vm.pushBool(v1 > v2)
}

func (vm *VMInterpreter) i64LeU() {
	v2 := vm.popUint64()
	v1 := vm.popUint64()
	vm.pushBool(v1 <= v2)
}

func (vm *VMInterpreter) i64LeS() {
	v2 := vm.popInt64()
	v1 := vm.popInt64()
	vm.pushBool(v1 <= v2)
}

func (vm *VMInterpreter) i64GeS() {
	v2 := vm.popInt64()
	v1 := vm.popInt64()
	vm.pushBool(v1 >= v2)
}

func (vm *VMInterpreter) i64GeU() {
	v2 := vm.popUint64()
	v1 := vm.popUint64()
	vm.pushBool(v1 >= v2)
}

// float32 operators

func (vm *VMInterpreter) f32Abs() {
	vm.pushFloat32(float32(math.Abs(float64(vm.popFloat32()))))
}

func (vm *VMInterpreter) f32Neg() {
	vm.pushFloat32(-vm.popFloat32())
}

func (vm *VMInterpreter) f32Ceil() {
	vm.pushFloat32(float32(math.Ceil(float64(vm.popFloat32()))))
}

func (vm *VMInterpreter) f32Floor() {
	vm.pushFloat32(float32(math.Floor(float64(vm.popFloat32()))))
}

func (vm *VMInterpreter) f32Trunc() {
	vm.pushFloat32(float32(math.Trunc(float64(vm.popFloat32()))))
}

func (vm *VMInterpreter) f32Nearest() {
	f := vm.popFloat32()
	vm.pushFloat32(float32(int32(f + float32(math.Copysign(0.5, float64(f))))))
}

func (vm *VMInterpreter) f32Sqrt() {
	vm.pushFloat32(float32(math.Sqrt(float64(vm.popFloat32()))))
}

func (vm *VMInterpreter) f32Add() {
	vm.pushFloat32(vm.popFloat32() + vm.popFloat32())
}

func (vm *VMInterpreter) f32Sub() {
	v2 := vm.popFloat32()
	v1 := vm.popFloat32()
	vm.pushFloat32(v1 - v2)
}

func (vm *VMInterpreter) f32Mul() {
	vm.pushFloat32(vm.popFloat32() * vm.popFloat32())
}

func (vm *VMInterpreter) f32Div() {
	v2 := vm.popFloat32()
	v1 := vm.popFloat32()
	vm.pushFloat32(v1 / v2)
}

func (vm *VMInterpreter) f32Min() {
	vm.pushFloat32(float32(math.Min(float64(vm.popFloat32()), float64(vm.popFloat32()))))
}

func (vm *VMInterpreter) f32Max() {
	vm.pushFloat32(float32(math.Max(float64(vm.popFloat32()), float64(vm.popFloat32()))))
}

func (vm *VMInterpreter) f32Copysign() {
	vm.pushFloat32(float32(math.Copysign(float64(vm.popFloat32()), float64(vm.popFloat32()))))
}

func (vm *VMInterpreter) f32Eq() {
	vm.pushBool(vm.popFloat32() == vm.popFloat32())
}

func (vm *VMInterpreter) f32Ne() {
	vm.pushBool(vm.popFloat32() != vm.popFloat32())
}

func (vm *VMInterpreter) f32Lt() {
	v2 := vm.popFloat32()
	v1 := vm.popFloat32()
	vm.pushBool(v1 < v2)
}

func (vm *VMInterpreter) f32Gt() {
	v2 := vm.popFloat32()
	v1 := vm.popFloat32()
	vm.pushBool(v1 > v2)
}

func (vm *VMInterpreter) f32Le() {
	v2 := vm.popFloat32()
	v1 := vm.popFloat32()
	vm.pushBool(v1 <= v2)
}

func (vm *VMInterpreter) f32Ge() {
	v2 := vm.popFloat32()
	v1 := vm.popFloat32()
	vm.pushBool(v1 >= v2)
}

// float64 operators

func (vm *VMInterpreter) f64Abs() {
	vm.pushFloat64(math.Abs(vm.popFloat64()))
}

func (vm *VMInterpreter) f64Neg() {
	vm.pushFloat64(-vm.popFloat64())
}

func (vm *VMInterpreter) f64Ceil() {
	vm.pushFloat64(math.Ceil(vm.popFloat64()))
}

func (vm *VMInterpreter) f64Floor() {
	vm.pushFloat64(math.Floor(vm.popFloat64()))
}

func (vm *VMInterpreter) f64Trunc() {
	vm.pushFloat64(math.Trunc(vm.popFloat64()))
}

func (vm *VMInterpreter) f64Nearest() {
	f := vm.popFloat64()
	vm.pushFloat64(float64(int64(f + math.Copysign(0.5, f))))
}

func (vm *VMInterpreter) f64Sqrt() {
	vm.pushFloat64(math.Sqrt(vm.popFloat64()))
}

func (vm *VMInterpreter) f64Add() {
	vm.pushFloat64(vm.popFloat64() + vm.popFloat64())
}

func (vm *VMInterpreter) f64Sub() {
	v2 := vm.popFloat64()
	v1 := vm.popFloat64()
	vm.pushFloat64(v1 - v2)
}

func (vm *VMInterpreter) f64Mul() {
	vm.pushFloat64(vm.popFloat64() * vm.popFloat64())
}

func (vm *VMInterpreter) f64Div() {
	v2 := vm.popFloat64()
	v1 := vm.popFloat64()
	vm.pushFloat64(v1 / v2)
}

func (vm *VMInterpreter) f64Min() {
	vm.pushFloat64(math.Min(vm.popFloat64(), vm.popFloat64()))
}

func (vm *VMInterpreter) f64Max() {
	vm.pushFloat64(math.Max(vm.popFloat64(), vm.popFloat64()))
}

func (vm *VMInterpreter) f64Copysign() {
	vm.pushFloat64(math.Copysign(vm.popFloat64(), vm.popFloat64()))
}

func (vm *VMInterpreter) f64Eq() {
	vm.pushBool(vm.popFloat64() == vm.popFloat64())
}

func (vm *VMInterpreter) f64Ne() {
	vm.pushBool(vm.popFloat64() != vm.popFloat64())
}

func (vm *VMInterpreter) f64Lt() {
	v2 := vm.popFloat64()
	v1 := vm.popFloat64()
	vm.pushBool(v1 < v2)
}

func (vm *VMInterpreter) f64Gt() {
	v2 := vm.popFloat64()
	v1 := vm.popFloat64()
	vm.pushBool(v1 > v2)
}

func (vm *VMInterpreter) f64Le() {
	v2 := vm.popFloat64()
	v1 := vm.popFloat64()
	vm.pushBool(v1 <= v2)
}

func (vm *VMInterpreter) f64Ge() {
	v2 := vm.popFloat64()
	v1 := vm.popFloat64()
	vm.pushBool(v1 >= v2)
}
