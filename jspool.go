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

func package_lx() map[string]interface{} {
	return map[string]interface{}{
		"GRADE_ROOT":   grade.GRADE_ROOT,
		"GRADE_TAG":    grade.GRADE_TAG,
		"GradeCanUse":  grade.JSGradeCanUse,
		"TABLE_ADD":    TABLE_ADD,
		"TABLE_EDIT":   TABLE_EDIT,
		"TABLE_DELETE": TABLE_DELETE,
		"TABLE_BROWSE": TABLE_DELETE,
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
	first.Set("lx", package_lx())
	first.Set("fmt", package_fmt())
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
