package grade

import (
	"fmt"
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/linlexing/pghelper"
)

type DataColumn struct {
	*pghelper.DataColumn
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
func NewColumnT(name string, gradestr Grade, dt *pghelper.PGType, def string) *DataColumn {
	rev := &DataColumn{
		pghelper.NewColumnT(name, dt, def),
	}
	rev.Grade(gradestr)
	return rev
}
func NewColumn(name string, gradestr Grade, dataType pghelper.PGTypeType, param ...interface{}) *DataColumn {
	rev := &DataColumn{
		pghelper.NewColumn(name, dataType, param...),
	}
	rev.Grade(gradestr)
	return rev
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
