package binio

import (
	"encoding/binary"
	"io"

	"golang.org/x/exp/constraints"
)

type pod interface {
	constraints.Integer | constraints.Float
}

func decodePod[E pod](rd io.Reader) (val E) {
	if err := binary.Read(rd, binary.LittleEndian, &val); err != nil {
		panic(err)
	}
	return
}
