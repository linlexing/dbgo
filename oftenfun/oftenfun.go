// oftenfun project oftenfun.go
package oftenfun

import (
	"fmt"
	"github.com/linlexing/dbgo/jsmvcerror"
	"github.com/robertkrimen/otto"
	"os"
	"runtime/debug"
	"strconv"
)

func SafeToStrings(s []interface{}) []string {
	r := make([]string, len(s), len(s))
	for i, v := range s {
		r[i] = SafeToString(v)
	}
	return r
}
func SafeToInt64s(s []interface{}) []int64 {
	r := make([]int64, len(s), len(s))
	for i, v := range s {
		r[i] = SafeToInt64(v)
	}
	return r
}
func SafeToFloat64s(s []interface{}) []float64 {
	r := make([]float64, len(s), len(s))
	for i, v := range s {
		r[i] = SafeToFloat64(v)
	}
	return r
}
func SafeToBools(s []interface{}) []bool {
	r := make([]bool, len(s), len(s))
	for i, v := range s {
		r[i] = SafeToBool(v)
	}
	return r
}
func SafeToString(s interface{}) string {

	if s == nil {
		return ""
	}
	switch r := s.(type) {
	case string:
		return r
	case []byte:
		return string(r)
	default:
		return fmt.Sprintf("%v", s)
	}

}
func SafeToInt64(s interface{}) int64 {
	if s == nil {
		return int64(0)
	}
	switch r := s.(type) {
	case int64:
		return r
	case int:
		return int64(r)
	case int8:
		return int64(r)
	case int16:
		return int64(r)
	case string:
		if i, err := strconv.ParseInt(r, 10, 64); err != nil {
			return int64(0)
		} else {
			return i
		}
	default:
		return 0
	}
}
func SafeToFloat64(s interface{}) float64 {
	if s == nil {
		return float64(0)
	}
	switch r := s.(type) {
	case int64:
		return float64(r)
	case int:
		return float64(r)
	case int8:
		return float64(r)
	case int16:
		return float64(r)
	case float32:
		return float64(r)
	case float64:
		return r
	case string:
		if i, err := strconv.ParseFloat(r, 64); err != nil {
			return float64(0)
		} else {
			return i
		}
	default:
		return float64(0)
	}
}
func SafeToBool(s interface{}) bool {
	if s == nil {
		return false
	}
	switch r := s.(type) {
	case bool:
		return r
	case int:
		return r != 0
	case int8:
		return r != 0
	case int16:
		return r != 0
	case int64:
		return r != 0
	case string:
		result, err := strconv.ParseBool(r)
		if err != nil {
			return false
		} else {
			return result
		}
	case float32:
		return r != 0
	case float64:
		return r != 0
	default:
		return SafeToBool(SafeToString(r))
	}
}

func SafeToBytes(s interface{}) []byte {
	if s == nil {
		return []byte{}
	}
	if b, ok := s.([]byte); ok {
		return b
	}
	if r, ok := s.(string); ok {
		return []byte(r)
	}
	return []byte(fmt.Sprintf("%v", s))
}
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
func AssertObject(v otto.Value) map[string]interface{} {
	if !v.IsObject() {
		panic(jsmvcerror.JSNotIsObject)
	}
	nv, err := v.Export()
	if err != nil {
		panic(err)
	}
	return nv.(map[string]interface{})
}
func AssertInteger(v interface{}) int {
	switch t := v.(type) {
	case otto.Value:
		if !t.IsNumber() {
			panic(jsmvcerror.JSNotIsNumber)
		}
		nv, err := t.ToInteger()
		if err != nil {
			panic(err)
		}
		return int(nv)
	case int64:
		return int(t)
	default:
		return v.(int)
	}
}
func AssertFloat64(v otto.Value) float64 {
	if !v.IsNumber() {
		panic(jsmvcerror.JSNotIsNumber)
	}
	nv, err := v.ToFloat()
	if err != nil {
		panic(err)
	}
	return nv
}
func AssertBool(v interface{}) bool {
	switch t := v.(type) {
	case otto.Value:
		if !t.IsBoolean() {
			panic(jsmvcerror.JSNotIsBool)
		}
		nv, err := t.ToBoolean()
		if err != nil {
			panic(err)
		}
		return nv
	default:
		return v.(bool)
	}
}
func AssertArray(v otto.Value) []interface{} {
	if v.Class() != "Array" {
		panic(jsmvcerror.JSNotIsArray)
	}
	nv, err := v.Export()

	if err != nil {
		panic(err)
	}
	return nv.([]interface{})
}
func AssertStringArray(v otto.Value) []string {
	if v.Class() != "Array" && v.Class() != "GoArray" {
		panic(jsmvcerror.JSNotIsArray)
	}
	nv, err := v.Export()

	if err != nil {
		panic(err)
	}
	switch array := nv.(type) {
	case []string:
		return array
	case []interface{}:
		rev := make([]string, len(array))
		for i, v := range array {
			switch tv := v.(type) {
			case string:
				rev[i] = tv
			default:
				panic(jsmvcerror.JSNotIsString)
			}
		}
		return rev
	default:
		panic(fmt.Errorf("value type %T not is string array", nv))
	}
}
func AssertByteArray(value otto.Value) []byte {
	switch value.Class() {
	case "GoArray", "Array":
		nv, err := value.Export()
		if err != nil {
			panic(err)
		}
		switch tv := nv.(type) {
		case []byte:
			return tv
		case []interface{}:
			rev := make([]byte, len(tv))
			for i, v := range tv {
				switch ttv := v.(type) {
				case byte:
					rev[i] = ttv
				default:
					panic(fmt.Errorf("the value %v(%T) not is byte", ttv, ttv))
				}
			}
			return rev
		default:
			panic(fmt.Errorf("the value %v(%T) not is byte array", tv, tv))
		}
	default:
		panic(jsmvcerror.JSNotIsArray)
	}
}
func AssertString(v interface{}) string {
	switch t := v.(type) {
	case otto.Value:
		if !t.IsString() {
			panic(fmt.Errorf("the value %v not is string\n%s", v, string(debug.Stack())))
		}
		nv, err := t.ToString()
		if err != nil {
			panic(err)
		}
		return nv
	default:
		return v.(string)
	}

}
func JSToValue(rt *otto.Otto, rv interface{}) otto.Value {
	if err, ok := rv.(error); ok {
		panic(err)
	}
	if rv == nil {
		return otto.NullValue()
	}
	v, err := rt.ToValue(rv)
	if err != nil {
		panic(err)
	}
	return v
}
func AssertValue(v ...otto.Value) []interface{} {
	result := make([]interface{}, len(v))
	for i, v := range v {
		t, err := v.Export()
		if err != nil {
			panic(err)
		}
		result[i] = t
	}
	return result
}
func In(str string, list ...string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}
