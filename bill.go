package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/linlexing/dbgo/pghelp"
	"github.com/robertkrimen/otto"
	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/parser"
	"strings"
	"time"
)

const (
	CHECK_SQL_OUTLINE_TABLENAME = "_check_outline_"
)
const (
	BILL_ADD BillOperateType = iota
	BILL_EDIT
	BILL_DELETE
	BILL_BROWSE
)

type BillOperateType int64

var (
	Error_MultiBillRecordForKey = errors.New("has multi row at same key")
)

//only one mainrow
type Bill struct {
	dbhelp      *pghelp.PGHelp
	Grade       string
	Name        string
	Tables      map[string]*pghelp.DataTable
	Checks      map[string][]*Check
	CheckResult map[string][]*CheckResult
}

func (b *Bill) MainRow() map[string]interface{} {
	if b.Tables[b.Name].RowCount() == 0 {
		return b.Tables[b.Name].NewRow()
	} else {
		return b.Tables[b.Name].GetRow(0)

	}
}
func (b *Bill) MainTable() *pghelp.DataTable {

	tab, ok := b.Tables[b.Name]
	if !ok {
		panic(fmt.Errorf("the table %q not exists", b.Name))
	}
	return tab
}

func (r *Bill) UpdateMainRow(row map[string]interface{}) error {
	if r.MainTable().RowCount() == 0 {
		return r.MainTable().AddRow(row)
	} else {
		return r.MainTable().UpdateRow(0, row)
	}
}
func (r *Bill) KeyValues() []interface{} {
	if r.MainTable().RowCount() == 0 {
		return nil
	}
	return r.MainTable().KeyValues(0)
}
func (b *Bill) loadCheckResult() error {
	if err := b.getTableCheckResult(b.Name); err != nil {
		return err
	}
	for childTableName, _ := range b.ChildTables() {

		if err := b.getTableCheckResult(childTableName); err != nil {
			return err
		}
	}
	return nil
}
func (b *Bill) getTableCheckResult(tablename string) error {
	chkIDs := make([][]interface{}, b.Tables[tablename].RowCount())
	for rowIndex := 0; rowIndex < b.Tables[tablename].RowCount(); rowIndex++ {
		chkIDs[rowIndex] = []interface{}{tablename, pghelp.StringSlice(oftenfun.SafeToStrings(b.Tables[tablename].KeyValues(rowIndex)...)), b.Grade}
	}
	chkResults := []*CheckResult{}
	if err := b.dbhelp.QueryBatch(func(rows *sql.Rows) error {
		for rows.Next() {
			var refreshBy string
			var PKs pghelp.StringSlice
			var checkID int64
			var refreshTime time.Time
			if err := rows.Scan(&PKs, &checkID, &refreshTime, &refreshBy); err != nil {
				return err
			}
			chkResults = append(chkResults, &CheckResult{
				PKs:         []string(PKs),
				CheckID:     checkID,
				RefreshTime: refreshTime,
				RefreshBy:   refreshBy,
			})
		}
		return nil
	}, SQL_GetCheckResult, chkIDs...); err != nil {
		return err
	}
	b.CheckResult[tablename] = chkResults
	return nil
}
func (b *Bill) FillByID(ids ...interface{}) (err error) {

	if err = b.dbhelp.FillTableByID(b.MainTable(), ids...); err != nil {
		return
	}
	if b.MainTable().RowCount() == 0 {
		return
	}
	if b.MainTable().RowCount() > 1 {
		err = Error_MultiBillRecordForKey
		return
	}

	for childName, rel := range b.MainTable().Desc.Relations {
		params := []string{}
		childIds := []interface{}{}
		for colIdx, colName := range rel.ChildColumns {
			params = append(params, fmt.Sprintf("%v = $%v", colName, colIdx+1))
			mainRow := b.MainRow()
			childIds = append(childIds, mainRow[rel.MainColumns[colIdx]])
		}
		strSql := fmt.Sprintf("SELECT %v from $v WHERE %v", strings.Join(b.Tables[childName].ColumnNames(), ","), b.Tables[childName].TableName, strings.Join(params, " AND "))
		if err = b.dbhelp.FillTable(b.Tables[childName], strSql, childIds...); err != nil {
			return
		}
	}
	err = b.loadCheckResult()
	return
}
func (b *Bill) Save() (err error) {
	err = pghelp.RunAtTrans(b.dbhelp.DbUrl(), func(help *pghelp.PGHelp) (err error) {
		for _, v := range b.ChildTables() {
			if _, err = help.UpdateTable(v); err != nil {
				return
			}
		}
		if _, err = help.UpdateTable(b.MainTable()); err != nil {
			return
		}
		return
	})
	return
}
func (b *Bill) ChildTables() map[string]*pghelp.DataTable {
	rev := map[string]*pghelp.DataTable{}
	for k, v := range b.Tables {
		if k != b.Name {
			rev[k] = v
		}
	}
	return rev
}
func (b *Bill) jsMainRow(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, b.MainRow())
}

//func (b *Bill) jsData(call otto.FunctionCall) otto.Value {
//	childData := map[string]interface{}{}
//	for k, v := range b.ChildTables() {
//		childData[k] = v.Rows()
//	}
//	clientCheck := map[string][]*ClientCheck{}
//	for tname, chks := range b.Checks {
//		oneTableChecks := []*ClientCheck{}
//		for _, chk := range chks {
//			oneCheck := &ClientCheck{
//				ID:           chk.ID,
//				DisplayLabel: chk.DisplayLabel,
//				Level:        chk.Level,
//				Fields:       chk.Fields,
//				RunAtServer:  chk.RunAtServer,
//				Script:       chk.Script,
//			}
//			if chk.RunAtServer {
//				oneCheck.Script = ""
//			}
//			oneTableChecks = append(oneTableChecks, oneCheck)
//		}
//		clientCheck[tname] = oneTableChecks
//	}
//	tableStruct := map[string]interface{}{}
//	for tableName, tab := range b.Tables {
//		colsStruct := map[string]interface{}{}
//		for _, col := range tab.Columns() {
//			colsStruct[col.Name] = map[string]interface{}{
//				"DataType": col.PGType.Type,
//				"MaxSize":  col.PGType.MaxSize,
//				"NotNull":  col.PGType.NotNull,
//			}
//		}
//		tableStruct[tableName] = colsStruct
//	}
//	result := map[string]interface{}{
//		"Name":        b.Name,
//		"MainRow":     b.MainRow(),
//		"Childs":      childData,
//		"Checks":      clientCheck,
//		"CheckResult": b.CheckResult,
//		"TableStruct": tableStruct,
//	}
//	return oftenfun.JSToValue(call.Otto, result)
//}
func (b *Bill) jsTables(call otto.FunctionCall) otto.Value {
	result := map[string]interface{}{}
	for k, v := range b.Tables {
		result[k] = v.Object()
	}
	return oftenfun.JSToValue(call.Otto, result)
}
func (b *Bill) jsMainTable(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, b.MainTable().Object())
}

func (b *Bill) jsKeyValues(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, b.KeyValues())
}
func (b *Bill) jsName(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, b.Name)
}
func (b *Bill) jsFillByID(call otto.FunctionCall) otto.Value {
	var ids []interface{}
	if len(call.ArgumentList) == 1 && call.Argument(0).Class() == "Array" {
		ids = oftenfun.AssertArray(call.Argument(0))
	} else {
		ids = oftenfun.AssertValue(call.ArgumentList...)
	}
	return oftenfun.JSToValue(call.Otto, b.FillByID(ids...))
}
func (b *Bill) jsSave(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, b.Save())
}
func (b *Bill) jsUpdateMainRow(call otto.FunctionCall) otto.Value {

	row := oftenfun.AssertObject(call.Argument(0))
	return oftenfun.JSToValue(call.Otto, b.UpdateMainRow(row))
}
func (b *Bill) jsChecks(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, b.Checks)
}
func (b *Bill) jsCheckResult(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, b.CheckResult)
}

func (b *Bill) Object() map[string]interface{} {
	return map[string]interface{}{
		"CheckResult":   b.jsCheckResult,
		"Checks":        b.jsChecks,
		"FillByID":      b.jsFillByID,
		"KeyValues":     b.jsKeyValues,
		"MainRow":       b.jsMainRow,
		"MainTable":     b.jsMainTable,
		"Name":          b.jsName,
		"Save":          b.jsSave,
		"Tables":        b.jsTables,
		"UpdateMainRow": b.jsUpdateMainRow,
	}
}

func (b *Bill) buildCheckChildCount(child string) (string, error) {
	if _, ok := b.MainTable().Desc.Relations[child]; !ok {
		return "", fmt.Errorf("the %q not in bill's child table list", child)
	}
	strWhere := []string{}
	for i, v := range b.MainTable().Desc.Relations[child].MainColumns {
		strWhere = append(strWhere,
			fmt.Sprintf("%s=%s.%s", b.MainTable().Desc.Relations[child].ChildColumns[i], CHECK_SQL_OUTLINE_TABLENAME, v))
	}
	return fmt.Sprintf("(select count(*) from %s where %s)", child, strings.Join(strWhere, " and ")), nil
}
func (b *Bill) processChildParse(childTableName string, exp ast.Expression) (string, bool, error) {
	switch t := exp.(type) {
	case *ast.DotExpression:
		switch left := t.Left.(type) {
		case *ast.Identifier:
			if left.Name != b.Name {
				return "", false, fmt.Errorf("the %q not is main table's name %q", left.Name, b.Name)
			}
			if !oftenfun.In(t.Identifier.Name, b.MainTable().ColumnNames()...) {
				return "", false, fmt.Errorf("the %q not is main table [%s]'s column", t.Identifier, b.Name)
			}
			strWhere := make([]string, len(b.MainTable().Desc.Relations[childTableName].MainColumns))
			for i, v := range b.MainTable().Desc.Relations[childTableName].MainColumns {
				strWhere[i] = fmt.Sprintf("%s=%s.%s", v, CHECK_SQL_OUTLINE_TABLENAME, b.MainTable().Desc.Relations[childTableName].ChildColumns[i])
			}
			return fmt.Sprintf("(select %s from %s where %s)", t.Identifier.Name, b.Name, strings.Join(strWhere, " and ")),
				b.MainTable().Columns()[b.MainTable().ColumnIndex(t.Identifier.Name)].PGType.Type == pghelp.TypeString,
				nil
		default:
			return "", false, fmt.Errorf("invalid express:%T", exp)
		}
	default:
		return "", false, fmt.Errorf("invalid express:%T", exp)
	}
}
func (b *Bill) processMainParse(exp ast.Expression) (string, bool, error) {
	switch t := exp.(type) {
	case *ast.CallExpression:
		switch call := t.Callee.(type) {
		case *ast.DotExpression:
			switch left := call.Left.(type) {
			case *ast.Identifier:
				switch call.Identifier.Name {
				case "count":
					if len(t.ArgumentList) != 0 {
						return "", false, fmt.Errorf("the count function can't have param")
					}
					str, err := b.buildCheckChildCount(left.Name)
					return str, false, err
				default:
					return "", false, fmt.Errorf("invalid function %q call", call.Identifier.Name)
				}
			default:
				return "", false, fmt.Errorf("invalid express:%T", exp)
			}

		default:
			return "", false, fmt.Errorf("invalid express:%T", exp)
		}
	default:
		return "", false, fmt.Errorf("invalid express:%T", exp)
	}
}
func (b *Bill) ParseCheckToSql(tablename, script string) (sqlWhere string, fields []string, hasSubQuery bool, err error) {
	p := &SqlParser{}
	if tablename == b.Name {
		//处理主表
		p.AdditionParse = b.processMainParse
	} else {
		p.AdditionParse = func(exp ast.Expression) (string, bool, error) {
			return b.processChildParse(tablename, exp)
		}
	}
	strFields := []string{}

	for _, col := range b.Tables[tablename].Columns() {
		if col.PGType.Type == pghelp.TypeString {
			strFields = append(strFields, col.Name)
		}
	}
	f, err := parser.ParseFunction("", script)
	if err == nil {
		sqlWhere, _, err = p.ParseExpression(f.Body.(*ast.BlockStatement).List[0].(*ast.ExpressionStatement).Expression)
		fields = p.Fields
		hasSubQuery = p.HasSubQuery
	}
	return
}
