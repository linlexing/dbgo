package jsmvcerror

import (
	"errors"
	"fmt"
)

var (
	NotFoundControl  = errors.New("the controller not exists")
	FoundControlFile = errors.New("found the file")
	ForbiddenError   = errors.New("Do not have permission to access")

	JSNotIsObject   = errors.New("the param not is object")
	JSNotIsNumber   = errors.New("the param not is number")
	JSNotIsString   = errors.New("the param not is string")
	JSNotIsArray    = errors.New("the param not is array")
	JSNotIsBool     = errors.New("the param not is bool")
	NotFoundProject = errors.New("the project not exists")
)

func CannotUseTable(tname, grade, tgrade string) error {
	return fmt.Errorf("Can't use the table [%s],because user's grade [%s] not is table's grade [%s] uplevel", tname, grade, tgrade)
}

type JavascriptError struct {
	Script string
	Err    error
}

func NewJavascriptError(script string, err error) *JavascriptError {
	return &JavascriptError{
		Script: script,
		Err:    err,
	}
}
func (j *JavascriptError) Error() string {
	return fmt.Sprintf("%s at:\n%s", j.Err.Error(), j.Script)
}
