// Copyright (c) 2017-2018 Alexander Lokhman. All rights reserved.
// This source code and usage is governed by a MIT style license that can be found in the LICENSE file.

package godat

import (
	"bytes"
	"os"
)

const (
	t8 byte = 0x1A * iota
	t16
	t32
	t64
)

const (
	tNil = 'Z' + t8 // 0x5A

	tTrue  = 'T' + t8 // 0x54
	tFalse = 'F' + t8 // 0x46

	tInt8  = 'I' + t8  // 0x49
	tInt16 = 'I' + t16 // 0x63
	tInt32 = 'I' + t32 // 0x7D
	tInt64 = 'I' + t64 // 0x97

	tUint8  = 'U' + t8  // 0x55
	tUint16 = 'U' + t16 // 0x6F
	tUint32 = 'U' + t32 // 0x89
	tUint64 = 'U' + t64 // 0xA3

	_        = 'D' + t8  // 0x44
	_        = 'D' + t16 // 0x5E
	tFloat32 = 'D' + t32 // 0x78
	tFloat64 = 'D' + t64 // 0x92

	tString8  = 'S' + t8  // 0x53
	tString16 = 'S' + t16 // 0x6D
	tString32 = 'S' + t32 // 0x87
	_         = 'S' + t64 // 0xA1

	tArray8  = 'A' + t8  // 0x41
	tArray16 = 'A' + t16 // 0x5B
	tArray32 = 'A' + t32 // 0x75
	_        = 'A' + t64 // 0x8F

	tObject8  = 'O' + t8  // 0x4F
	tObject16 = 'O' + t16 // 0x69
	tObject32 = 'O' + t32 // 0x83
	_         = 'O' + t64 // 0x9D

	tBinary8  = 'B' + t8  // 0x42
	tBinary16 = 'B' + t16 // 0x5C
	tBinary32 = 'B' + t32 // 0x76
	_         = 'B' + t64 // 0x90
)

func encode(enc *Encoder, vv []interface{}) error {
	for _, v := range vv {
		if err := enc.Encode(v); err != nil {
			return err
		}
	}
	return nil
}

func Marshal(v interface{}, vv ...interface{}) ([]byte, error) {
	vv = append([]interface{}{v}, vv...)

	buf := new(bytes.Buffer)
	if err := encode(NewEncoder(buf), vv); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func Dump(filename string, v interface{}, vv ...interface{}) error {
	vv = append([]interface{}{v}, vv...)

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return encode(NewEncoder(f), vv)
}

func decode(dec *Decoder, vv []interface{}) error {
	for _, v := range vv {
		if err := dec.Decode(v); err != nil {
			return err
		}
	}
	return nil
}

func Unmarshal(data []byte, v interface{}, vv ...interface{}) error {
	return decode(NewDecoder(bytes.NewReader(data)), append([]interface{}{v}, vv...))
}

func Load(filename string, v interface{}, vv ...interface{}) error {
	vv = append([]interface{}{v}, vv...)

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return decode(NewDecoder(f), vv)
}
