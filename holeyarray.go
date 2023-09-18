package binio

import "reflect"

func (dec *Decoder) holeyArray(v reflect.Value) error {
	ptrs := reflect.ValueOf(dec.current().Ptrs)
	if ptrs.Kind() != reflect.Slice {
		panic("need slice here!")
	}

	if ptrs.Len() == 0 {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}
	if ptrs.Len() >= maxArraySize {
		panic("max array")
	}

	sl := reflect.MakeSlice(v.Type(), ptrs.Len(), ptrs.Len())
	for i := 0; i < ptrs.Len(); i++ {
		if ptrs.Index(i).IsZero() {
			continue
		}
		if err := dec.decodeValue(sl.Index(i)); err != nil {
			return err
		}
	}
	v.Set(sl)
	return nil
}
