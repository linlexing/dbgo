package pghelp

import (
	"database/sql/driver"
)

type StringSlice []string

func (f *StringSlice) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		*f = parsePGArray(string(t))
		return nil
	case string:
		*f = parsePGArray(t)
		return nil
	case StringSlice:
		*f = t
		return nil
	case []string:
		*f = t
		return nil
	default:
		return ERROR_Convert(value, f)
	}
}
func (f StringSlice) Value() (driver.Value, error) {
	if len(f) == 0 {
		return nil, nil
	}
	return encodePGArray(f), nil
}

type NullStringSlice struct {
	Slice StringSlice
	Valid bool
}

func (f *NullStringSlice) Scan(value interface{}) error {
	switch t := value.(type) {
	case NullStringSlice:
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
func (f NullStringSlice) Value() (driver.Value, error) {
	if !f.Valid {
		return nil, nil
	} else {
		return f.Slice.Value()
	}
}
func (f NullStringSlice) IsNull() bool {
	return !f.Valid
}
