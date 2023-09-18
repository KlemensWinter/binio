package binio

import (
	"reflect"
)

func (dec *Decoder) dynArray(v reflect.Value) error {
	var (
		err  error
		size int
	)

	size = dec.Uint(dec.current().Size)

	if err != nil {
		return err
	}
	return dec.sliceValue(v, size)
}
