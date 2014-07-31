package main

import (
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/robertkrimen/otto"
	"os"
)

type jsfile struct {
	file *os.File
}

func (f *jsfile) Object() map[string]interface{} {
	return map[string]interface{}{
		"Close": f.jsClose,
	}
}
func (f *jsfile) jsWriteString(call otto.FunctionCall) otto.Value {
	str := oftenfun.AssertString(call.Argument(0))
	i, err := f.file.WriteString(str)
	if err != nil {
		panic(err)
	}
	return oftenfun.JSToValue(call.Otto, i)
}

func (f *jsfile) jsWrite(call otto.FunctionCall) otto.Value {
	bys := oftenfun.AssertByteArray(call.Argument(0))
	i, err := f.file.Write(bys)
	if err != nil {
		panic(err)
	}
	return oftenfun.JSToValue(call.Otto, i)
}
func (f *jsfile) jsClose(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, f.file.Close())
}
