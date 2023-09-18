package binio

import "reflect"

const (
	tagName = "bin"
)

var (
	cache = map[reflect.Type]*structDef{}
)

type (

	// a struct field
	field struct {
		Struct *structDef

		Name string
		Typ  reflect.Type

		/*
			Size string
			IsDynArray   bool
			IsHoleyArray bool
			IsDynString  bool
			Ptrs string
			// Condition string
			// Condition *govaluate.EvaluableExpression
			Condition string
		*/
		Tag *Tag
		// Vars map[string]string
	}

	structDef struct {
		Name string

		Fields []*field
	}
)

func (f *field) HasCondition() bool {
	return f.Tag.If != nil
}

func checkField(f *field) error {
	switch f.Typ.Kind() {
	case reflect.String:
		if f.Tag == nil {
			return ErrMissingTag
		}
		if f.Tag.Size == nil {
			return ErrMissingSize
		}
	}
	return nil
}

func generateStructDef(v reflect.Type) (*structDef, error) {
	if d, found := cache[v]; found {
		return d, nil
	}
	if v.Kind() != reflect.Struct {
		panic("not a struct!")
	}
	def := &structDef{
		Name: v.PkgPath() + "." + v.Name(),
	}

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		field := &field{
			Struct: def,
			Name:   f.Name,
			Typ:    f.Type,
		}
		if str, found := f.Tag.Lookup(tagName); found {
			tg, err := ParseTag(str)
			if err != nil {
				panic(err)
			}
			field.Tag = tg
		}

		if err := checkField(field); err != nil {
			return nil, err
		}

		def.Fields = append(def.Fields, field)
	}
	cache[v] = def
	return def, nil
}
