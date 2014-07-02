package grade

import (
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/linlexing/pghelper"
	"github.com/robertkrimen/otto"
)

const (
	ImportBatch = 1000
	ExportBatch = 1000
)

type DBTable struct {
	*DataTable
	dbHelper *PGHelper
}

func NewDBTable(ahelp *PGHelper, table *DataTable) *DBTable {
	return &DBTable{table, ahelp}
}
func (t *DBTable) Fill(strSql string, params ...interface{}) (result_err error) {
	return pghelper.NewDBTable(t.dbHelper.PGHelper, t.DataTable.DataTable).Fill(strSql, params...)
}
func (t *DBTable) FillByID(ids ...interface{}) (err error) {
	return pghelper.NewDBTable(t.dbHelper.PGHelper, t.DataTable.DataTable).FillByID(ids...)
}
func (t *DBTable) FillWhere(strWhere string, params ...interface{}) (err error) {
	return pghelper.NewDBTable(t.dbHelper.PGHelper, t.DataTable.DataTable).FillWhere(strWhere, params...)
}
func (t *DBTable) Save() (rcount int64, result_err error) {
	return pghelper.NewDBTable(t.dbHelper.PGHelper, t.DataTable.DataTable).Save()

}
func (t *DBTable) Count(strWhere string, params ...interface{}) (count int64, err error) {
	return pghelper.NewDBTable(t.dbHelper.PGHelper, t.DataTable.DataTable).Count(strWhere, params...)
}
func (t *DBTable) BatchFillWhere(callBack func(table *DBTable, eof bool) error, batchRow int64, strWhere string, params ...interface{}) (err error) {
	return pghelper.NewDBTable(t.dbHelper.PGHelper, t.DataTable.DataTable).BatchFillWhere(
		func(table *pghelper.DBTable, eof bool) error {
			return callBack(t, eof)
		}, batchRow, strWhere, params...)
}

func (t *DBTable) jsFill(call otto.FunctionCall) otto.Value {
	sql := oftenfun.AssertString(call.Argument(0))
	vals := oftenfun.AssertValue(call.ArgumentList[1:]...)
	return oftenfun.JSToValue(call.Otto, t.Fill(sql, vals...))
}
func (t *DBTable) jsFillByID(call otto.FunctionCall) otto.Value {
	var vals []interface{}
	if call.Argument(0).Class() == "Array" && len(call.ArgumentList) > 1 {
		vals = oftenfun.AssertArray(call.Argument(0))
	} else {
		vals = oftenfun.AssertValue(call.ArgumentList...)
	}
	return oftenfun.JSToValue(call.Otto, t.FillByID(vals...))
}
func (t *DBTable) jsFillWhere(call otto.FunctionCall) otto.Value {
	sql := oftenfun.AssertString(call.Argument(0))
	vals := oftenfun.AssertValue(call.ArgumentList[1:]...)
	return oftenfun.JSToValue(call.Otto, t.FillWhere(sql, vals...))
}
func (t *DBTable) Object() map[string]interface{} {
	m := t.DataTable.Object()
	m["Fill"] = t.jsFill
	m["FillByID"] = t.jsFillByID
	m["FillWhere"] = t.jsFillWhere
	return m
}
func (t *DBTable) UpdateStruct() error {
	return t.dbHelper.UpdateStruct(t.DataTable)
}
