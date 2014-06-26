package pghelp

import (
	"database/sql/driver"
	"github.com/linlexing/datatable.go"
)

type IsNull interface {
	driver.Valuer
	IsNull() bool
}
type ColumnDesc struct {
	Aliases    string `json:",omitempty"`
	OriginName string `json:",omitempty"`
	Grade      string `json:",omitempty"`
	Desc       string `json:",omitempty"`
}

func (d *ColumnDesc) Clone() *ColumnDesc {
	v := ColumnDesc{}
	v = *d
	return &v
}

type DataColumn struct {
	*datatable.DataColumn
	PGType  *PGType
	Default string
	Desc    *ColumnDesc
}

func (d *DataColumn) Clone() *DataColumn {
	result := DataColumn{}
	result = *d
	result.Desc = d.Desc.Clone()
	result.PGType = d.PGType.Clone()
	return &result
}
func (d *DataColumn) Object() map[string]interface{} {
	return map[string]interface{}{
		"Index":   d.Index(),
		"Name":    d.Name,
		"Desc":    d.Desc,
		"Default": d.Default,
		"PGType":  d.PGType,
	}
}
func NewColumnT(name string, dt *PGType, def string) *DataColumn {
	return &DataColumn{
		datatable.NewDataColumn(name, dt.ReflectType()),
		dt,
		def,
		&ColumnDesc{},
	}

}
func NewColumn(name string, dataType PGTypeType, param ...interface{}) *DataColumn {
	dt := NewPGType(dataType, 0, false)
	if len(param) > 0 {
		dt.NotNull = param[0].(bool)
	}
	if len(param) > 1 {
		dt.MaxSize = param[1].(int)

	}
	def := ""
	if len(param) > 2 {
		def = param[2].(string)
	}
	if len(param) > 3 {
		panic("too much param")
	}
	return NewColumnT(name, dt, def)
} /*
func NewStringColumn(name string, notNull bool, maxSize int) *DataColumn {
	if notNull {
		return newPGDataColumn(name, reflect.TypeOf(""), notNull, maxSize, def)
	} else {
		return newPGDataColumn(name, reflect.TypeOf(NullString{}), notNull, maxSize, def)
	}

}
func NewFloat64Column(name string, notNull bool, def string) *DataColumn {
	if notNull {
		return newPGDataColumn(name, reflect.TypeOf(float64(0)), notNull, 0, def)
	} else {
		return newPGDataColumn(name, reflect.TypeOf(NullFloat64{}), notNull, 0, def)
	}
}

func NewInt64Column(name string, notNull bool, def string) *DataColumn {
	if notNull {
		return newPGDataColumn(name, reflect.TypeOf(int64(0)), notNull, 0, def)
	} else {
		return newPGDataColumn(name, reflect.TypeOf(NullInt64{}), notNull, 0, def)
	}
}

func NewBoolColumn(name string, notNull bool, def string) *DataColumn {
	if notNull {
		return newPGDataColumn(name, reflect.TypeOf(true), notNull, 0, def)
	} else {
		return newPGDataColumn(name, reflect.TypeOf(NullBool{}), notNull, 0, def)

	}
}

func NewByteaColumn(name string, notNull bool, def string) *DataColumn {
	if notNull {
		return newPGDataColumn(name, reflect.TypeOf(Bytea{}), notNull, 0, def)
	} else {
		return newPGDataColumn(name, reflect.TypeOf(NullBytea{}), notNull, 0, def)
	}

}

func NewTimeColumn(name string, notNull bool, def string) *DataColumn {
	if notNull {
		return newPGDataColumn(name, reflect.TypeOf(time.Now()), notNull, 0, def)
	} else {
		return newPGDataColumn(name, reflect.TypeOf(NullTime{}), notNull, 0, def)
	}
}
func NewJSONColumn(name string, notNull bool, def string) *DataColumn {
	if notNull {
		return newPGDataColumn(name, reflect.TypeOf(JSON{}), notNull, 0, def)
	} else {
		return newPGDataColumn(name, reflect.TypeOf(NullJSON{}), notNull, 0, def)
	}
}

func NewStringArrayColumn(name string, notNull bool, maxSize int, def string) *DataColumn {
	if notNull {
		return newPGDataColumn(name, reflect.TypeOf(StringSlice{}), notNull, maxSize, def)
	} else {
		return newPGDataColumn(name, reflect.TypeOf(NullStringSlice{}), notNull, maxSize, def)
	}

}
func NewFloat64ArrayColumn(name string, notNull bool, def string) *DataColumn {
	if notNull {

		return newPGDataColumn(name, reflect.TypeOf(Float64Slice{}), notNull, 0, def)
	} else {
		return newPGDataColumn(name, reflect.TypeOf(NullFloat64Slice{}), notNull, 0, def)
	}
}
func NewInt64ArrayColumn(name string, notNull bool, def string) *DataColumn {
	if notNull {
		return newPGDataColumn(name, reflect.TypeOf(Int64Slice{}), notNull, 0, def)
	} else {
		return newPGDataColumn(name, reflect.TypeOf(NullInt64Slice{}), notNull, 0, def)
	}
}
func NewBoolArrayColumn(name string, notNull bool, def string) *DataColumn {
	if notNull {
		return newPGDataColumn(name, reflect.TypeOf(BoolSlice{}), notNull, 0, def)
	} else {
		return newPGDataColumn(name, reflect.TypeOf(NullBoolSlice{}), notNull, 0, def)
	}
}
func NewTimeArrayColumn(name string, notNull bool, def string) *DataColumn {
	if notNull {
		return newPGDataColumn(name, reflect.TypeOf(TimeSlice{}), notNull, 0, def)
	} else {
		return newPGDataColumn(name, reflect.TypeOf(NullTimeSlice{}), notNull, 0, def)
	}
}
func NewJSONArrayColumn(name string, notNull bool, def string) *DataColumn {
	if notNull {
		return newPGDataColumn(name, reflect.TypeOf(JSONSlice{}), notNull, 0, def)
	} else {
		return newPGDataColumn(name, reflect.TypeOf(NullJSONSlice{}), notNull, 0, def)
	}
}
*/
