package main

import (
	"code.google.com/p/go-uuid/uuid"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/linlexing/dbgo/grade"
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"
	"net/url"
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
		"GRADE_ROOT":  grade.GRADE_ROOT.String(),
		"GRADE_TAG":   grade.GRADE_TAG.String(),
	}
}
func package_uuid() map[string]interface{} {
	return map[string]interface{}{
		"NewRandom": func(call otto.FunctionCall) otto.Value {
			return oftenfun.JSToValue(call.Otto, base64.StdEncoding.EncodeToString(uuid.NewRandom()))
		},
	}
}
func package_url() map[string]interface{} {
	return map[string]interface{}{
		"SetQuery": func(call otto.FunctionCall) otto.Value {
			str := oftenfun.AssertString(call.Argument(0))
			values := oftenfun.AssertObject(call.Argument(1))
			u, err := url.Parse(str)
			if err != nil {
				panic(err)
			}
			q := u.Query()
			for key, value := range values {
				switch tv := value.(type) {
				case string:
					q.Set(key, tv)
				case []interface{}:
					for _, v := range tv {
						q.Add(key, v.(string))
					}
				case []string:
					for _, v := range tv {
						q.Add(key, v)
					}
				default:
					panic(fmt.Errorf("the value %#v not is string or []string", value))
				}
			}
			u.RawQuery = q.Encode()
			return oftenfun.JSToValue(call.Otto, u.String())
		},
	}
}
func package_convert() map[string]interface{} {
	return map[string]interface{}{
		"Str2Bytes": func(call otto.FunctionCall) otto.Value {
			str := oftenfun.AssertString(call.Argument(0))
			rev := []byte(str)
			return oftenfun.JSToValue(call.Otto, rev)
		},
		"Bytes2Str": func(call otto.FunctionCall) otto.Value {
			bys := oftenfun.AssertByteArray(call.Argument(0))
			rev := string(bys)
			return oftenfun.JSToValue(call.Otto, rev)
		},
		"NewBytes": func(call otto.FunctionCall) otto.Value {
			l := oftenfun.AssertInteger(call.Argument(0))
			rev := make([]byte, l)
			return oftenfun.JSToValue(call.Otto, rev)
		},
		"EncodeBase64": func(call otto.FunctionCall) otto.Value {
			bys := oftenfun.AssertByteArray(call.Argument(0))
			return oftenfun.JSToValue(call.Otto, base64.StdEncoding.EncodeToString(bys))
		},
	}
}
func package_sha256() map[string]interface{} {
	return map[string]interface{}{
		"Sum256": func(call otto.FunctionCall) otto.Value {
			bys := oftenfun.AssertByteArray(call.Argument(0))
			arr := sha256.Sum256(bys)
			rev := arr[:]
			return oftenfun.JSToValue(call.Otto, rev)
		},
	}
}
func package_crypto_rand() map[string]interface{} {
	return map[string]interface{}{
		"Read": func(call otto.FunctionCall) otto.Value {
			bys := oftenfun.AssertByteArray(call.Argument(0))
			if _, err := rand.Read(bys); err != nil {
				panic(err)
			}
			return oftenfun.JSToValue(call.Otto, bys)
		},
	}
}
func package_fmt() map[string]interface{} {
	return map[string]interface{}{
		"Print": func(call otto.FunctionCall) otto.Value {
			vs := oftenfun.AssertValue(call.ArgumentList...)
			fmt.Print(vs...)
			return otto.UndefinedValue()
		},
		"Printf": func(call otto.FunctionCall) otto.Value {
			formatstr := oftenfun.AssertString(call.Argument(0))
			vs := oftenfun.AssertValue(call.ArgumentList[1:]...)
			fmt.Printf(formatstr, vs...)
			return otto.UndefinedValue()
		},
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
	first.Set("_go", map[string]interface{}{
		"grade":       package_grade(),
		"fmt":         package_fmt(),
		"sha256":      package_sha256(),
		"url":         package_url(),
		"uuid":        package_uuid(),
		"convert":     package_convert(),
		"crypto_rand": package_crypto_rand(),
	})
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
