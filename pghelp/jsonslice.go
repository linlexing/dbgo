package pghelp

import (
	"database/sql/driver"
	"encoding/json"
)

type JSONSlice []map[string]interface{}

func (f *JSONSlice) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		vals := parsePGArray(string(t))
		rev := make(JSONSlice, len(vals))
		for i, v := range vals {
			if err := json.Unmarshal([]byte(v), &(rev[i])); err != nil {
				return err
			}

		}
		*f = rev
		return nil
	case string:
		vals := parsePGArray(t)
		rev := make(JSONSlice, len(vals))
		for i, v := range vals {
			if err := json.Unmarshal([]byte(v), &(rev[i])); err != nil {
				return err
			}

		}
		*f = rev
		return nil
	case JSONSlice:
		*f = t
		return nil
	case []map[string]interface{}:
		*f = t

		return nil
	default:
		return ERROR_Convert(value, f)
	}
}
func (f JSONSlice) Value() (driver.Value, error) {
	if len(f) == 0 {
		return nil, nil
	}
	rev := make([]string, len(f))
	for i, v := range f {
		bys, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		rev[i] = string(bys)
	}
	return encodePGArray(rev), nil
}

type NullJSONSlice struct {
	Slice JSONSlice
	Valid bool
}

func (f *NullJSONSlice) Scan(value interface{}) error {
	switch t := value.(type) {
	case NullJSONSlice:
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
func (f NullJSONSlice) Value() (driver.Value, error) {
	if !f.Valid {
		return nil, nil
	} else {
		return f.Slice.Value()
	}
}
func (f NullJSONSlice) IsNull() bool {
	return !f.Valid
}
