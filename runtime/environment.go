package runtime

import (
	"github.com/ghosind/gjs/value"
)

type Runtime struct {
	store map[string]value.Value
	outer *Runtime
}

func New() *Runtime {
	s := make(map[string]value.Value)
	return &Runtime{store: s, outer: nil}
}

func NewEnclosedEnvironment(outer *Runtime) *Runtime {
	env := New()
	env.outer = outer
	return env
}

func (e *Runtime) Get(name string) (value.Value, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		return e.outer.Get(name)
	}
	return obj, ok
}

func (e *Runtime) Set(name string, val value.Value) value.Value {
	e.store[name] = val
	return val
}
