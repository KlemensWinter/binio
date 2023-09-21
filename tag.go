package binio

import (
	"errors"
	"fmt"
	"strings"

	"github.com/KlemensWinter/go-binio/expr"
	"golang.org/x/exp/maps"
)

var (
	ErrUnknownTagOption = errors.New("unknown tag option")
	ErrInvalidTagOption = errors.New("invalid tag option")
)

var (
	validKeywords = []string{
		"type",
		"size",
		"if",
		"ptrs",
	}

	// these must be lowercase
	strDynArray   = "dynarray"
	strHoleyArray = "holeyarray"
	strDynString  = "dynstring"
)

type (
	Tag struct {
		Size expr.Expr
		If   expr.Expr
		Ptrs expr.Expr

		Vars map[string]expr.Expr

		typ string
	}

	Type byte
)

func (t *Tag) IsDynArray() bool   { return t.typ == strDynArray }
func (t *Tag) IsHoleyArray() bool { return t.typ == strHoleyArray }
func (t *Tag) IsDynString() bool  { return t.typ == strDynString }

func (t *Tag) AddVar(name string, value expr.Expr) {
	if t.Vars == nil {
		t.Vars = make(map[string]expr.Expr)
	}
	t.Vars[name] = value
}

func (t *Tag) HasVar(name string) bool {
	if t.Vars == nil {
		return false
	}
	_, ok := t.Vars[name]
	return ok
}

func (t *Tag) VarNames() []string {
	if t.Vars == nil {
		return nil
	}
	return maps.Keys(t.Vars)
}

func ParseTag(str string) (*Tag, error) {
	var tg Tag
	for _, str := range strings.Split(str, ",") {
		s := strings.SplitN(str, "=", 2)
		key := strings.TrimSpace(s[0])
		value := strings.TrimSpace(s[1])

		switch key {
		case "type":
			switch strings.ToLower(value) {
			case strDynArray, strHoleyArray, strDynString:
				tg.typ = strings.ToLower(value)
			default:
				panic(fmt.Errorf("invalid type %q", value))
			}
		case "size":
			exp, err := expr.Parse(value)
			if err != nil {
				return nil, fmt.Errorf("failed to parse size: %w", err)
			}
			tg.Size = exp
		case "if":
			cond, err := expr.Parse(value)
			if err != nil {
				panic(fmt.Errorf("failed to parse condition %q: %w", value, err))
			}
			tg.If = cond // TODO: syntax check?
		case "ptrs":
			e, err := expr.Parse(value)
			if err != nil {
				return nil, fmt.Errorf("failed to parse ptrs: %w", err)
			}
			tg.Ptrs = e
		default:
			if strings.HasPrefix(key, "$") {
				e, err := expr.Parse(value)
				if err != nil {
					return nil, fmt.Errorf("failed to parse variable %q: %w", key, err)
				}
				tg.AddVar(key[1:], e)
			} else {
				return nil, fmt.Errorf("%w: %q", ErrInvalidTagOption, key)
			}
		}
	}
	return &tg, nil
}
