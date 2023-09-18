package binio

import (
	"errors"
	"fmt"
	ref "reflect"
)

const (
	maxArraySize = 1_000_000
)

var (
	// ErrSkipField = errors.New("skip field")

	// ErrMissingTag will be returned if a required tag is missing
	// e.g. for a field with the type string we need one
	ErrMissingTag  = errors.New("missing field tag")
	ErrMissingSize = errors.New("missing size")
)

func RegisterDecoder(typ ref.Type, dec DecodeFunc) {
	if decoderFuncs == nil {
		decoderFuncs = make(map[ref.Type]DecodeFunc)
	}
	decoderFuncs[typ] = dec
}

var (
	sizes = map[ref.Kind]int{
		ref.Uint8:  1,
		ref.Uint16: 2,
		ref.Uint32: 4,
		ref.Uint64: 8,
		ref.Int8:   1,
		ref.Int16:  2,
		ref.Int32:  4,
		ref.Int64:  8,
	}

	// var intNames = strings.Fields("int16 int32 int64 uint16 uint32 uint64")

	intNames = map[string]ref.Kind{
		"int8":   ref.Int8,
		"int16":  ref.Int16,
		"int32":  ref.Int32,
		"int64":  ref.Int64,
		"uint8":  ref.Uint8,
		"uint16": ref.Uint16,
		"uint32": ref.Uint32,
		"uint64": ref.Uint64,
	}
)

func ValueSize(v ref.Type) (int, error) {
	if n, found := sizes[v.Kind()]; found {
		return n, nil
	}
	switch v.Kind() {
	case ref.Array:
		n, err := ValueSize(v.Elem())
		if err != nil {
			return 0, fmt.Errorf("unable to get size of array element: %w", err)
		}
		return n * v.Len(), nil
	default:
		return 0, fmt.Errorf("sizeof() unhandled type %s", v.Kind())
	}
}

func IntSize(name string) int {
	if kind, found := intNames[name]; found {
		return sizes[kind]
	}
	return -1
}

/*
// return true if the given type is true
func isTrue(v any) bool {
	if t, ok := v.(bool); ok {
		return t
	}
	val := ref.ValueOf(v)
	return !val.IsZero()
}
*/
