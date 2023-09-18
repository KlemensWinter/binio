package expr

import (
	"errors"
	"fmt"
	"reflect"

	"golang.org/x/exp/constraints"
)

var (
	ErrVarNotDefined = errors.New("variable not defined")
	ErrIdentNotFound = errors.New("unknown identifier")
	ErrFieldNotFound = errors.New("field not found")
)

func GetFieldFn(strkt reflect.Value) func(string) (any, bool) {
	return func(name string) (any, bool) {
		v := strkt.FieldByName(name)
		if !v.IsValid() {
			return nil, false
		}
		return v.Interface(), true
	}
}

type Context struct {
	GetField func(name string) (any, bool)
	GetVar   func(name string) (any, bool)
	GetIdent func(name string) (any, bool)
}

func cmp[E constraints.Ordered](op Token, x, y E) (res bool, err error) {
	switch op {
	case LSS:
		res = x < y
	case GTR:
		res = x > y
	case EQL:
		res = x == y
	case NEQ:
		res = x != y
	case LEQ:
		res = x <= y
	case GEQ:
		res = x >= y
	default:
		return false, fmt.Errorf("invalid op %s", op)
	}
	return
}

func Bool(val any) bool {
	v, ok := val.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(val)
	}

	switch v.Kind() {
	case reflect.Invalid:
		return false
	case reflect.Bool:
		return v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() > 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr:
		return v.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return v.Float() > 0
	case reflect.String:
		return v.Len() > 0
	case reflect.Slice:
		return v.Len() != 0
	case reflect.Ptr:
		return !v.IsNil()
	}

	return !v.IsZero()
}

func isBool(v reflect.Value) bool { return v.Kind() == reflect.Bool }

func isFloat(k reflect.Value) bool {
	return k.Kind() == reflect.Float32 || k.Kind() == reflect.Float64
}

var (
	typFloat = reflect.TypeOf(float64(0))
)

func Compare(op Token, lhs, rhs any) (res bool, err error) {
	x, ok := lhs.(reflect.Value)
	if !ok {
		x = reflect.ValueOf(lhs)
	}
	y, ok := rhs.(reflect.Value)
	if !ok {
		y = reflect.ValueOf(rhs)
	}

	switch {
	case isBool(x) || isBool(y):
		// if bool, convert types to bool
		a := Bool(x)
		b := Bool(y)
		switch op {
		case LAND:
			return a && b, nil
		case LOR:
			return a || b, nil
		case EQL:
			return a == b, nil
		case NEQ:
			return a != b, nil
		default:
			return false, fmt.Errorf("invalid OP %s for bool", op)
		}
	case isFloat(x) || isFloat(y):
		x = x.Convert(typFloat)
		y = y.Convert(typFloat)
		// convert to float and compare
		res, err = cmp(op, x.Float(), y.Float())

	case x.CanInt() || y.CanInt():
		res, err = cmp(op, x.Int(), y.Int())
	default:
		return false, fmt.Errorf("cant ompare %s with %s", x.Kind(), y.Kind())
	}

	return
}

func eval(ctx *Context, expr Expr) (v any, err error) {
	var ok bool

	switch e := expr.(type) {
	case *Field:
		if ctx.GetField != nil {
			if v, ok = ctx.GetField(e.Name); ok {
				return v, nil
			}
		}
		return nil, fmt.Errorf("%w: %q", ErrFieldNotFound, e.Name)
	case *Var:
		if ctx.GetVar != nil {
			if v, ok = ctx.GetVar(e.Name); ok {
				return v, nil
			}
		}
		return nil, fmt.Errorf("%w: %q", ErrVarNotDefined, e.Name)

	case *Const:
		v = e.Value
	case *Ident:
		if ctx.GetIdent != nil {
			if v, ok = ctx.GetIdent(e.Name); ok {
				return v, nil
			}
		}
		return nil, fmt.Errorf("%w: %q", ErrIdentNotFound, e.Name)
	case *UnaryExpr:
		v, err := eval(ctx, e.X)
		if err != nil {
			return nil, err
		}
		switch e.Op {
		case SUB:
			switch num := v.(type) {
			case int64:
				return -num, nil
			case float64:
				return -num, nil
			default:
				return nil, fmt.Errorf("invalid type for unary op '-': %T", num)
			}
		case NOT:
			res := Bool(v)
			return !res, nil
		default:
			return nil, fmt.Errorf("unaryexpr %s not supported", e.Op)
		}
	case *BinExpr:
		lhs, er := eval(ctx, e.Lhs)
		if er != nil {
			return nil, er
		}
		rhs, er := eval(ctx, e.Rhs)
		if er != nil {
			return nil, er
		}
		v, err = Compare(e.Op, lhs, rhs)
	default:
		panic(fmt.Sprintf("expr.eval(): implement me for type: %T", e))
	}
	return
}

func Eval(ctx *Context, expr Expr) (any, error) {
	return eval(ctx, expr)
}
