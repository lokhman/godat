// Copyright (c) 2017-2018 Alexander Lokhman. All rights reserved.
// This source code and usage is governed by a MIT style license that can be found in the LICENSE file.

package godat

import (
	"bytes"
	"os"
)

const (
	tNil = 90 // 'Z'

	tTrue  = 84 // 'T'
	tFalse = 70 // 'F'

	tInt8  = 73  // 'I'
	tInt16 = 99  // 'I(+0x1A)'
	tInt32 = 125 // 'I(+0x34)'
	tInt64 = 151 // 'I(+0x4E)'

	tUint8  = 85  // 'U'
	tUint16 = 111 // 'U(+0x1A)'
	tUint32 = 137 // 'U(+0x34)'
	tUint64 = 163 // 'U(+0x4E)'

	tFloat32 = 68 // 'D'
	tFloat64 = 94 // 'D(+0x1A)'

	tString8  = 83  // 'S'
	tString16 = 109 // 'S(+0x1A)'
	tString32 = 135 // 'S(+0x34)'

	tArray8  = 65  // 'A'
	tArray16 = 91  // 'A(+0x1A)'
	tArray32 = 117 // 'A(+0x34)'

	tObject8  = 79  // 'O'
	tObject16 = 105 // 'O(+0x1A)'
	tObject32 = 131 // 'O(+0x34)'

	tBinary8  = 66  // 'B'
	tBinary16 = 92  // 'B(+0x1A)'
	tBinary32 = 118 // 'B(+0x34)'
)

func Marshal(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := NewEncoder(buf).Encode(v)
	return buf.Bytes(), err
}

func Dump(filename string, v interface{}) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return NewEncoder(f).Encode(v)
}

func Unmarshal(data []byte, v interface{}) error {
	return NewDecoder(bytes.NewReader(data)).Decode(v)
}

func Load(filename string, v interface{}) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return NewDecoder(f).Decode(v)
}
