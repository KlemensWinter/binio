package binio

/**
TODO: support DynArrars with DynStrings
*/

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"

	"github.com/KlemensWinter/binio/expr"
)

var (
	unmarshalerType = reflect.TypeOf((*Unmarshaler)(nil)).Elem()
	decoderFuncs    map[reflect.Type]DecodeFunc
)

type (
	Unmarshaler interface {
		UnmarshalDAT(dec *Decoder) error
	}

	DecodingError struct {
		Pos  int64
		Err  error
		Path []string
	}

	DecodeFunc func(dec *Decoder, v reflect.Value) error

	Decoder struct {
		rd    io.Reader
		pos   int64
		stack []state
	}
)

func (err *DecodingError) Error() string {
	return fmt.Sprintf("decoding error at %v (%d): %s",
		strings.Join(err.Path, "."),
		err.Pos, err.Err)
}

func (err *DecodingError) Unwrap() error {
	return err.Err
}

func (dec *Decoder) Pos() int64 {
	return dec.pos
}

func (dec *Decoder) current() *state {
	if len(dec.stack) == 0 {
		panic("stack is empty")
	}
	return &dec.stack[len(dec.stack)-1]
}

func NewDecoder(rd io.Reader) *Decoder {
	return &Decoder{
		rd:    rd,
		pos:   0,
		stack: make([]state, 0, 100),
	}
}

func (dec *Decoder) Skip(n int64) error {
	l, err := io.CopyN(io.Discard, dec, n)
	if err != nil {
		return err
	}
	if l != n {
		panic("implement me!")
	}
	dec.pos += l
	return nil
}

func (dec *Decoder) Read(p []byte) (n int, err error) {
	n, err = dec.rd.Read(p)
	dec.pos += int64(n)
	return
}

func (dec *Decoder) Uint8() (v uint8)   { return decodePod[uint8](dec) }
func (dec *Decoder) Uint16() (v uint16) { return decodePod[uint16](dec) }
func (dec *Decoder) Uint32() (v uint32) { return decodePod[uint32](dec) }
func (dec *Decoder) Uint64() (v uint64) { return decodePod[uint64](dec) }
func (dec *Decoder) Int8() (v int8)     { return decodePod[int8](dec) }
func (dec *Decoder) Int16() (v int16)   { return decodePod[int16](dec) }
func (dec *Decoder) Int32() (v int32)   { return decodePod[int32](dec) }
func (dec *Decoder) Int64() (v int64)   { return decodePod[int64](dec) }

func (dec *Decoder) getVar(name string) (v any, found bool) {
	if len(dec.stack) == 0 {
		panic("stack empty")
	}
	for i := len(dec.stack); i > 0; i-- {
		v, ok := dec.stack[i-1].Vars[name]
		if !ok {
			continue
		}
		return v, true
	}
	// panic(fmt.Errorf("variable %q not defined", name))
	return nil, false
}

func (dec *Decoder) dynString(v reflect.Value) error {
	size := dec.Uint(dec.current().Size)
	return dec.stringValue(v, size)
}

func (dec *Decoder) Uint(n int) int {
	switch n {
	case 1:
		return int(dec.Uint8())
	case 2:
		return int(dec.Uint16())
	case 4:
		return int(dec.Uint32())
	default:
		panic(fmt.Errorf("uint: size %d not implemented", n))
	}
}

func (dec *Decoder) stringValue(v reflect.Value, size int) error {
	buf := make([]byte, size)
	if _, err := io.ReadAtLeast(dec, buf, size); err != nil {
		return fmt.Errorf("failed to read string: %w", err)
	}

	buf = bytes.TrimRightFunc(buf, func(r rune) bool {
		return r == 0
	})

	v.SetString(string(buf))
	return nil
}

func (dec *Decoder) skip(v reflect.Value) error {
	n, err := ValueSize(v.Type())
	if err != nil {
		return err
	}
	return dec.Skip(int64(n))
}

func (dec *Decoder) structField(strkt, field reflect.Value, fieldIndex int) (err error) {
	if dec.current().Condition != nil {
		cond := expr.Bool(dec.current().Condition)
		if !cond {
			return nil
		}
	}

	t := strkt.Type().Field(fieldIndex)
	if t.Name == "_" { // skipped
		return dec.skip(field)
	}

	switch field.Kind() {
	case reflect.Slice:
		switch {
		case dec.current().Field.Tag.IsDynArray():
			err = dec.dynArray(field)
		case dec.current().Field.Tag.IsHoleyArray():
			err = dec.holeyArray(field)
		default:
			err = dec.sliceValue(field, dec.current().Size)
		}

	case reflect.String:
		if dec.current().Field.Tag.IsDynString() {
			err = dec.dynString(field)
		} else {
			if dec.current().Size == 0 {
				return fmt.Errorf("string with size 0")
			}
			err = dec.stringValue(field, dec.current().Size)
		}
	default:
		err = dec.decodeValue(field)
	}
	return
}

func (dec *Decoder) sliceValue(v reflect.Value, size int) error {
	if size < 0 {
		panic("negative slice size")
	}
	if size == 0 {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}
	if size >= maxArraySize {
		panic(fmt.Errorf("array to big! have=%d, max=%d", size, maxArraySize))
	}
	sl := reflect.MakeSlice(v.Type(), size, size)
	for i := 0; i < size; i++ {
		if err := dec.decodeValue(sl.Index(i)); err != nil {
			return err
		}
	}
	v.Set(sl)
	return nil
}

func (dec *Decoder) beginField() {
	dec.stack = append(dec.stack, state{})
}

func (dec *Decoder) endField() {
	if len(dec.stack) == 0 {
		panic("stack is empty")
	}
	dec.stack = dec.stack[:len(dec.stack)-1]
}

func (dec *Decoder) addErrorContext(err error, name string) error {
	e, ok := err.(*DecodingError)
	if !ok {
		e = &DecodingError{
			Err: err,
			// Pos: dec.Pos(),
			Pos: 0,
		}
	}
	e.Path = append([]string{name}, e.Path...)
	return e
}

func (dec *Decoder) eval(ex expr.Expr, this reflect.Value) (v any, err error) {
	if this.Kind() != reflect.Struct {
		panic("Decoder.eval(): implement me!")
	}

	ctx := &expr.Context{
		GetField: func(name string) (v any, ok bool) {
			field := this.FieldByName(name)
			if !field.IsValid() {
				return nil, false
			}
			switch {
			case field.CanInt():
				v = field.Int()
			case field.CanUint():
				v = int64(field.Uint()) // TODO: check overflow
			default:
				v = field.Interface()
			}
			return v, true
		},
		GetIdent: func(name string) (v any, ok bool) {
			size := IntSize(name)
			if size != -1 {
				return size, true
			}
			return "", false
		},
		GetVar: func(name string) (v any, ok bool) {
			return dec.getVar(name)
		},
	}
	v, err = expr.Eval(ctx, ex)
	if err != nil {
		if errors.Is(err, expr.ErrVarNotDefined) {
			log.Printf("evailable variables:")
			for i, state := range dec.stack {
				log.Printf("%d: %v", i, state.Vars)
			}
		}
		panic(err)
	}
	return v, err
}

func (dec *Decoder) evalField(this reflect.Value, f *field) {
	cur := dec.current()
	cur.Field = f

	if f.Tag == nil {
		// panic(fmt.Errorf("Tag==nil; should not happen for %s.%s", strkt.Type().PkgPath(), strkt.Type().Name()))
		return
	}

	for key, value := range f.Tag.Vars {
		v, err := dec.eval(value, this)
		if err != nil {
			panic(err)
		}
		cur.Set(key, v)
	}

	if f.Tag.Size != nil {
		v, err := dec.eval(f.Tag.Size, this)
		if err != nil {
			panic(err)
		}
		if f.Tag.IsDynArray() || f.Tag.IsDynString() {
			size := v.(int)
			if size == -1 {
				panic(fmt.Errorf("invalid count type %q for dynarray/dynstring", v))
			}
			cur.Size = size
		} else {
			cur.Size = int(v.(int64))
		}
	}

	if f.Tag.Ptrs != nil {
		ptrs, err := dec.eval(f.Tag.Ptrs, this)
		if err != nil {
			panic(err)
		}
		cur.Ptrs = ptrs
	}
	if f.HasCondition() {
		v, err := dec.eval(f.Tag.If, this)
		if err != nil {
			structName := this.Type().PkgPath()
			log.Printf("%s.%s ERROR: %#v", structName, f.Name, err)
			panic(err)
		}
		// log.Printf("field %s: condition=%q result=%#v", f.Name, f.Tag.If, v)
		cur.Condition = v
	}
}

func (dec *Decoder) structValue(v reflect.Value) error {
	typ := v.Type()
	// TODO: cache result
	// log.Printf("ParseStrukt: %s name: %s", typ.PkgPath(), typ.Name())

	def, err := generateStructDef(typ)
	if err != nil {
		panic(err)
	}

	for i := 0; i < typ.NumField(); i++ {
		field := def.Fields[i]

		dec.beginField()
		dec.evalField(v, field)

		if err := dec.structField(v, v.Field(i), i); err != nil {
			err = dec.addErrorContext(err, typ.Field(i).Name)
			return err
		}
		dec.endField()
	}
	return nil
}

func (dec *Decoder) arrayValue(v reflect.Value) error {
	for i := 0; i < v.Len(); i++ {
		if err := dec.decodeValue(v.Index(i)); err != nil {
			return err
		}
	}
	return nil
}

func (dec *Decoder) decodeValue(v reflect.Value) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if er, ok := e.(error); ok {
				err = er
			} else {
				err = fmt.Errorf("error: %s", e)
			}
		}
	}()

	if reflect.PointerTo(v.Type()).Implements(unmarshalerType) {
		m := v.Addr().Interface().(Unmarshaler)
		return m.UnmarshalDAT(dec)
	}

	if fn, found := decoderFuncs[v.Type()]; found {
		return fn(dec, v)
	}

	switch v.Kind() {
	case reflect.Bool:
		err = binary.Read(dec, binary.LittleEndian, v.Addr().Interface())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		err = binary.Read(dec, binary.LittleEndian, v.Addr().Interface())
	case reflect.Float32, reflect.Float64:
		err = binary.Read(dec, binary.LittleEndian, v.Addr().Interface())
	case reflect.Struct:
		err = dec.structValue(v)
	case reflect.Array:
		err = dec.arrayValue(v)
	case reflect.Ptr:
		// init empty ptr
		p := reflect.New(v.Type().Elem())
		err = dec.decodeValue(p.Elem())
		if err == nil {
			v.Set(p)
		}
	default:
		err = fmt.Errorf("decodeValue() invalid type: %s", v.Kind())
	}
	return
}

func (dec *Decoder) DecodeValue(v reflect.Value) error { return dec.decodeValue(v) }

func (dec *Decoder) Decode(v any) (err error) {
	switch val := v.(type) {
	case []byte:
		_, err = io.ReadFull(dec, val)
		return
	}
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("v must be a pointer, got %T", v)

	}
	val = val.Elem()
	return dec.DecodeValue(val)
}

func Unmarshal(rd io.Reader, v any) error {
	dec := NewDecoder(rd)
	return dec.Decode(v)
}
