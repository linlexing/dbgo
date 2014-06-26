package main

import (
	"github.com/robertkrimen/otto"
	"strings"
)

const (
	GRADE_ROOT = "root" //最顶层
	GRADE_TAG  = ""     //最低层
)

//判断指定的Grade能否使用，规则是本级及以上的可以使用
func GradeCanUse(currentGrade, canUseGrade string) bool {
	return currentGrade == GRADE_TAG || strings.HasPrefix(currentGrade, canUseGrade)
}
func jsGradeCanUse(call otto.FunctionCall) otto.Value {
	v, _ := otto.ToValue(GradeCanUse(call.Argument(0).String(), call.Argument(1).String()))
	return v
}
