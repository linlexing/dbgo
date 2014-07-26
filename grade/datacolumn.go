package grade

import (
	"fmt"
	"github.com/linlexing/datatable.go"
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/linlexing/dbhelper"
)

type DataColumn struct {
	*dbhelper.DataColumn
}

func (d *DataColumn) Clone() *DataColumn {
	return &DataColumn{d.DataColumn.Clone()}
}
func (d *DataColumn) Grade(params ...Grade) Grade {
	if len(params) == 0 {
		return Grade(oftenfun.SafeToString(d.Desc["Grade"]))
	}
	if len(params) == 1 {
		d.Desc["Grade"] = params[0]
		return params[0]
	}
	panic(fmt.Sprintf("invalid params number,only 0 or 1,param:%v", params))
}
func NewColumn(name string, gradestr Grade, dataType datatable.ColumnType, maxsize int, notnull bool) *DataColumn {
	c := &DataColumn{dbhelper.NewDataColumn(name, dataType, maxsize, notnull)}
	c.Grade(gradestr)
	return c
}

func (d *DataColumn) Object() map[string]interface{} {
	return map[string]interface{}{
		"Index":    d.Index(),
		"Name":     d.Name,
		"Desc":     d.Desc,
		"DataType": d.DataType,
		"MaxSize":  d.MaxSize,
		"NotNUll":  d.NotNull,
	}
}
