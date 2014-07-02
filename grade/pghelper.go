package grade

import (
	"fmt"
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/linlexing/pghelper"
	"github.com/robertkrimen/otto"
)

type PGHelper struct {
	*pghelper.PGHelper
}

func NewPGHelper(dburl string) *PGHelper {
	return NewPGHelperT(pghelper.NewPGHelper(dburl))
}
func NewPGHelperT(ahelp *pghelper.PGHelper) *PGHelper {
	return &PGHelper{ahelp}
}
func (p *PGHelper) GetDataTable(strSql string, params ...interface{}) (*DataTable, error) {
	tab, err := p.PGHelper.GetDataTable(strSql, params...)
	if err != nil {
		return nil, err
	}
	return NewDataTableT(tab), nil

}
func (p *PGHelper) Table(tablename string) (*DBTable, error) {
	tab, err := p.PGHelper.Table(tablename)
	if err != nil {
		return nil, err
	}
	return NewDBTable(p, NewDataTableT(tab.DataTable)), nil

}
func (p *PGHelper) UpdateStruct(newStruct *DataTable) error {
	oldStruct, err := p.Table(newStruct.TableName)
	if _, ok := err.(pghelper.ERROR_NotFoundTable); err != nil && !ok {
		return err
	}
	if oldStruct == nil {
		return p.PGHelper.UpdateStruct(nil, newStruct.DataTable)
	}
	trueOld, ok := oldStruct.DataTable.Reduced(newStruct.Grade())
	if !ok {
		return fmt.Errorf("the oldStruct's grade is %q,newStruct can't use it", oldStruct.DataTable.Grade())
	}
	return p.PGHelper.UpdateStruct(trueOld.DataTable, newStruct.DataTable)

}

func RunAtTrans(dburl string, txFunc func(help *PGHelper) error) (result_err error) {
	return pghelper.RunAtTrans(dburl, func(help *pghelper.PGHelper) error {
		return txFunc(NewPGHelperT(help))
	})
}
func (p *PGHelper) jsTable(call otto.FunctionCall) otto.Value {
	tablename := oftenfun.AssertString(call.Argument(0))
	tab, err := p.Table(tablename)
	if err != nil {
		panic(err)
	}
	return oftenfun.JSToValue(call.Otto, tab.Object())
}
func (p *PGHelper) jsGetDataTable(call otto.FunctionCall) otto.Value {

	strSql := oftenfun.AssertString(call.Argument(0))
	params := make([]interface{}, len(call.ArgumentList)-1)
	for i := 1; i < len(call.ArgumentList); i++ {
		var err error
		params[i-1], err = call.ArgumentList[i].Export()
		if err != nil {
			panic(err)
		}
	}
	result, err := p.GetDataTable(strSql, params...)
	if err != nil {
		panic(err)
	}
	return oftenfun.JSToValue(call.Otto, result.Object())
}
func (p *PGHelper) Object() map[string]interface{} {
	return map[string]interface{}{
		"GetDataTable": p.jsGetDataTable,
		"Table":        p.jsTable,
	}
}
