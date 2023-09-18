package binio

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateStructDef(t *testing.T) {
	type Test1 struct {
		Foo bool
		Bar string `bin:"size=10"`
	}

	def, err := generateStructDef(reflect.TypeOf(Test1{}))

	if assert.NoError(t, err) {
		assert.Len(t, def.Fields, 2)

		{
			assert.Equal(t, "Foo", def.Fields[0].Name)
			assert.Nil(t, def.Fields[0].Tag)
		}

		{
			assert.Equal(t, "Bar", def.Fields[1].Name)
			assert.NotNil(t, def.Fields[1].Tag)
		}
	}

}
