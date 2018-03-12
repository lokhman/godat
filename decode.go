// Copyright (c) 2017-2018 Alexander Lokhman. All rights reserved.
// This source code and usage is governed by a MIT style license that can be found in the LICENSE file.

package godat

import (
	"encoding"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

type DecoderError struct {
	ErrorString string
}

func (e DecoderError) Error() string {
	return fmt.Sprintf("godat: Unmarshal(%s)", e.ErrorString)
}

type DecoderTypeError struct {
	Value string
	Type  reflect.Type
}

func (e DecoderTypeError) Error() string {
	return fmt.Sprintf("godat: cannot unmarshal %s into Go value of type %s", e.Value, e.Type.String())
}

type Decoder struct {
	r io.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r}
}

func (d *Decoder) read(v ...interface{}) error {
	for _, vv := range v {
		if err := binary.Read(d.r, binary.BigEndian, vv); err != nil {
			return err
		}
	}
	return nil
}

func (d *Decoder) next(n int) ([]byte, error) {
	buf := make([]byte, n)
	if _, err := d.r.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func (d *Decoder) decodeNil(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Interface, reflect.Ptr, reflect.Map, reflect.Slice:
		v.Set(reflect.Zero(v.Type()))
	}
	return nil
}

func (d *Decoder) decodeBool(v reflect.Value, x bool) error {
	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(x)
	case reflect.Interface:
		if v.NumMethod() != 0 {
			return &DecoderTypeError{"bool", v.Type()}
		}
		v.Set(reflect.ValueOf(x))
	case reflect.Ptr:
		return d.decodeBool(indirect(v), x)
	default:
		return &DecoderTypeError{"bool", v.Type()}
	}
	return nil
}

func (d *Decoder) decodeNumber(v reflect.Value, x interface{}, desc string) error {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, _ := x.(int64) // fast no-panic conversion
		switch x := x.(type) {
		case uint64:
			n = int64(x)
		case float64:
			n = int64(x)
		}
		if v.OverflowInt(n) {
			return &DecoderTypeError{
				Value: fmt.Sprintf("%s(%s)", desc, strconv.FormatInt(n, 10)),
				Type:  v.Type(),
			}
		}
		v.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, _ := x.(uint64)
		switch x := x.(type) {
		case int64:
			n = uint64(x)
		case float64:
			n = uint64(x)
		}
		if v.OverflowUint(n) {
			return &DecoderTypeError{
				Value: fmt.Sprintf("%s(%s)", desc, strconv.FormatUint(n, 10)),
				Type:  v.Type(),
			}
		}
		v.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, _ := x.(float64)
		switch x := x.(type) {
		case int64:
			n = float64(x)
		case uint64:
			n = float64(x)
		}
		if v.OverflowFloat(n) {
			return &DecoderTypeError{
				Value: fmt.Sprintf("%s(%s)", desc, strconv.FormatFloat(n, 'f', -1, 64)),
				Type:  v.Type(),
			}
		}
		v.SetFloat(n)
	case reflect.Interface:
		if v.NumMethod() != 0 {
			return &DecoderTypeError{desc, v.Type()}
		}
		v.Set(reflect.ValueOf(x))
	case reflect.Ptr:
		return d.decodeNumber(indirect(v), x, desc)
	default:
		return &DecoderTypeError{desc, v.Type()}
	}
	return nil
}

func (d *Decoder) decodeString(v reflect.Value, n int) error {
	switch v.Kind() {
	case reflect.String:
		data, err := d.next(n)
		if err != nil {
			return err
		}
		v.SetString(string(data))
	case reflect.Slice:
		if v.Type().Elem().Kind() != reflect.Uint8 {
			return &DecoderTypeError{"string", v.Type()}
		}
		data, err := d.next(n)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(data))
	case reflect.Bool:
		data, err := d.next(n)
		if err != nil {
			return err
		}
		n, err := strconv.ParseBool(string(data))
		if err != nil {
			return &DecoderTypeError{"string", v.Type()}
		}
		v.SetBool(n)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		data, err := d.next(n)
		if err != nil {
			return err
		}
		n, err := strconv.ParseInt(string(data), 10, 64)
		if err != nil || v.OverflowInt(n) {
			return &DecoderTypeError{"string", v.Type()}
		}
		v.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		data, err := d.next(n)
		if err != nil {
			return err
		}
		n, err := strconv.ParseUint(string(data), 10, 64)
		if err != nil || v.OverflowUint(n) {
			return &DecoderTypeError{"string", v.Type()}
		}
		v.SetUint(n)
	case reflect.Float32, reflect.Float64:
		data, err := d.next(n)
		if err != nil {
			return err
		}
		n, err := strconv.ParseFloat(string(data), v.Type().Bits())
		if err != nil || v.OverflowFloat(n) {
			return &DecoderTypeError{"string", v.Type()}
		}
		v.SetFloat(n)
	case reflect.Interface:
		if v.NumMethod() != 0 {
			return &DecoderTypeError{"string", v.Type()}
		}
		data, err := d.next(n)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(string(data)))
	case reflect.Ptr:
		return d.decodeString(indirect(v), n)
	default:
		return &DecoderTypeError{"string", v.Type()}
	}
	return nil
}

func (d *Decoder) decodeBinary(v reflect.Value, n int) error {
	switch v.Kind() {
	case reflect.Slice:
		if v.Type().Elem().Kind() != reflect.Uint8 {
			return &DecoderTypeError{"binary", v.Type()}
		}
		data, err := d.next(n)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(data))
	case reflect.Interface:
		if v.NumMethod() != 0 {
			return &DecoderTypeError{"binary", v.Type()}
		}
		data, err := d.next(n)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(data))
	case reflect.Struct:
		if vb, ok := v.Addr().Interface().(encoding.BinaryUnmarshaler); ok {
			data, err := d.next(n)
			if err != nil {
				return err
			}
			return vb.UnmarshalBinary(data)
		}
		return &DecoderTypeError{"binary", v.Type()}
	case reflect.Ptr:
		return d.decodeBinary(indirect(v), n)
	default:
		return &DecoderTypeError{"binary", v.Type()}
	}
	return nil
}

func (d *Decoder) decodeArrayItems(v reflect.Value, n int) error {
	for i := 0; i < n; i++ {
		if err := d.DecodeValue(v.Index(i).Addr()); err != nil {
			return err
		}
	}
	return nil
}

func (d *Decoder) decodeArray(v reflect.Value, n int) error {
	switch v.Kind() {
	case reflect.Array:
		if n > v.Len() {
			return &DecoderTypeError{fmt.Sprintf("array(%d)", n), v.Type()}
		}
		if err := d.decodeArrayItems(v, n); err != nil {
			return err
		}
		if n < v.Len() {
			z := reflect.Zero(v.Type().Elem())
			for i := n; i < v.Len(); i++ {
				v.Index(i).Set(z)
			}
		}
	case reflect.Slice:
		if n > v.Cap() {
			nv := reflect.MakeSlice(v.Type(), v.Len(), n)
			reflect.Copy(nv, v)
			v.Set(nv)
		}
		if n != v.Len() {
			v.SetLen(n)
		}
		if err := d.decodeArrayItems(v, n); err != nil {
			return err
		}
	case reflect.Interface:
		if v.NumMethod() != 0 {
			return &DecoderTypeError{fmt.Sprintf("array(%d)", n), v.Type()}
		}
		xv := reflect.ValueOf(make([]interface{}, n))
		if err := d.decodeArrayItems(xv, n); err != nil {
			return err
		}
		v.Set(xv)
	case reflect.Ptr:
		return d.decodeArray(indirect(v), n)
	default:
		return &DecoderTypeError{fmt.Sprintf("array(%d)", n), v.Type()}
	}
	return nil
}

func (d *Decoder) decodeObjectItems(v reflect.Value, n int) error {
	for i := 0; i < n; i++ {
		vk := reflect.New(v.Type().Key())
		if err := d.DecodeValue(vk); err != nil {
			return err
		}
		vv := reflect.New(v.Type().Elem())
		if err := d.DecodeValue(vv); err != nil {
			return err
		}
		v.SetMapIndex(vk.Elem(), vv.Elem())
	}
	return nil
}

func (d *Decoder) decodeObject(v reflect.Value, n int) error {
	switch v.Kind() {
	case reflect.Map:
		if v.IsNil() {
			v.Set(reflect.MakeMap(v.Type()))
		} else {
			// delete existing items
			zeroValue := reflect.Value{}
			for _, vk := range v.MapKeys() {
				v.SetMapIndex(vk, zeroValue)
			}
		}
		if err := d.decodeObjectItems(v, n); err != nil {
			return err
		}
	case reflect.Struct:
		vn := v.NumField()
		xv := reflect.New(v.Type()).Elem()
		for i := 0; i < n; i++ {
			var xk string
			vk := reflect.ValueOf(&xk)
			if err := d.DecodeValue(vk); err != nil {
				return err
			}
			decoded := false
			for j := 0; j < vn; j++ {
				f := xv.Field(j)
				if xv.Type().Field(j).Name == xk && f.CanSet() {
					if err := d.DecodeValue(f.Addr()); err != nil {
						return err
					}
					decoded = true
				}
			}
			if !decoded {
				return &DecoderTypeError{fmt.Sprintf("object(%d)", n), v.Type()}
			}
		}
		v.Set(xv)
	case reflect.Interface:
		if v.NumMethod() != 0 {
			return &DecoderTypeError{fmt.Sprintf("object(%d)", n), v.Type()}
		}
		xv := reflect.ValueOf(make(map[interface{}]interface{}))
		if err := d.decodeObjectItems(xv, n); err != nil {
			return err
		}
		v.Set(xv)
	case reflect.Ptr:
		return d.decodeObject(indirect(v), n)
	default:
		return &DecoderTypeError{fmt.Sprintf("object(%d)", n), v.Type()}
	}
	return nil
}

func (d *Decoder) DecodeValue(v reflect.Value) error {
	if v.Kind() != reflect.Ptr {
		return &DecoderError{fmt.Sprintf("non-pointer %s", v.Type().String())}
	}
	if v.IsNil() {
		return &DecoderError{fmt.Sprintf("nil %s", v.Type().String())}
	}

	p := make([]byte, 1)
	if _, err := d.r.Read(p); err != nil {
		return err
	}

	v = v.Elem()
	switch p[0] {
	case tNil:
		return d.decodeNil(v)

	case tTrue:
		return d.decodeBool(v, true)
	case tFalse:
		return d.decodeBool(v, false)

	case tInt8:
		var x int8
		if err := d.read(&x); err != nil {
			return err
		}
		return d.decodeNumber(v, int64(x), "int8")
	case tInt16:
		var x int16
		if err := d.read(&x); err != nil {
			return err
		}
		return d.decodeNumber(v, int64(x), "int16")
	case tInt32:
		var x int32
		if err := d.read(&x); err != nil {
			return err
		}
		return d.decodeNumber(v, int64(x), "int32")
	case tInt64:
		var x int64
		if err := d.read(&x); err != nil {
			return err
		}
		return d.decodeNumber(v, x, "int64")

	case tUint8:
		var x uint8
		if err := d.read(&x); err != nil {
			return err
		}
		return d.decodeNumber(v, uint64(x), "uint8")
	case tUint16:
		var x uint16
		if err := d.read(&x); err != nil {
			return err
		}
		return d.decodeNumber(v, uint64(x), "uint16")
	case tUint32:
		var x uint32
		if err := d.read(&x); err != nil {
			return err
		}
		return d.decodeNumber(v, uint64(x), "uint32")
	case tUint64:
		var x uint64
		if err := d.read(&x); err != nil {
			return err
		}
		return d.decodeNumber(v, x, "uint64")

	case tFloat32:
		var x float32
		if err := d.read(&x); err != nil {
			return err
		}
		return d.decodeNumber(v, float64(x), "float32")
	case tFloat64:
		var x float64
		if err := d.read(&x); err != nil {
			return err
		}
		return d.decodeNumber(v, float64(x), "float64")

	case tString8:
		var n uint8
		if err := d.read(&n); err != nil {
			return err
		}
		return d.decodeString(v, int(n))
	case tString16:
		var n uint16
		if err := d.read(&n); err != nil {
			return err
		}
		return d.decodeString(v, int(n))
	case tString32:
		var n uint32
		if err := d.read(&n); err != nil {
			return err
		}
		return d.decodeString(v, int(n))

	case tBinary8:
		var n uint8
		if err := d.read(&n); err != nil {
			return err
		}
		return d.decodeBinary(v, int(n))
	case tBinary16:
		var n uint16
		if err := d.read(&n); err != nil {
			return err
		}
		return d.decodeBinary(v, int(n))
	case tBinary32:
		var n uint32
		if err := d.read(&n); err != nil {
			return err
		}
		return d.decodeBinary(v, int(n))

	case tArray8:
		var n uint8
		if err := d.read(&n); err != nil {
			return err
		}
		return d.decodeArray(v, int(n))
	case tArray16:
		var n uint16
		if err := d.read(&n); err != nil {
			return err
		}
		return d.decodeArray(v, int(n))
	case tArray32:
		var n uint32
		if err := d.read(&n); err != nil {
			return err
		}
		return d.decodeArray(v, int(n))

	case tObject8:
		var n uint8
		if err := d.read(&n); err != nil {
			return err
		}
		return d.decodeObject(v, int(n))
	case tObject16:
		var n uint16
		if err := d.read(&n); err != nil {
			return err
		}
		return d.decodeObject(v, int(n))
	case tObject32:
		var n uint32
		if err := d.read(&n); err != nil {
			return err
		}
		return d.decodeObject(v, int(n))
	}
	return nil
}

func (d *Decoder) Decode(v interface{}) error {
	return d.DecodeValue(reflect.ValueOf(v))
}

func indirect(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Ptr {
		return v
	}

	t := v.Type().Elem()
	if e := v.Elem(); e.IsValid() {
		return e
	}
	v.Set(reflect.New(t))
	return v.Elem()
}
