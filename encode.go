// Copyright (c) 2017-2018 Alexander Lokhman. All rights reserved.
// This source code and usage is governed by a MIT style license that can be found in the LICENSE file.

package godat

import (
	"encoding"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
)

type EncoderError struct {
	ErrorString string
}

func (e EncoderError) Error() string {
	return fmt.Sprintf("godat: Marshal(%s)", e.ErrorString)
}

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w}
}

func (e *Encoder) write(t byte, v ...interface{}) error {
	if _, err := e.w.Write([]byte{t}); err != nil {
		return err
	}
	for _, vv := range v {
		if err := binary.Write(e.w, binary.BigEndian, vv); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) encodeNil() error {
	return e.write(tNil)
}

func (e *Encoder) encodeBool(v bool) error {
	if v {
		return e.write(tTrue)
	} else {
		return e.write(tFalse)
	}
}

func (e *Encoder) encodeInt(v int64) error {
	if v >= -128 && v <= 127 {
		return e.write(tInt8, int8(v))
	} else if v >= -32768 && v <= 32767 {
		return e.write(tInt16, int16(v))
	} else if v >= -2147483648 && v <= 2147483647 {
		return e.write(tInt32, int32(v))
	} else {
		return e.write(tInt64, v)
	}
}

func (e *Encoder) encodeUint(v uint64) error {
	if v <= 255 {
		return e.write(tUint8, uint8(v))
	} else if v <= 65535 {
		return e.write(tUint16, uint16(v))
	} else if v <= 4294967295 {
		return e.write(tUint32, uint32(v))
	} else {
		return e.write(tUint64, v)
	}
}

func (e *Encoder) encodeFloat(v float64) error {
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return &EncoderError{fmt.Sprintf("unsupported value %s", strconv.FormatFloat(v, 'g', -1, 64))}
	}
	if abs := math.Abs(v); abs >= math.SmallestNonzeroFloat32 && abs <= math.MaxFloat32 {
		return e.write(tFloat32, float32(v))
	} else {
		return e.write(tFloat64, v)
	}
}

func (e *Encoder) encodeString(v string) error {
	if n := len(v); n <= 255 {
		return e.write(tString8, uint8(n), []byte(v))
	} else if n <= 65535 {
		return e.write(tString16, uint16(n), []byte(v))
	} else {
		return e.write(tString32, uint32(n), []byte(v))
	}
}

func (e *Encoder) encodeBinary(v []byte) error {
	if n := len(v); n <= 255 {
		return e.write(tBinary8, uint8(n), []byte(v))
	} else if n <= 65535 {
		return e.write(tBinary16, uint16(n), []byte(v))
	} else {
		return e.write(tBinary32, uint32(n), []byte(v))
	}
}

func (e *Encoder) writeArrayType(n int) error {
	if n <= 255 {
		return e.write(tArray8, uint8(n))
	} else if n <= 65535 {
		return e.write(tArray16, uint16(n))
	} else {
		return e.write(tArray32, uint32(n))
	}
}

func (e *Encoder) encodeArray(v reflect.Value) error {
	n := v.Len()
	if err := e.writeArrayType(n); err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		if err := e.EncodeValue(v.Index(i)); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) writeObjectType(n int) error {
	if n <= 255 {
		return e.write(tObject8, uint8(n))
	} else if n <= 65535 {
		return e.write(tObject16, uint16(n))
	} else {
		return e.write(tObject32, uint32(n))
	}
}

func (e *Encoder) encodeMap(v reflect.Value) error {
	k := v.MapKeys()
	if err := e.writeObjectType(len(k)); err != nil {
		return err
	}
	for _, kk := range k {
		if err := e.EncodeValue(kk); err != nil {
			return err
		}
		if err := e.EncodeValue(v.MapIndex(kk)); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) encodeObject(v reflect.Value) error {
	if vb, ok := v.Interface().(encoding.BinaryMarshaler); ok {
		data, err := vb.MarshalBinary()
		if err != nil {
			return err
		}
		return e.encodeBinary(data)
	}

	x := make(map[string]reflect.Value)
	for i := 0; i < v.NumField(); i++ {
		if f := v.Field(i); !skipValue(f) {
			x[v.Type().Field(i).Name] = f
		}
	}
	if err := e.writeObjectType(len(x)); err != nil {
		return err
	}
	for k, v := range x {
		if err := e.encodeString(k); err != nil {
			return err
		}
		if err := e.EncodeValue(v); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) EncodeValue(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Bool:
		return e.encodeBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return e.encodeInt(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return e.encodeUint(v.Uint())
	case reflect.Float32, reflect.Float64:
		return e.encodeFloat(v.Float())
	case reflect.String:
		return e.encodeString(v.String())
	case reflect.Array, reflect.Slice:
		iv := v.Interface()
		switch iv := iv.(type) {
		case []byte:
			return e.encodeBinary(iv)
		}
		return e.encodeArray(v)
	case reflect.Map:
		return e.encodeMap(v)
	case reflect.Struct:
		return e.encodeObject(v)
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return e.encodeNil()
		}
		return e.EncodeValue(v.Elem())
	}
	return e.encodeNil()
}

func (e *Encoder) Encode(v interface{}) error {
	return e.EncodeValue(reflect.ValueOf(v))
}

func skipValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !skipValue(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !skipValue(v.Field(i)) {
				return false
			}
		}
		return true
	}
	return true
}
