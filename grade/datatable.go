package grade

import (
	"encoding/json"
	"github.com/linlexing/datatable.go"
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/linlexing/dbhelper"
	"github.com/robertkrimen/otto"
)

type Index struct {
	*dbhelper.Index
}

func (i *Index) Clone() *Index {
	return &Index{i.Index.Clone()}
}

func (i *Index) Grade(params ...Grade) Grade {
	if len(params) == 0 {
		return Grade(oftenfun.SafeToString(i.Desc["Grade"]))
	}
	if len(params) == 1 {
		i.Desc["Grade"] = params[0]
	}
	panic("invalid params number,only 0 or 1")
}

type DataTable struct {
	*dbhelper.DataTable
	Columns []*DataColumn
	Indexes map[string]*Index
}

func NewDataTable(tableName string, gradestr Grade) *DataTable {
	rev := &DataTable{
		dbhelper.NewDataTable(tableName),
		nil,
		map[string]*Index{},
	}
	rev.Grade(gradestr)
	return rev
}
func NewDataTableT(tab *dbhelper.DataTable) *DataTable {
	rev := &DataTable{
		tab,
		nil,
		map[string]*Index{},
	}
	for _, v := range tab.Columns {
		rev.Columns = append(rev.Columns, &DataColumn{v})
	}
	for k, v := range tab.Indexes {
		rev.Indexes[k] = &Index{v}
	}
	return rev
}

func (d *DataTable) Grade(params ...Grade) Grade {
	if len(params) == 0 {
		return Grade(oftenfun.SafeToString(d.Desc["Grade"]))
	}
	if len(params) == 1 {
		d.Desc["Grade"] = params[0]
		return params[0]
	}
	panic("invalid params number,only 0 or 1")
}

func (d *DataTable) AddColumn(col *DataColumn) *DataColumn {
	d.DataTable.AddColumn(col.DataColumn)
	d.Columns = append(d.Columns, col)
	return col
}
func (d *DataTable) AddIndex(indexName string, index *Index) {
	d.DataTable.AddIndex(indexName, index.Index)
	d.Indexes[indexName] = index
}
func (d *DataTable) Reduced(gradestr Grade) (*DataTable, bool) {
	if !gradestr.CanUse(d.Grade()) {
		return nil, false
	}
	result := NewDataTable(d.TableName, d.Grade())

	//process the columns
	for _, col := range d.Columns {
		//the key column must to be add
		if gradestr.CanUse(col.Grade()) || d.IsPrimaryKey(col.Name) {
			result.AddColumn(col.Clone())
		}
	}
	//process the pk
	result.SetPK(d.PK...)
	//process the indexes
	for idxname, idx := range d.Indexes {
		if gradestr.CanUse(idx.Grade()) {
			result.AddIndex(idxname, idx.Clone())
		}
	}
	return result, true
}
func (d *DataTable) Object() map[string]interface{} {
	return map[string]interface{}{
		"AddColumn":     d.jsAddColumn,
		"AddRow":        d.jsAddRow,
		"AddValues":     d.jsAddValues,
		"ColumnNames":   d.jsColumnNames,
		"Columns":       d.jsColumns,
		"DeleteAll":     d.jsDeleteAll,
		"DeleteRow":     d.jsDeleteRow,
		"Find":          d.jsFind,
		"GetOriginRow":  d.jsGetOriginRow,
		"GetValue":      d.jsGetValue,
		"GetValues":     d.jsGetValues,
		"HasChange":     d.jsHasChange,
		"HasPrimaryKey": d.jsHasPrimaryKey,
		"IsPrimaryKey":  d.jsIsPrimaryKey,
		"KeyValues":     d.jsKeyValues,
		"NewRow":        d.jsNewRow,
		"PK":            d.jsPK,
		"Row":           d.jsRow,
		"RowCount":      d.jsRowCount,
		"Rows":          d.jsRows,
		"Search":        d.jsSearch,
		"SetPK":         d.jsSetPK,
		"SetValues":     d.jsSetValues,
		"UpdateRow":     d.jsUpdateRow,
		"Desc":          d.Desc,
		"TableName":     d.TableName,
		"Indexes":       d.Indexes,
	}
}
func (d *DataTable) jsRows(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, d.Rows())
}
func (d *DataTable) jsRowCount(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, d.RowCount())
}
func (d *DataTable) jsSetPK(call otto.FunctionCall) otto.Value {
	vals := make([]string, len(call.ArgumentList))
	for i, v := range call.ArgumentList {
		vals[i] = oftenfun.AssertString(v)
	}
	d.SetPK(vals...)
	return otto.NullValue()

}
func (d *DataTable) jsAddColumn(call otto.FunctionCall) otto.Value {
	name := oftenfun.AssertString(call.Argument(0))
	datatype := datatable.ColumnType(oftenfun.AssertString(call.Argument(1)))
	maxsize := oftenfun.AssertInteger(call.Argument(2))
	notnull := oftenfun.AssertBool(call.Argument(3))
	grade := Grade(oftenfun.AssertString(call.Argument(4)))
	d.AddColumn(NewColumn(name, grade, datatype, maxsize, notnull))
	return otto.NullValue()
}

func (d *DataTable) jsRow(call otto.FunctionCall) otto.Value {

	index := oftenfun.AssertInteger(call.Argument(0))

	return oftenfun.JSToValue(call.Otto, d.Row(index))
}
func (d *DataTable) jsNewRow(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, d.NewRow())
}
func (d *DataTable) jsGetOriginRow(call otto.FunctionCall) otto.Value {

	index := oftenfun.AssertInteger(call.Argument(0))

	return oftenfun.JSToValue(call.Otto, d.GetOriginRow(index))
}
func (d *DataTable) jsAddRow(call otto.FunctionCall) otto.Value {

	row := oftenfun.AssertObject(call.Argument(0))

	return oftenfun.JSToValue(call.Otto, d.AddRow(row))
}
func (d *DataTable) jsGetValues(call otto.FunctionCall) otto.Value {
	index := oftenfun.AssertInteger(call.Argument(0))
	return oftenfun.JSToValue(call.Otto, d.GetValues(index))
}
func (d *DataTable) jsGetValue(call otto.FunctionCall) otto.Value {
	rowindex := oftenfun.AssertInteger(call.Argument(0))
	colindex := oftenfun.AssertInteger(call.Argument(1))
	return oftenfun.JSToValue(call.Otto, d.GetValue(rowindex, colindex))
}
func (d *DataTable) jsAddValues(call otto.FunctionCall) otto.Value {

	vals := oftenfun.AssertArray(call.Argument(0))

	return oftenfun.JSToValue(call.Otto, d.AddValues(vals...))
}
func (d *DataTable) jsSetValues(call otto.FunctionCall) otto.Value {
	index := oftenfun.AssertInteger(call.Argument(0))
	vals := oftenfun.AssertArray(call.Argument(1))
	return oftenfun.JSToValue(call.Otto, d.SetValues(index, vals...))
}
func (d *DataTable) jsUpdateRow(call otto.FunctionCall) otto.Value {

	index := oftenfun.AssertInteger(call.Argument(0))
	row := oftenfun.AssertObject(call.Argument(1))

	return oftenfun.JSToValue(call.Otto, d.UpdateRow(index, row))
}
func (d *DataTable) jsPK(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, d.PK)
}

func (d *DataTable) jsColumnNames(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, d.ColumnNames())
}
func (d *DataTable) jsColumns(call otto.FunctionCall) otto.Value {
	v, _ := call.Otto.ToValue(map[string]interface{}{})
	obj := v.Object()
	for _, col := range d.Columns {
		obj.Set(col.Name, col.Object())
	}
	return obj.Value()
}

func (d *DataTable) jsKeyValues(call otto.FunctionCall) otto.Value {

	index := oftenfun.AssertInteger(call.Argument(0))
	return oftenfun.JSToValue(call.Otto, d.KeyValues(int(index)))
}
func (d *DataTable) jsIsPrimaryKey(call otto.FunctionCall) otto.Value {
	str := oftenfun.AssertString(call.Argument(0))
	return oftenfun.JSToValue(call.Otto, d.IsPrimaryKey(str))
}
func (d *DataTable) jsHasPrimaryKey(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, d.HasPrimaryKey())
}
func (d *DataTable) jsHasChange(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, d.HasChange())
}
func (d *DataTable) jsDeleteAll(call otto.FunctionCall) otto.Value {
	d.DeleteAll()
	return otto.NullValue()
}
func (d *DataTable) jsDeleteRow(call otto.FunctionCall) otto.Value {

	index := oftenfun.AssertInteger(call.Argument(0))
	return oftenfun.JSToValue(call.Otto, d.DeleteRow(index))
}
func (d *DataTable) jsFind(call otto.FunctionCall) otto.Value {

	vals := oftenfun.AssertArray(call.Argument(0))
	return oftenfun.JSToValue(call.Otto, d.Find(vals...))
}
func (d *DataTable) jsSearch(call otto.FunctionCall) otto.Value {

	vals := oftenfun.AssertArray(call.Argument(0))

	return oftenfun.JSToValue(call.Otto, d.Search(vals...))
}
func NewDataTableJSON(data []byte) (*DataTable, error) {
	tab := NewDataTable("new", GRADE_TAG)
	if err := json.Unmarshal(data, tab); err != nil {
		return nil, err
	}
	rev := NewDataTable(tab.TableName, tab.Grade())

	for _, v := range tab.Columns {
		rev.AddColumn(NewColumn(v.Name, v.Grade(), v.DataType, v.MaxSize, v.NotNull))
	}
	for k, v := range tab.Indexes {
		rev.AddIndex(k, v)
	}
	rev.SetPK(tab.PK...)
	return rev, nil
}
