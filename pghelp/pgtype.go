package pghelp

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"time"
)

const (
	TypeString = iota
	TypeBool
	TypeInt64
	TypeFloat64
	TypeTime
	TypeBytea
	TypeStringSlice
	TypeBoolSlice
	TypeInt64Slice
	TypeFloat64Slice
	TypeTimeSlice
	TypeJSON
	TypeJSONSlice
)

var (
	regVarchar      = regexp.MustCompile(`^character varying\((\d+)\)$`)
	regVarcharArray = regexp.MustCompile(`^character varying\((\d+)\)\[\]$`)
)

type PGTypeType int
type PGType struct {
	Type    PGTypeType
	MaxSize int
	NotNull bool
}

func NewPGType(t PGTypeType, maxsize int, notnull bool) *PGType {
	return &PGType{t, maxsize, notnull}
}
func (p *PGType) Clone() *PGType {
	rev := PGType{}
	rev = *p
	return &rev
}

func (p *PGType) DBString() string {
	notnull := ""
	if p.NotNull {
		notnull = " NOT NULL"
	}
	switch p.Type {
	case TypeBool:
		return "boolean" + notnull
	case TypeBoolSlice:
		return "boolean[]" + notnull
	case TypeBytea:
		return "bytea" + notnull
	case TypeFloat64:
		return "double precision" + notnull
	case TypeFloat64Slice:
		return "double precision[]" + notnull
	case TypeInt64:
		return "bigint" + notnull
	case TypeInt64Slice:
		return "bigint[]" + notnull
	case TypeString:
		if p.MaxSize == 0 {
			return "text" + notnull
		} else {
			return fmt.Sprintf("character varying(%v)", p.MaxSize) + notnull
		}
	case TypeStringSlice:
		if p.MaxSize == 0 {
			return "text[]" + notnull
		} else {
			return fmt.Sprintf("character varying(%v)[]", p.MaxSize) + notnull
		}
	case TypeTime:
		return "timestamp without time zone" + notnull
	case TypeTimeSlice:
		return "timestamp without time zone[]" + notnull
	case TypeJSON:
		return "jsonb" + notnull
	case TypeJSONSlice:
		return "jsonb[]" + notnull
	default:
		panic(ERROR_DataTypeInvalid(p))

	}
}
func (p *PGType) ReflectType() reflect.Type {
	if !p.NotNull {
		switch p.Type {
		case TypeBool:
			return reflect.TypeOf(NullBool{})
		case TypeBoolSlice:
			return reflect.TypeOf(NullBoolSlice{})
		case TypeBytea:
			return reflect.TypeOf(NullBytea{})
		case TypeFloat64:
			return reflect.TypeOf(NullFloat64{})
		case TypeFloat64Slice:
			return reflect.TypeOf(NullFloat64Slice{})
		case TypeInt64:
			return reflect.TypeOf(NullInt64{})
		case TypeInt64Slice:
			return reflect.TypeOf(NullInt64Slice{})
		case TypeString:
			return reflect.TypeOf(NullString{})
		case TypeStringSlice:
			return reflect.TypeOf(NullStringSlice{})
		case TypeTime:
			return reflect.TypeOf(NullTime{})
		case TypeTimeSlice:
			return reflect.TypeOf(NullTimeSlice{})
		case TypeJSON:
			return reflect.TypeOf(NullJSON{})
		case TypeJSONSlice:
			return reflect.TypeOf(NullJSONSlice{})
		default:
			panic(ERROR_DataTypeInvalid(p))

		}

	} else {
		switch p.Type {
		case TypeBool:
			return reflect.TypeOf(true)
		case TypeBoolSlice:
			return reflect.TypeOf(BoolSlice{})
		case TypeBytea:
			return reflect.TypeOf(Bytea{})
		case TypeFloat64:
			return reflect.TypeOf(float64(0))
		case TypeFloat64Slice:
			return reflect.TypeOf(Float64Slice{})
		case TypeInt64:
			return reflect.TypeOf(int64(0))
		case TypeInt64Slice:
			return reflect.TypeOf(Int64Slice{})
		case TypeString:
			return reflect.TypeOf("")
		case TypeStringSlice:
			return reflect.TypeOf(StringSlice{})
		case TypeTime:
			return reflect.TypeOf(time.Time{})
		case TypeTimeSlice:
			return reflect.TypeOf(TimeSlice{})
		case TypeJSON:
			return reflect.TypeOf(JSON{})
		case TypeJSONSlice:
			return reflect.TypeOf(JSONSlice{})
		default:
			panic(ERROR_DataTypeInvalid(p))

		}
	}
}

/*func DecodeReflectType(t reflect.Type) *PGType {
	result := &PGType{}
	if typeIsBool(t) {
		result.Type = TypeBool
	} else if typeIsBoolArray(t) {
		result.Type = TypeBoolSlice
	} else if typeIsBytea(t) {
		result.Type = TypeBytea
	} else if typeIsFloat64(t) {
		result.Type = TypeFloat64
	} else if typeIsFloat64Array(t) {
		result.Type = TypeFloat64Slice
	} else if typeIsInt64(t) {
		result.Type = TypeInt64
	} else if typeIsInt64Array(t) {
		result.Type = TypeInt64Slice
	} else if typeIsString(t) {
		result.Type = TypeString
	} else if typeIsStringArray(t) {
		result.Type = TypeStringSlice
	} else if typeIsTime(t) {
		result.Type = TypeTime
	} else if typeIsTimeArray(t) {
		result.Type = TypeTimeSlice
	} else if typeIsJSON(t) {
		result.Type = TypeJSON
	} else if typeIsJSONArray(t) {
		result.Type = TypeJSONSlice
	} else {
		panic(ERROR_DataTypeInvalid(t))
	}
	return result
}*/
func (p *PGType) SetReflectType(value interface{}) error {
	switch value.(type) {
	case string:
		p.Type = TypeString
	case int64:
		p.Type = TypeInt64
	case bool:
		p.Type = TypeBool
	case float64:
		p.Type = TypeFloat64
	case time.Time:
		p.Type = TypeTime
	case []byte:
		p.Type = TypeBytea
	default:
		return ERROR_DataTypeInvalid(value)
	}
	return nil
}
func (p *PGType) SetDBType(t string) error {
	switch {
	case t == "text":
		p.Type = TypeString
		p.MaxSize = 0
	case t == "text[]":
		p.Type = TypeStringSlice
		p.MaxSize = 0
	case t == "boolean":
		p.Type = TypeBool
	case t == "boolean[]":
		p.Type = TypeBoolSlice
	case t == "bigint":
		p.Type = TypeInt64
	case t == "bigint[]":
		p.Type = TypeInt64Slice
	case t == "double precision":
		p.Type = TypeFloat64
	case t == "double precision[]":
		p.Type = TypeFloat64Slice
	case regVarchar.MatchString(t):
		p.Type = TypeString
		var err error

		if p.MaxSize, err = strconv.Atoi(regVarchar.FindStringSubmatch(t)[1]); err != nil {
			return err
		}
	case regVarcharArray.MatchString(t):
		p.Type = TypeStringSlice
		var err error
		if p.MaxSize, err = strconv.Atoi(regVarcharArray.FindStringSubmatch(t)[1]); err != nil {
			return err
		}
	case t == "timestamp without time zone" ||
		t == "timestamp with time zone" ||
		t == "date":
		p.Type = TypeTime
	case t == "timestamp without time zone[]" ||
		t == "timestamp with time zone[]" ||
		t == "date[]":
		p.Type = TypeTimeSlice
	case t == "bytea":
		p.Type = TypeBytea
	case t == "jsonb" || t == "json":
		p.Type = TypeJSON
	case t == "jsonb[]" || t == "json[]":
		p.Type = TypeJSONSlice
	default:
		return ERROR_DataTypeInvalid(t)
	}
	return nil
} /*
func typeIsStringArray(t reflect.Type) bool {
	return t.Kind() == reflect.Slice && typeIsString(t.Elem())
}
func typeIsInt64Array(t reflect.Type) bool {
	return t.Kind() == reflect.Slice && typeIsInt64(t.Elem())
}
func typeIsFloat64Array(t reflect.Type) bool {
	return t.Kind() == reflect.Slice && typeIsFloat64(t.Elem())
}
func typeIsBoolArray(t reflect.Type) bool {
	return t.Kind() == reflect.Slice && typeIsBool(t.Elem())
}
func typeIsTimeArray(t reflect.Type) bool {
	return t.Kind() == reflect.Slice && typeIsTime(t.Elem())
}
func typeIsByteaArray(t reflect.Type) bool {
	return t.Kind() == reflect.Slice && typeIsBytea(t.Elem())
}
func typeIsBool(t reflect.Type) bool {
	return t.Kind() == reflect.Bool
}
func typeIsJSON(t reflect.Type) bool {
	return t.Kind() == reflect.Map &&
		t.Key().Kind() == reflect.String &&
		t.Elem().Kind() == reflect.Interface
}
func typeIsJSONArray(t reflect.Type) bool {
	return t.Kind() == reflect.Slice &&
		t.Elem().Kind() == reflect.Map &&
		t.Elem().Key().Kind() == reflect.String &&
		t.Elem().Elem().Kind() == reflect.Interface
}

func typeIsBytea(t reflect.Type) bool {
	return t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8
}
func typeIsFloat64(t reflect.Type) bool {
	return t.Kind() == reflect.Float64
}
func typeIsInt64(t reflect.Type) bool {
	return t.Kind() == reflect.Int64
}
func typeIsString(t reflect.Type) bool {
	return t.Kind() == reflect.String
}
func typeIsTime(t reflect.Type) bool {
	return reflect.DeepEqual(t, reflect.TypeOf(time.Now()))
}
*/
