// datatable project datatable.go
package pghelp

import (
	"database/sql"
	"fmt"
	"github.com/linlexing/datatable.go"
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/robertkrimen/otto"
	"reflect"
	"strings"
)

type pkIndex struct {
	dataTable *DataTable
	index     []int
}
type Check struct {
	ID           int64
	DisplayLabel string
	Level        int64
	Fields       []string
	Script       string
	Grade        string
}
type MainChildRelation struct {
	Grade        string
	MainColumns  []string
	ChildColumns []string
}

func (m *MainChildRelation) Clone() *MainChildRelation {
	maincol := make([]string, len(m.MainColumns))
	copy(maincol, m.MainColumns)
	childcol := make([]string, len(m.ChildColumns))
	copy(childcol, m.ChildColumns)
	return &MainChildRelation{
		Grade:        m.Grade,
		MainColumns:  maincol,
		ChildColumns: childcol,
	}
}

type IndexDesc struct {
	Grade string `json:",omitempty"`
}

func (i *IndexDesc) Clone() *IndexDesc {
	return &IndexDesc{i.Grade}
}

type Index struct {
	Define string
	Desc   *IndexDesc
}

func (i *Index) Clone() *Index {
	return &Index{
		Define: i.Define,
		Desc:   i.Desc.Clone(),
	}
}

type TableDesc struct {
	ViewTemplate string                        `json:",omitempty"`
	MainTable    string                        `json:",omitempty"`
	Relations    map[string]*MainChildRelation `json:",omitempty"`
	Grade        string                        `json:",omitempty"`
}
type DataTable struct {
	*datatable.DataTable
	Desc    *TableDesc
	Indexes map[string]*Index
	columns []*DataColumn
	Checks  []*Check
}

func NewIndex(define string) *Index {
	return &Index{Define: define, Desc: &IndexDesc{}}
}

func NewDataTable(name string) *DataTable {
	return &DataTable{
		datatable.NewDataTable(name),
		&TableDesc{},
		map[string]*Index{},
		nil,
		nil,
	}
}
func (d *DataTable) PrimaryKeys() []*DataColumn {
	pks := d.DataTable.PrimaryKeys()
	rev := make([]*DataColumn, len(pks))
	for i, v := range pks {
		rev[i] = d.columns[v.Index()]
	}
	return rev
}

//Assign each column empty value pointer,General used by database/sql scan
func (d *DataTable) NewPtrValues() []interface{} {
	result := make([]interface{}, d.ColumnCount())
	for i, c := range d.Columns() {
		result[i] = c.PtrZeroValue()
	}
	return result
}
func nullToNil(value ...interface{}) []interface{} {
	rev := make([]interface{}, len(value))
	for i, v := range value {
		switch tv := v.(type) {
		case IsNull:
			if tv.IsNull() {
				rev[i] = nil
			}
			tmp, err := tv.Value()
			if err != nil {
				panic(err)
			}
			rev[i] = tmp
		default:
			rev[i] = tv
		}
	}
	return rev
}
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (d *DataTable) AsTabText(columns ...string) string {
	result := []string{}
	if len(columns) > 0 {
		result = append(result, strings.Join(columns, "\t"))
	} else {
		result = append(result, strings.Join(d.ColumnNames(), "\t"))
	}
	for i := 0; i < d.RowCount(); i++ {
		r := d.GetRow(i)
		line := []string{}
		for j := 0; j < d.ColumnCount(); j++ {
			c := d.Columns()[j]
			if len(columns) > 0 && !stringInSlice(c.Name, columns) {
				continue
			}
			if r[c.Name] == nil {
				line = append(line, "")
			} else {
				line = append(line, fmt.Sprintf("%v", r[c.Name]))
			}
		}
		result = append(result, strings.Join(line, "\t"))
	}
	return strings.Join(result, "\n")
}

//convert NULL to nil
func (d *DataTable) GetValue(rowIndex, colIndex int) interface{} {
	return nullToNil(d.DataTable.GetValue(rowIndex, colIndex))[0]
}
func (d *DataTable) GetColumnValues(columnIndex int) []interface{} {
	newValues := make([]interface{}, d.RowCount())
	for i := 0; i < d.RowCount(); i++ {
		newValues[i] = d.GetValue(i, columnIndex)
	}
	return newValues
}
func (d *DataTable) GetColumnStrings(columnIndex int) []string {
	rev := make([]string, d.RowCount())
	for i, v := range d.GetColumnValues(columnIndex) {
		rev[i] = oftenfun.SafeToString(v)
	}
	return rev
}
func (d *DataTable) nilToNULL(row []interface{}) ([]interface{}, error) {
	rev := make([]interface{}, len(row))
	for i, col := range d.Columns() {
		tmp := col.PtrZeroValue()
		switch t := tmp.(type) {
		case sql.Scanner:
			err := t.Scan(row[i])
			if err != nil {
				return nil, err
			}
			rev[i] = reflect.ValueOf(tmp).Elem().Interface()
		default:
			if row[i] == nil {
				panic(fmt.Errorf("nil --> %s error", col.DataType.String()))
			}
			rev[i] = row[i]
		}
	}
	return rev, nil
}
func (d *DataTable) getSequenceValues(r map[string]interface{}) []interface{} {
	vals := make([]interface{}, d.ColumnCount())
	for i, col := range d.columns {
		var ok bool
		if vals[i], ok = r[col.Name]; !ok {
			panic(fmt.Errorf("can't find column:[%s] at %v", col.Name, r))
		}

	}
	return vals

}
func (d *DataTable) AddRow(r map[string]interface{}) error {
	return d.AddValues(d.getSequenceValues(r)...)
}
func (d *DataTable) NewRow() map[string]interface{} {
	result := map[string]interface{}{}
	for _, col := range d.columns {
		result[col.Name] = nullToNil(col.ZeroValue())[0]
	}
	return result
}
func (d *DataTable) GetRow(rowIndex int) map[string]interface{} {
	vals := d.GetValues(rowIndex)
	result := map[string]interface{}{}
	for i, col := range d.columns {
		result[col.Name] = vals[i]
	}
	return result
}
func (d *DataTable) Rows() []map[string]interface{} {
	rev := []map[string]interface{}{}
	for i := 0; i < d.RowCount(); i++ {
		vals := d.GetValues(i)
		result := map[string]interface{}{}
		for i, col := range d.columns {
			result[col.Name] = vals[i]
		}
		rev = append(rev, result)
	}
	return rev
}
func (d *DataTable) UpdateRow(rowIndex int, r map[string]interface{}) error {
	return d.SetValues(rowIndex, d.getSequenceValues(r)...)
}
func (d *DataTable) AddValues(vs ...interface{}) (err error) {
	v, err := d.nilToNULL(vs)
	if err != nil {
		return err
	}
	return d.DataTable.AddValues(v...)
}
func (d *DataTable) SetValues(rowIndex int, values ...interface{}) (err error) {
	vs, err := d.nilToNULL(values)
	if err != nil {
		return err
	}
	return d.DataTable.SetValues(rowIndex, vs...)
}
func (d *DataTable) GetValues(rowIndex int) []interface{} {
	return nullToNil(d.DataTable.GetValues(rowIndex)...)
}
func (d *DataTable) AddColumn(col *DataColumn) *DataColumn {

	d.DataTable.AddColumn(col.DataColumn)
	d.columns = append(d.columns, col)
	return col
}
func (d *DataTable) Columns() []*DataColumn {
	return d.columns

}
func (d *DataTable) Object() map[string]interface{} {
	return map[string]interface{}{
		"AddColumn":        d.jsAddColumn,
		"AddRow":           d.jsAddRow,
		"AddValues":        d.jsAddValues,
		"ColumnNames":      d.jsColumnNames,
		"Columns":          d.jsColumns,
		"DeleteAll":        d.jsDeleteAll,
		"DeleteRow":        d.jsDeleteRow,
		"Find":             d.jsFind,
		"GetOriginRow":     d.jsGetOriginRow,
		"GetPK":            d.jsGetPK,
		"GetRow":           d.jsGetRow,
		"GetValue":         d.jsGetValue,
		"GetValues":        d.jsGetValues,
		"HasChange":        d.jsHasChange,
		"HasPrimaryKey":    d.jsHasPrimaryKey,
		"IsPrimaryKey":     d.jsIsPrimaryKey,
		"KeyValues":        d.jsKeyValues,
		"NewRow":           d.jsNewRow,
		"PrimaryKeys":      d.jsPrimaryKeys,
		"RowCount":         d.jsRowCount,
		"Search":           d.jsSearch,
		"SetPK":            d.jsSetPK,
		"SetValues":        d.jsSetValues,
		"UpdateRow":        d.jsUpdateRow,
		"Desc":             d.Desc,
		"PKConstraintName": d.PKConstraintName,
		"TableName":        d.TableName,
		"Indexes":          d.Indexes,
	}
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
	dt := PGTypeType(oftenfun.AssertInteger(call.Argument(1)))
	param := []interface{}{}
	if len(call.ArgumentList) > 2 {
		param = append(param, oftenfun.AssertBool(call.Argument(2)))
	}
	if len(call.ArgumentList) > 3 {
		param = append(param, oftenfun.AssertInteger(call.Argument(3)))

	}
	if len(call.ArgumentList) > 4 {
		param = append(param, oftenfun.AssertString(call.Argument(4)))
	}
	d.AddColumn(NewColumn(name, dt, param...))
	return otto.NullValue()
}

func (d *DataTable) jsGetRow(call otto.FunctionCall) otto.Value {

	index := oftenfun.AssertInteger(call.Argument(0))

	return oftenfun.JSToValue(call.Otto, d.GetRow(index))
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
func (d *DataTable) jsGetPK(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, d.GetPK())
}

func (d *DataTable) jsColumnNames(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, d.ColumnNames())
}
func (d *DataTable) jsColumns(call otto.FunctionCall) otto.Value {
	v, _ := call.Otto.ToValue(map[string]interface{}{})
	obj := v.Object()
	for _, col := range d.Columns() {
		obj.Set(col.Name, col.Object())
	}
	return obj.Value()
}
func (d *DataTable) jsPrimaryKeys(call otto.FunctionCall) otto.Value {
	result := []map[string]interface{}{}
	for _, col := range d.PrimaryKeys() {
		result = append(result, col.Object())
	}
	return oftenfun.JSToValue(call.Otto, result)
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
