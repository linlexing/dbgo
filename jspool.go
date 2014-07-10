package main

import (
	"fmt"
	"github.com/linlexing/dbgo/grade"
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"
)

const (
	TABLE_ADD = iota
	TABLE_EDIT
	TABLE_DELETE
	TABLE_BROWSE
)

type JSPool struct {
	size int
	pool chan *otto.Otto
}

func JSGradeCanUse(call otto.FunctionCall) otto.Value {
	g1 := oftenfun.AssertString(call.Argument(0))
	g2 := oftenfun.AssertString(call.Argument(1))

	return oftenfun.JSToValue(call.Otto, grade.Grade(g1).CanUse(g2))
}
func package_grade() map[string]interface{} {
	return map[string]interface{}{
		"GradeCanUse": JSGradeCanUse,
	}
}

func package_fmt() map[string]interface{} {
	return map[string]interface{}{
		"Sprint": func(call otto.FunctionCall) otto.Value {
			vs := oftenfun.AssertValue(call.ArgumentList...)
			return oftenfun.JSToValue(call.Otto, fmt.Sprint(vs...))
		},
		"Sprintf": func(call otto.FunctionCall) otto.Value {
			formatstr := oftenfun.AssertString(call.Argument(0))
			vs := oftenfun.AssertValue(call.ArgumentList[1:]...)
			return oftenfun.JSToValue(call.Otto, fmt.Sprintf(formatstr, vs...))
		},
	}
}
func NewJSPool(size int) *JSPool {
	p := JSPool{}
	// Create a buffered channel allowing 'size' senders
	p.pool = make(chan *otto.Otto, size)
	first := otto.New()
	first.Set("_grade", package_grade())
	first.Set("_fmt", package_fmt())
	p.pool <- first
	for x := 1; x < size; x++ {
		p.pool <- first.Copy()
	}
	p.size = size
	return &p
}
func (p *JSPool) Get() *otto.Otto {
	return <-p.pool
}

/**
Return a connection we have used to the pool
*/
func (p *JSPool) Release(runtime *otto.Otto) {
	p.pool <- runtime
}
