package binio_test

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"testing"

	"github.com/KlemensWinter/go-binio"
	"github.com/stretchr/testify/assert"
)

func TestDecodeCondition_bool(t *testing.T) {

	{
		type Struct struct {
			Cond bool
			Val1 uint16 `bin:"if=!%Cond"`
			Val2 uint64 `bin:"if=!%Cond"`
			End  uint64
		}

		var buf []byte
		buf = append(buf, byte(1))
		buf = le.AppendUint64(buf, math.MaxUint64)

		var res Struct

		err := binio.Unmarshal(bytes.NewReader(buf), &res)
		if assert.NoError(t, err) {
			assert.Equal(t, true, res.Cond)
			assert.Equal(t, uint16(0), res.Val1)
			assert.Equal(t, uint64(0), res.Val2)
			assert.Equal(t, uint64(math.MaxUint64), res.End)
		}
	}

}

func TestDecodeCondition(t *testing.T) {
	type Struct struct {
		FieldA bool
		FieldB int64 `bin:"if=%FieldA"`
		FieldC int64 `bin:"if=%FieldB == 0"`
	}

	le := binary.LittleEndian

	var buf []byte
	buf = append(buf, 1)                              // FieldA
	buf = le.AppendUint64(buf, uint64(math.MaxInt64)) // FieldB
	buf = le.AppendUint64(buf, 123455)                // FieldC should be empty

	var res Struct

	err := binio.Unmarshal(bytes.NewReader(buf), &res)

	if assert.NoError(t, err) {
		assert.True(t, res.FieldA)
		assert.Equal(t, int64(math.MaxInt64), res.FieldB)
		assert.Equal(t, int64(0), res.FieldC)
	}

}

func TestConditionError(t *testing.T) {
	type Struct struct {
		Foo struct {
			Value bool `bin:"if=123"`
		}
	}

	var v Struct

	rd, _ := os.Open(os.DevNull)
	err := binio.Unmarshal(rd, &v)
	assert.Error(t, err)
}
