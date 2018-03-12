// Copyright (c) 2017-2018 Alexander Lokhman. All rights reserved.
// This source code and usage is governed by a MIT style license that can be found in the LICENSE file.

package godat

import (
	"crypto/rand"
	"encoding"
	"encoding/hex"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	TestBool      = true
	TestPtr       = &TestBool
	TestInterface = interface{}(TestBool)

	TestInt         = math.MaxInt64
	TestInt8  int8  = math.MaxInt8
	TestInt16 int16 = math.MaxInt16
	TestInt32 int32 = math.MaxInt32
	TestInt64 int64 = math.MaxInt64

	TestUint   uint   = math.MaxUint64
	TestUint8  uint8  = math.MaxUint8
	TestUint16 uint16 = math.MaxUint16
	TestUint32 uint32 = math.MaxUint32
	TestUint64 uint64 = math.MaxUint64

	TestFloat32 float32 = math.MaxFloat32
	TestFloat64         = math.MaxFloat64

	TestString8  = strings.Repeat("A", math.MaxUint8)
	TestString16 = strings.Repeat("B", math.MaxUint16)
	TestString32 = strings.Repeat("C", math.MaxUint16+1)

	TestBinary8  = make([]byte, math.MaxUint8)
	TestBinary16 = make([]byte, math.MaxUint16)
	TestBinary32 = make([]byte, math.MaxUint16+1)

	TestArray8  [math.MaxUint8]bool
	TestArray16 [math.MaxUint16]*bool
	TestArray32 [math.MaxUint16 + 1]interface{}

	TestMap8  map[int]bool
	TestMap16 map[int]bool
	TestMap32 map[int]bool
)

func init() {
	TestArray8[0] = TestBool
	TestArray16[0] = TestPtr
	TestArray32[0] = TestInterface

	TestMap8 = make(map[int]bool)
	TestMap16 = make(map[int]bool)
	TestMap32 = make(map[int]bool)
	for i := 0; i < math.MaxUint16+1; i++ {
		if i > math.MaxUint16 {
			TestMap32[i] = TestBool
		}
		if i > math.MaxUint8 {
			TestMap16[i] = TestBool
		}
		TestMap8[i] = TestBool
	}
}

type TestInputNil struct {
	A [4]int
	B []int
	C bool
	D int
	E uint
	F float32
	G *int
	H interface{}
	I struct{}
	J chan int
}

type TestInputAnon struct {
	A *bool
}

type TestInputBool struct {
	TestInputAnon
	A bool
}

type TestInputInt struct {
	A int
	B int8
	C int16
	D int32
	E int64
}

type TestInputUint struct {
	A uint
	B uint8
	C uint16
	D uint32
	E uint64
}

type TestInputFloat struct {
	A float32
	B float64
}

type TestInputString struct {
	A string
	B string
	C string
}

type TestInputBinary struct {
	A []byte
	B []byte
	C []byte
}

type TestInputArray struct {
	A [math.MaxUint8]bool
	B [math.MaxUint16]*bool
	C [math.MaxUint16 + 1]interface{}
}

type TestInputObject struct {
	A map[int]bool
	B map[int]bool
	C map[int]bool
}

type TestInputMarshaler struct {
	x string
}

func (v TestInputMarshaler) MarshalBinary() ([]byte, error) {
	return []byte(strings.ToUpper(v.x)), nil
}

func (v *TestInputMarshaler) UnmarshalBinary(data []byte) error {
	v.x = strings.ToLower(string(data))
	return nil
}

var (
	TestInputBrokenMarshalerMarshalError   = errors.New("marshal error")
	TestInputBrokenMarshalerUnmarshalError = errors.New("unmarshal error")
)

type TestInputBrokenMarshaler struct{}

func (v TestInputBrokenMarshaler) MarshalBinary() ([]byte, error) {
	return nil, TestInputBrokenMarshalerMarshalError
}

func (v *TestInputBrokenMarshaler) UnmarshalBinary(data []byte) error {
	return TestInputBrokenMarshalerUnmarshalError
}

type TestInput struct {
	Nil       *TestInputNil
	Bool      *TestInputBool
	Int       *TestInputInt
	Uint      *TestInputUint
	Float     *TestInputFloat
	String    *TestInputString
	Binary    *TestInputBinary
	Array     *TestInputArray
	Object    *TestInputObject
	Marshaler *TestInputMarshaler
}

func NewTestInputBool() *TestInputBool {
	x := &TestInputBool{}
	x.A = TestBool
	x.TestInputAnon.A = TestPtr
	return x
}

func NewTestInputInt() *TestInputInt {
	x := &TestInputInt{}
	x.A = TestInt
	x.B = TestInt8
	x.C = TestInt16
	x.D = TestInt32
	x.E = TestInt64
	return x
}

func NewTestInputUint() *TestInputUint {
	x := &TestInputUint{}
	x.A = TestUint
	x.B = TestUint8
	x.C = TestUint16
	x.D = TestUint32
	x.E = TestUint64
	return x
}

func NewTestInputFloat() *TestInputFloat {
	x := &TestInputFloat{}
	x.A = TestFloat32
	x.B = TestFloat64
	return x
}

func NewTestInputString() *TestInputString {
	x := &TestInputString{}
	x.A = TestString8
	x.B = TestString16
	x.C = TestString32
	return x
}

func NewTestInputBinary() *TestInputBinary {
	x := &TestInputBinary{}
	x.A = TestBinary8
	x.B = TestBinary16
	x.C = TestBinary32
	return x
}

func NewTestInputArray() *TestInputArray {
	x := &TestInputArray{}
	x.A = TestArray8
	x.B = TestArray16
	x.C = TestArray32
	return x
}

func NewTestInputObject() *TestInputObject {
	x := &TestInputObject{}
	x.A = TestMap8
	x.B = TestMap16
	x.C = TestMap32
	return x
}

func NewTestInputMarshaler() *TestInputMarshaler {
	x := &TestInputMarshaler{}
	x.x = "lower_case"
	return x
}

func NewTestInput() *TestInput {
	x := &TestInput{}
	x.Nil = &TestInputNil{}
	x.Bool = NewTestInputBool()
	x.Int = NewTestInputInt()
	x.Uint = NewTestInputUint()
	x.Float = NewTestInputFloat()
	x.String = NewTestInputString()
	x.Binary = NewTestInputBinary()
	x.Array = NewTestInputArray()
	x.Object = NewTestInputObject()
	x.Marshaler = NewTestInputMarshaler()
	return x
}

func TestMarshalNil(t *testing.T) {
	x := &TestInputNil{}
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	y := &TestInputNil{}
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMarshalBool(t *testing.T) {
	x := NewTestInputBool()
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	y := &TestInputBool{}
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalBoolInterfaceError(t *testing.T) {
	x := TestBool
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var ye encoding.BinaryMarshaler
	err = Unmarshal(data, &ye)
	if err == nil {
		t.FailNow()
	}
}

func TestUnmarshalBoolIncompatibleError(t *testing.T) {
	x := TestBool
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var ye string
	err = Unmarshal(data, &ye)
	if err == nil {
		t.FailNow()
	}
}

func TestMarshalInt(t *testing.T) {
	x := NewTestInputInt()
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	y := &TestInputInt{}
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalIntUint64(t *testing.T) {
	x := TestUint64
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y int64
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalIntFloat64(t *testing.T) {
	x := TestFloat64
	data3, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y int64
	err = Unmarshal(data3, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalIntOverflowError(t *testing.T) {
	x := TestFloat64
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var ye int8
	err = Unmarshal(data, &ye)
	if err == nil {
		t.FailNow()
	}
}

func TestUnmarshalIntPtr(t *testing.T) {
	x := TestInt
	data1, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y *int64
	err = Unmarshal(data1, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalIntInterface(t *testing.T) {
	x := TestInt
	data1, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y interface{}
	err = Unmarshal(data1, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalIntInterfaceError(t *testing.T) {
	x := TestInt
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var ye encoding.BinaryMarshaler
	err = Unmarshal(data, &ye)
	if err == nil {
		t.FailNow()
	}
}

func TestUnmarshalIntIncompatibleError(t *testing.T) {
	x := TestInt
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var ye string
	err = Unmarshal(data, &ye)
	if err == nil {
		t.FailNow()
	}
}

func TestMarshalUint(t *testing.T) {
	x := NewTestInputUint()
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	y := &TestInputUint{}
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalUintInt64(t *testing.T) {
	x := TestInt64
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y uint64
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalUintFloat64(t *testing.T) {
	x := TestFloat64
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y uint64
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalUintOverflowError(t *testing.T) {
	x := TestFloat64
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var ye uint8
	err = Unmarshal(data, &ye)
	if err == nil {
		t.FailNow()
	}
}

func TestMarshalFloat(t *testing.T) {
	x := NewTestInputFloat()
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	y := &TestInputFloat{}
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMarshalFloatError(t *testing.T) {
	x := math.NaN()
	_, err := Marshal(x)
	if err == nil {
		t.FailNow()
	}
	_ = err.Error()
}

func TestUnmarshalFloatInt64(t *testing.T) {
	x := TestInt64
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y float64
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalFloatUint64(t *testing.T) {
	x := TestUint64
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y float64
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalFloatOverflowError(t *testing.T) {
	x := TestFloat64
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var ye float32
	err = Unmarshal(data, &ye)
	if err == nil {
		t.FailNow()
	}
}

func TestMarshalString(t *testing.T) {
	x := NewTestInputString()
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	y := &TestInputString{}
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalStringSlice(t *testing.T) {
	x := TestString8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y []byte
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalStringSliceError(t *testing.T) {
	x := TestString8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y []int
	err = Unmarshal(data, &y)
	if err == nil {
		t.FailNow()
	}
}

func TestUnmarshalStringBool(t *testing.T) {
	x := "true"
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y bool
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalStringBoolError(t *testing.T) {
	x := TestString8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y bool
	err = Unmarshal(data, &y)
	if err == nil {
		t.FailNow()
	}
}

func TestUnmarshalStringInt(t *testing.T) {
	x := "-1234567"
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y int
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalStringIntError(t *testing.T) {
	x := TestString8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y int
	err = Unmarshal(data, &y)
	if err == nil {
		t.FailNow()
	}
}

func TestUnmarshalStringUint(t *testing.T) {
	x := "1234567"
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y uint
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalStringUintError(t *testing.T) {
	x := TestString8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y uint
	err = Unmarshal(data, &y)
	if err == nil {
		t.FailNow()
	}
}

func TestUnmarshalStringFloat(t *testing.T) {
	x := "123.4567"
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y float32
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalStringFloatError(t *testing.T) {
	x := TestString8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y float32
	err = Unmarshal(data, &y)
	if err == nil {
		t.FailNow()
	}
}

func TestUnmarshalStringPtr(t *testing.T) {
	x := TestString8
	data1, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y *string
	err = Unmarshal(data1, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalStringInterface(t *testing.T) {
	x := TestString8
	data1, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y interface{}
	err = Unmarshal(data1, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalStringInterfaceError(t *testing.T) {
	x := TestString8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var ye encoding.BinaryMarshaler
	err = Unmarshal(data, &ye)
	if err == nil {
		t.FailNow()
	}
}

func TestUnmarshalStringIncompatibleError(t *testing.T) {
	x := TestString8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var ye struct{}
	err = Unmarshal(data, &ye)
	if err == nil {
		t.FailNow()
	}
}

func TestMarshalBinary(t *testing.T) {
	x := NewTestInputBinary()
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	y := &TestInputBinary{}
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalBinarySliceError(t *testing.T) {
	x := TestBinary8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y []int
	err = Unmarshal(data, &y)
	if err == nil {
		t.FailNow()
	}
}

func TestUnmarshalBinaryInterface(t *testing.T) {
	x := TestBinary8
	data1, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y interface{}
	err = Unmarshal(data1, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalBinaryInterfaceError(t *testing.T) {
	x := TestBinary8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var ye encoding.BinaryMarshaler
	err = Unmarshal(data, &ye)
	if err == nil {
		t.FailNow()
	}
}

func TestUnmarshalBinaryStructError(t *testing.T) {
	x := TestBinary8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var ye struct{}
	err = Unmarshal(data, &ye)
	if err == nil {
		t.FailNow()
	}
}

func TestUnmarshalBinaryIncompatibleError(t *testing.T) {
	x := TestBinary8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var ye int
	err = Unmarshal(data, &ye)
	if err == nil {
		t.FailNow()
	}
}

func TestMarshalArray(t *testing.T) {
	x := NewTestInputArray()
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	y := &TestInputArray{}
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalArrayArraySmall(t *testing.T) {
	x := TestArray8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	y := TestArray16
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalArrayArrayLargeError(t *testing.T) {
	x := TestArray16
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	y := TestArray8
	err = Unmarshal(data, &y)
	if err == nil {
		t.FailNow()
	}
}

func TestUnmarshalArrayArrayIncompatibleError(t *testing.T) {
	x := TestArray8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y [255]int
	err = Unmarshal(data, &y)
	if err == nil {
		t.FailNow()
	}
}

func TestUnmarshalArraySlice(t *testing.T) {
	x := TestArray8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y []bool
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalArraySliceIncompatibleError(t *testing.T) {
	x := TestArray8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y []int
	err = Unmarshal(data, &y)
	if err == nil {
		t.FailNow()
	}
}

func TestUnmarshalArrayPtr(t *testing.T) {
	x := TestArray8
	data1, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y *[255]bool
	err = Unmarshal(data1, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalArrayInterface(t *testing.T) {
	x := TestArray8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y interface{}
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalArrayInterfaceError(t *testing.T) {
	x := TestArray8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var ye encoding.BinaryMarshaler
	err = Unmarshal(data, &ye)
	if err == nil {
		t.FailNow()
	}
}

func TestUnmarshalArrayIncompatibleError(t *testing.T) {
	x := TestArray8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var ye int
	err = Unmarshal(data, &ye)
	if err == nil {
		t.FailNow()
	}
}

func TestMarshalObject(t *testing.T) {
	x := NewTestInputObject()
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	y := &TestInputObject{}
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalObjectMapDirty(t *testing.T) {
	x := TestMap8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	y := TestMap8
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalObjectMapError(t *testing.T) {
	x := TestMap8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y map[int]int
	err = Unmarshal(data, &y)
	if err == nil {
		t.FailNow()
	}
}

func TestUnmarshalObjectStructEmptyError(t *testing.T) {
	x := NewTestInputObject()
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y struct{}
	err = Unmarshal(data, &y)
	if err == nil {
		t.FailNow()
	}
}

func TestUnmarshalObjectStructIncompatibleError(t *testing.T) {
	x := NewTestInputObject()
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y struct{ A int }
	err = Unmarshal(data, &y)
	if err == nil {
		t.FailNow()
	}
}

func TestUnmarshalObjectInterface(t *testing.T) {
	x := TestMap8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y interface{}
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalObjectInterfaceError(t *testing.T) {
	x := TestMap8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var ye encoding.BinaryMarshaler
	err = Unmarshal(data, &ye)
	if err == nil {
		t.FailNow()
	}
}

func TestUnmarshalObjectIncompatibleError(t *testing.T) {
	x := TestMap8
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var ye int
	err = Unmarshal(data, &ye)
	if err == nil {
		t.FailNow()
	}
}

func TestMarshalMarshaler(t *testing.T) {
	x := NewTestInputMarshaler()
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	y := &TestInputMarshaler{}
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMarshalMarshalerError(t *testing.T) {
	x := &TestInputBrokenMarshaler{}
	_, err := Marshal(x)
	if err != TestInputBrokenMarshalerMarshalError {
		t.FailNow()
	}
	_ = err.Error()
}

func TestUnmarshalMarshalerError(t *testing.T) {
	x := NewTestInputMarshaler()
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	y := &TestInputBrokenMarshaler{}
	err = Unmarshal(data, &y)
	if err != TestInputBrokenMarshalerUnmarshalError {
		t.FailNow()
	}
	_ = err.Error()
}

func TestMarshalIncompatible(t *testing.T) {
	var x chan int
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y chan int
	err = Unmarshal(data, &y)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalValueError(t *testing.T) {
	x := TestInt
	data, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	var y1 int
	err = Unmarshal(data, y1)
	if err == nil {
		t.FailNow()
	}

	var y2 *int
	err = Unmarshal(data, y2)
	if err == nil {
		t.FailNow()
	}
}

func TestDump(t *testing.T) {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	n := hex.EncodeToString(randBytes)
	fn := filepath.Join(os.TempDir(), n)
	defer os.Remove(fn)

	x := NewTestInput()
	err := Dump(fn, x)
	if err != nil {
		t.Fatal(err)
	}

	y := &TestInput{}
	err = Load(fn, &y)
	if err != nil {
		t.Fatal(err)
	}
}
