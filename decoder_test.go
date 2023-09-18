package binio_test

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/rand"
	"testing"

	"github.com/KlemensWinter/binio"
	"github.com/stretchr/testify/assert"
)

func TestDecodeInts(t *testing.T) {
	le := binary.LittleEndian
	buf := le.AppendUint16(nil, uint16(math.MaxUint16))
	buf = le.AppendUint32(buf, uint32(math.MaxUint32))
	buf = le.AppendUint64(buf, uint64(math.MaxUint64))

	i16 := int16(math.MinInt16)
	i32 := int32(math.MinInt32)
	i64 := int64(math.MinInt64)
	buf = le.AppendUint16(buf, uint16(i16))
	buf = le.AppendUint32(buf, uint32(i32))
	buf = le.AppendUint64(buf, uint64(i64))

	dec := binio.NewDecoder(bytes.NewReader(buf))
	assert.Equal(t, uint16(math.MaxUint16), dec.Uint16())
	assert.Equal(t, uint32(math.MaxUint32), dec.Uint32())
	assert.Equal(t, uint64(math.MaxUint64), dec.Uint64())

	assert.Equal(t, int16(math.MinInt16), dec.Int16())
	assert.Equal(t, int32(math.MinInt32), dec.Int32())
	assert.Equal(t, int64(math.MinInt64), dec.Int64())
}

func bdec[E any](b *testing.B) {
	buf := make([]E, 1024)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		dec := binio.NewDecoder(NullReader)

		_ = dec.Decode(buf)

	}
}

func BenchmarkDecode(b *testing.B) {
	b.Run("byte", bdec[byte])
	b.Run("int", bdec[int])
	b.Run("float32", bdec[float32])
}

func TestDecodeSkip(t *testing.T) {
	type Foo struct {
		_ [16]uint64
		A uint32
		_ [2]byte
		B uint64
	}
	buf := make([]byte, 128+4+2+8)

	binary.LittleEndian.PutUint32(buf[16*8:], math.MaxUint32)
	binary.LittleEndian.PutUint64(buf[16*8+4+2:], math.MaxUint64)

	var f Foo
	err := binio.Unmarshal(bytes.NewReader(buf), &f)
	assert.Nil(t, err)
	assert.Equal(t, uint32(math.MaxUint32), f.A)
	assert.Equal(t, uint64(math.MaxUint64), f.B)
}

func TestEmbeddedStruct(t *testing.T) {
	tst := struct {
		FieldA bool
		FieldB struct {
			FieldBA struct {
				FieldBAA int64
			}
			FieldBB float64
		}
	}{}

	_ = tst

}

func TestDecodeDynarray(t *testing.T) {
	type TestData struct {
		Int64 []uint64 `bin:"type=dynarray,size=uint32"`
	}

	buf := pack(
		uint32(12),
	)

	rnd := rand.New(rand.NewSource(0))

	const numVals = 12

	var values []uint64

	for i := 0; i < numVals; i++ {
		val := uint64(rnd.Int63())
		values = append(values, val)
		buf = binary.LittleEndian.AppendUint64(buf, val)
	}

	var have TestData

	err := binio.Unmarshal(bytes.NewReader(buf), &have)
	assert.Nil(t, err)
	assert.Len(t, have.Int64, numVals)
	assert.Equal(t, values, have.Int64)
}

func TestDecodeVars(t *testing.T) {
	testdata := struct {
		Foo   byte
		Inner struct {
			Data []byte `bin:"size=$size"`
		} `bin:"$size=%Foo"`
	}{}

	buf := []byte{3, 12, 4, 5}

	err := binio.Unmarshal(bytes.NewReader(buf), &testdata)

	if assert.NoError(t, err) {
		assert.Len(t, testdata.Inner.Data, 3)
	}
}

type TestUnmarshalType struct {
	V string
}

func (v *TestUnmarshalType) UnmarshalDAT(dec *binio.Decoder) error {
	v.V = "It works!"
	return nil
}

var (
	_ binio.Unmarshaler = &TestUnmarshalType{}
)

func TestDecodeUnmarshaler(t *testing.T) {
	var val TestUnmarshalType

	buf := pack(uint16(0), uint16(32))

	err := binio.Unmarshal(bytes.NewReader(buf), &val)
	assert.Nil(t, err)
	assert.Equal(t, "It works!", val.V)
}
