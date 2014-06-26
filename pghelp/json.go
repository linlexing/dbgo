package pghelp

import (
	"database/sql/driver"
	"encoding/json"
)

type JSON map[string]interface{}

func (f *JSON) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		if err := json.Unmarshal(t, f); err != nil {
			return err
		}
		return nil
	case string:
		if err := json.Unmarshal([]byte(t), f); err != nil {
			return err
		}
		return nil
	case JSON:
		*f = t
		return nil
	case map[string]interface{}:
		*f = t
		return nil
	default:
		return ERROR_Convert(value, f)
	}
}
func (f JSON) Value() (driver.Value, error) {
	if len(f) == 0 {
		return nil, nil
	}
	bys, err := json.Marshal(f)
	if err != nil {
		return nil, err
	}
	return bys, nil
}

type NullJSON struct {
	Json  JSON
	Valid bool
}

func (f *NullJSON) Scan(value interface{}) error {
	switch t := value.(type) {
	case NullJSON:
		*f = t
		return nil
	case nil:
		f.Valid = false
		f.Json = nil
		return nil
	default:
		f.Valid = true
		return (&f.Json).Scan(value)
	}
}
func (f NullJSON) Value() (driver.Value, error) {
	if !f.Valid {
		return nil, nil
	} else {
		return f.Json.Value()
	}
}
func (f NullJSON) IsNull() bool {
	return !f.Valid
}
