package pghelp

import (
	"database/sql/driver"
)

type Bytea []byte

func (b *Bytea) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		*b = t
		return nil
	case string:
		*b = []byte(t)
		return nil
	case Bytea:
		*b = t
		return nil
	default:
		return ERROR_Convert(value, b)

	}
}
func (b Bytea) Value() (driver.Value, error) {
	return b, nil
}

type NullBytea struct {
	Bytea []byte
	Valid bool
}

func (f *NullBytea) Scan(value interface{}) error {
	if value == nil {
		f.Valid = false
		f.Bytea = nil
	} else {
		f.Bytea, f.Valid = value.([]byte)
	}

	return nil
}
func (f NullBytea) Value() (driver.Value, error) {
	if !f.Valid {
		return nil, nil
	}
	return f.Bytea, nil
}

func (this NullBytea) IsNull() bool {
	return !this.Valid
}
