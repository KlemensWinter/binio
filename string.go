package binio

import (
	"encoding/binary"
	"io"
)

type SizedSigned interface {
	~int8 | ~int16 | ~int32 | ~int64
}

type SizedUnsigned interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64
}

type SizedInteger interface {
	SizedSigned | SizedUnsigned
}

/*
func AppendVarString[E SizedUnsigned](buf []byte, str string) []byte {
	buf = binary.LittleEndian.AppendUint16()
	return nil
}
*/

func WriteVarString[E SizedUnsigned](w io.Writer, str string) error {
	e := E(len(str))
	err := binary.Write(w, binary.LittleEndian, e)
	if err != nil {
		return err
	}
	if _, err = w.Write([]byte(str)); err != nil {
		return err
	}
	return nil
}
