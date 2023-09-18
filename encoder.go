package binio

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

type Encoder struct {
	w   io.Writer
	pos int64
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: w,
	}
}

func (enc *Encoder) Write(buf []byte) (n int, err error) {
	n, err = enc.w.Write(buf)
	enc.pos += int64(n)
	return
}

func (enc *Encoder) structValue(v reflect.Value) (err error) {

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		_ = field
	}
	return nil
}

func (enc *Encoder) encodeValue(v reflect.Value) (err error) {
	switch v.Kind() {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return binary.Write(enc, binary.LittleEndian, v.Interface())
	case reflect.Struct:
		return enc.structValue(v)
	default:
		return fmt.Errorf("encode: unhandled type: %s", v.Kind())
	}
}

func (enc *Encoder) EncodeValue(v reflect.Value) error {
	return enc.encodeValue(v)
}

func (enc *Encoder) Encode(v any) error {
	return enc.encodeValue(reflect.ValueOf(v))
}
