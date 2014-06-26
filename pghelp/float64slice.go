package pghelp

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

type Float64Slice []float64

func (f *Float64Slice) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		tmp := parsePGArray(string(t))
		rev := make([]float64, len(tmp))
		for i, tv := range tmp {
			var err error
			rev[i], err = strconv.ParseFloat(tv, 64)
			if err != nil {
				return err
			}
		}
		*f = rev
		return nil
	case Float64Slice:
		*f = t
		return nil
	case []float64:
		*f = t
		return nil
	default:
		return ERROR_Convert(value, f)
	}

}
func (f Float64Slice) Value() (driver.Value, error) {
	if len(f) == 0 {
		return nil, nil
	}
	rev := make([]string, len(f))
	for i, v := range f {
		rev[i] = fmt.Sprintf("%.17f", v)
	}
	return "{" + strings.Join(rev, ",") + "}", nil
}

type NullFloat64Slice struct {
	Slice Float64Slice
	Valid bool
}

func (f *NullFloat64Slice) Scan(value interface{}) error {
	switch t := value.(type) {
	case NullFloat64Slice:
		*f = t
		return nil
	case nil:
		f.Valid = false
		f.Slice = nil
		return nil
	default:
		f.Valid = true
		return (&f.Slice).Scan(value)
	}
}
func (f NullFloat64Slice) Value() (driver.Value, error) {
	if !f.Valid {
		return nil, nil
	} else {
		return f.Slice.Value()
	}
}
func (f NullFloat64Slice) IsNull() bool {
	return !f.Valid
}
