package pghelp

import (
	"database/sql/driver"
	"strings"
)

type BoolSlice []bool

func (f *BoolSlice) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		tmp := parsePGArray(string(t))
		rev := make([]bool, len(tmp))
		for i, tv := range tmp {
			if tv == "t" {
				rev[i] = true
			} else {
				rev[i] = false
			}
		}
		*f = rev
		return nil
	case BoolSlice:
		*f = t
		return nil
	case []bool:
		*f = t
		return nil
	default:
		return ERROR_Convert(value, f)
	}
}
func (f BoolSlice) Value() (driver.Value, error) {
	if len(f) == 0 {
		return nil, nil
	}
	rev := make([]string, len(f))
	for i, v := range f {
		if v {
			rev[i] = "t"
		} else {
			rev[i] = "f"
		}
	}
	return "{" + strings.Join(rev, ",") + "}", nil
}

type NullBoolSlice struct {
	Slice BoolSlice
	Valid bool
}

func (f *NullBoolSlice) Scan(value interface{}) error {
	switch t := value.(type) {
	case NullBoolSlice:
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
func (f NullBoolSlice) Value() (driver.Value, error) {
	if !f.Valid {
		return nil, nil
	} else {
		return f.Slice.Value()
	}
}
func (f NullBoolSlice) IsNull() bool {
	return !f.Valid

}
