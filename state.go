package binio

type (
	state struct {
		Size int // current slice size

		Field     *field
		Condition any

		Ptrs any // must be a slice; for holeyArray
		Vars map[string]any
	}
)

func (state *state) Set(name string, value any) {
	if state.Vars == nil {
		state.Vars = make(map[string]any)
	}
	state.Vars[name] = value
}

func (state *state) Get(name string) (v any, found bool) {
	if state.Vars == nil {
		return nil, false
	}
	v, found = state.Vars[name]
	return
}
