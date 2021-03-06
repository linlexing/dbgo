package grade

import (
	"github.com/linlexing/datatable.go"
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/robertkrimen/otto"
)

const (
	ImportBatch = 1000
	ExportBatch = 1000
)

type DBTable struct {
	*DataTable
	dbHelper *DBHelper
}

func NewDBTable(ahelp *DBHelper, table *DataTable) *DBTable {
	return &DBTable{table, ahelp}
}
func (t *DBTable) Fill(strSql string, params ...interface{}) (result_err error) {
	return t.FillT(strSql, nil, params...)
}
func (t *DBTable) FillT(strSql string, templateParam map[string]interface{}, params ...interface{}) (result_err error) {
	return t.dbHelper.FillTableT(t.DataTable.DataTable, strSql, templateParam, params...)
}
func (t *DBTable) FillByID(vals ...interface{}) (err error) {
	//flat array value
	ids := []interface{}{}
	switch tv := vals[0].(type) {
	case []string:
		for _, v := range tv {
			ids = append(ids, v)
		}
	case []interface{}:
		for _, v := range tv {
			ids = append(ids, v)
		}
	default:
		ids = append(ids, tv)
	}
	//convert string value to interface{}
	idVals := make([]interface{}, len(ids))
	for i, v := range t.PK {
		switch tv := ids[i].(type) {
		case string:
			col := t.Columns[t.ColumnIndex(v)]
			if col.DataType != datatable.String {
				vv, err := col.DecodeString(tv)
				if err != nil {
					return err
				}
				idVals[i] = vv
			} else {
				idVals[i] = tv
			}
		default:
			idVals[i] = tv
		}
	}
	return t.dbHelper.FillTable(t.DataTable.DataTable, t.SelectAllByID(), idVals...)
}
func (t *DBTable) FillWhere(strWhere string, params ...interface{}) (err error) {
	return t.FillWhereT(strWhere, nil, params...)
}
func (t *DBTable) FillWhereT(strWhere string, templateParam map[string]interface{}, params ...interface{}) (err error) {
	return t.dbHelper.FillTable(t.DataTable.DataTable, t.SelectAllByWhere(strWhere), params...)
}
func (t *DBTable) Save() (rcount int64, result_err error) {
	return t.dbHelper.SaveChange(t.DataTable.DataTable)
}
func (t *DBTable) Count(strWhere string, params ...interface{}) (count int64, err error) {
	return t.CountT(strWhere, nil, params...)
}
func (t *DBTable) CountT(strWhere string, templateParam map[string]interface{}, params ...interface{}) (count int64, err error) {
	if strWhere != "" {
		strWhere = "\nwhere\n" + strWhere
	}
	v, err := t.dbHelper.QueryOne("select count(*) from "+t.TableName+strWhere, params...)
	return v.(int64), nil
}
func (t *DBTable) DBHelper() *DBHelper {
	return t.dbHelper
}
func (t *DBTable) jsFill(call otto.FunctionCall) otto.Value {
	sql := oftenfun.AssertString(call.Argument(0))
	vals := oftenfun.AssertValue(call.ArgumentList[1:]...)
	return oftenfun.JSToValue(call.Otto, t.Fill(sql, vals...))
}
func (t *DBTable) jsSave(call otto.FunctionCall) otto.Value {
	rcount, err := t.Save()
	if err != nil {
		panic(err)
	}
	return oftenfun.JSToValue(call.Otto, rcount)
}
func (t *DBTable) jsFillByID(call otto.FunctionCall) otto.Value {
	var vals []interface{}
	if call.Argument(0).Class() == "Array" && len(call.ArgumentList) == 1 {
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
func (t *DBTable) jsDBHelper(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, t.DBHelper().Object())
}
func (t *DBTable) Object() map[string]interface{} {
	m := t.DataTable.Object()
	m["Fill"] = t.jsFill
	m["FillByID"] = t.jsFillByID
	m["FillWhere"] = t.jsFillWhere
	m["DataTable"] = t.DataTable.Object()
	m["DBHelper"] = t.jsDBHelper
	m["Save"] = t.jsSave
	return m
}
func (t *DBTable) UpdateStruct() error {
	return t.dbHelper.UpdateStruct(t.DataTable)
}
