package pghelp

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/linlexing/dbgo/log"
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/robertkrimen/otto"
	"reflect"
	"strings"
)

var (
	ERROR_TableNameIsEmpty = errors.New("the table name is empty")
	ERROR_NoRecord         = errors.New("can't found record")
)

type ERROR_NotFoundTable struct {
	TableName string
}

func (e ERROR_NotFoundTable) Error() string {
	return fmt.Sprintf("can't find the table [%s]", e.TableName)
}
func ERROR_DataTypeInvalid(t interface{}) error {
	return fmt.Errorf("the type %T: %s invalid", t, t)
}

type SqlError struct {
	sql    string
	params interface{}
	err    error
}

func (s *SqlError) Error() string {
	return fmt.Sprintf("%v:\n%v\nparams:%v\n", s.err, s.sql, s.params)
}
func NewSqlError(strSql string, err error, params ...interface{}) *SqlError {
	return &SqlError{
		sql:    strSql,
		params: params,
		err:    err,
	}
}

type PGHelp struct {
	tx               *sql.Tx
	connectionString string
}

func NewPGHelp(dburl string) *PGHelp {
	return &PGHelp{connectionString: dburl}
}
func RunAtTrans(dburl string, txFunc func(help *PGHelp) error) (result_err error) {
	help := NewPGHelp(dburl)
	var db *sql.DB
	if db, result_err = sql.Open("postgres", dburl); result_err != nil {
		return
	}
	defer func() {
		if result_err == nil {
			result_err = db.Close()
		} else {
			db.Close()
		}
	}()
	if help.tx, result_err = db.Begin(); result_err != nil {
		return
	}
	defer func() {
		if p := recover(); p != nil {
			switch p := p.(type) {
			case error:
				result_err = p
			default:
				result_err = fmt.Errorf("%s", p)
			}
		}
		if result_err != nil {
			help.tx.Rollback()
			return
		}
		result_err = help.tx.Commit()
	}()
	return txFunc(help)
}

func (p *PGHelp) Schema() (*PGSchema, error) {
	//获取当前用户默认的schema信息
	if tab, err := p.GetDataTable(SQL_GetCurrentSchemaAndDesc); err != nil {
		return nil, err
	} else {
		schema := &PGSchema{
			Name: tab.GetValue(0, 0).(string),
			Desc: &PGSchemaDesc{}}
		json.Unmarshal([]byte(tab.GetValues(0)[1].(string)), schema.Desc)
		return schema, nil
	}

}
func (p *PGHelp) UpdateTable(table *DataTable) (rcount int64, result_err error) {
	if p.tx == nil {
		rcount, result_err = internalUpdateTable(p.connectionString, table)
	} else {
		rcount, result_err = internalUpdateTableTx(p.tx, table)
	}
	return
}
func (p *PGHelp) TableByID(tname string, ids ...interface{}) (result *DataTable, err error) {
	var tab *DataTable
	if tab, err = p.Table(tname); err != nil {
		return
	}
	err = p.FillTableByID(tab, ids...)
	return
}
func (p *PGHelp) jsGetDataTable(call otto.FunctionCall) otto.Value {

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
func (p *PGHelp) Object() map[string]interface{} {
	return map[string]interface{}{
		"GetDataTable": p.jsGetDataTable,
	}
}

func (p *PGHelp) FillTableByID(tab *DataTable, ids ...interface{}) (err error) {
	strSql := buildSelectSql(tab)
	err = p.FillTable(tab, strSql, ids...)
	return
}
func (p *PGHelp) FillTableWhere(tab *DataTable, strWhere string, params ...interface{}) (err error) {
	return p.FillTable(tab, fmt.Sprintf("SELECT %s from %s WHERE %s",
		strings.Join(tab.ColumnNames(), ","), tab.TableName, strWhere), params...)
}
func (p *PGHelp) FillTable(tab *DataTable, strSql string, params ...interface{}) (result_err error) {
	result_err = p.Query(func(rows *sql.Rows) (err error) {
		err = internalRowsFillTable(rows, tab)
		return
	}, strSql, params...)
	return
}
func (p *PGHelp) DbUrl() string {
	return p.connectionString
}
func (p *PGHelp) GetDataTable(strSql string, params ...interface{}) (table *DataTable, result_err error) {
	result_err = p.Query(func(rows *sql.Rows) (err error) {
		table, err = internalRows2DataTable(rows)
		return
	}, strSql, params...)
	return
}
func decodePQDesc(descStr string) map[string]interface{} {
	desc := map[string]interface{}{}
	if err := json.Unmarshal([]byte(descStr), &desc); err != nil {
		return map[string]interface{}{}
	}
	return desc

}
func (p *PGHelp) QueryOne(strSql string, params ...interface{}) (result_err error) {
	result_err = p.QueryRow(strSql, params[:len(params)-1], params[len(params)-1])
	return
}
func (p *PGHelp) QueryRow(strSql string, params []interface{}, dest ...interface{}) (result_err error) {
	return p.Query(func(rows *sql.Rows) error {
		if !rows.Next() {
			return ERROR_NoRecord
		}
		if err := rows.Scan(dest...); err != nil {
			return err
		}
		return nil
	}, strSql, params...)

}
func (p *PGHelp) Query(callBack func(rows *sql.Rows) error, strSql string, params ...interface{}) (result_err error) {
	log.SQLTRACE.Printf("query row:[%v];params:%v\n", strSql, params)
	defer func() {
		if result_err != nil {
			result_err = NewSqlError(strSql, result_err, params...)
		}
	}()
	if p.tx == nil {
		result_err = internalQuery(p.connectionString, callBack, strSql, params...)
	} else {
		result_err = internalQueryTx(p.tx, callBack, strSql, params...)
	}
	return
}
func (p *PGHelp) QueryBatch(callBack func(rows *sql.Rows) error, strSql string, params ...[]interface{}) (result_err error) {
	log.SQLTRACE.Printf("batch query row:[%v];params:%v\n", strSql, params)
	defer func() {
		if result_err != nil {
			ps := make([]interface{}, len(params))
			for i, v := range params {
				ps[i] = v
			}
			result_err = NewSqlError(strSql, result_err, ps...)
		}
	}()
	if p.tx == nil {
		result_err = internalQueryBatch(p.connectionString, callBack, strSql, params...)
	} else {
		result_err = internalQueryBatchTx(p.tx, callBack, strSql, params...)
	}
	return
}

func (p *PGHelp) GetString(strSQL string, params ...interface{}) string {
	v := sql.NullString{}
	if err := p.QueryOne(strSQL, append(params, &v)...); err != nil {
		return ""
	} else {
		return v.String
	}
}
func (p *PGHelp) GetInt(strSQL string, params ...interface{}) int {
	return int(p.GetInt64(strSQL, params))
}
func (p *PGHelp) GetInt64(strSQL string, params ...interface{}) int64 {
	v := sql.NullInt64{}
	if err := p.QueryOne(strSQL, append(params, &v)...); err != nil {
		return 0
	} else {
		return v.Int64
	}
}
func (p *PGHelp) GetBool(strSQL string, params ...interface{}) bool {
	v := sql.NullBool{}
	if err := p.QueryOne(strSQL, append(params, &v)...); err != nil {
		return false
	} else {
		return v.Bool
	}
}
func (p *PGHelp) getTableDesc(tname string) string {
	return p.GetString(SQL_GetTableDesc, tname)
}
func (p *PGHelp) alterTableDesc(tname string, desc *TableDesc) error {
	if bys, err := json.Marshal(desc); err != nil {
		return err
	} else {
		return p.ExecuteSql(fmt.Sprintf(SQL_AlterTableDesc, tname, pqSignStr(string(bys))))
	}
}
func (p *PGHelp) TableExists(tablename string) bool {
	b := p.GetBool(SQL_TableExists, tablename)
	log.SQLTRACE.Printf("table [%v] exists:%v", tablename, b)
	return b
}

func (p *PGHelp) Table(tablename string) (*DataTable, error) {
	result := NewDataTable(tablename)
	if !p.TableExists(tablename) {
		return result, ERROR_NotFoundTable{tablename}
	}
	//获取描述
	json.Unmarshal([]byte(p.getTableDesc(tablename)), result.Desc)
	//获取字段

	tMeta, err := p.GetDataTable(SQL_TableColumns, tablename)
	if err != nil {
		return nil, err
	}
	for i := 0; i < tMeta.RowCount(); i++ {
		oneRow := tMeta.GetRow(i)
		dt := PGType{}
		if err := dt.SetDBType(oneRow["datatype"].(string)); err != nil {
			return nil, err
		}
		dt.NotNull = oneRow["notnull"].(bool)
		newColumn := NewColumnT(oneRow["columnname"].(string), &dt, oftenfun.SafeToString(oneRow["def"]))
		json.Unmarshal(oftenfun.SafeToBytes(oneRow["desc"]), newColumn.Desc)
		result.AddColumn(newColumn)
	}
	//获取主键
	tPks, err := p.GetDataTable(SQL_TablePrimaryKeys, tablename)
	if err != nil {
		return nil, err
	}
	pkColumns := []string{}
	if tPks.RowCount() > 0 {
		pkColumns = tPks.GetColumnStrings(0)
		result.PKConstraintName = tPks.GetColumnStrings(1)[0]
		result.SetPK(pkColumns...)
	}
	//获取索引
	tIndexes, err := p.GetDataTable(SQL_TableIndexes, tablename)
	if err != nil {
		return nil, err
	}
	for i := 0; i < tIndexes.RowCount(); i++ {
		oneRow := tIndexes.GetRow(i)
		oneIndex := NewIndex(oneRow["indexname"].(string))
		json.Unmarshal([]byte(oneRow["define"].(string)), oneIndex.Desc)
		result.Indexes[oneRow["desc"].(string)] = oneIndex

	}

	return result, nil
}

func (p *PGHelp) ExecuteSql(strSql string, params ...interface{}) (result_err error) {
	log.SQLTRACE.Printf("execute sql:[%v];params:%v\n", strSql, params)
	defer func() {
		if result_err != nil {
			result_err = NewSqlError(strSql, result_err, params...)
		}
	}()
	if p.tx == nil {
		result_err = internalExec(p.connectionString, strSql, params...)
	} else {
		result_err = internalExecTx(p.tx, strSql, params...)
	}
	return
}
func (p *PGHelp) ExecuteBatch(strSql string, params ...[]interface{}) (result_err error) {
	defer func() {
		if result_err != nil {
			ps := make([]interface{}, len(params))
			for i, v := range params {
				ps[i] = v
			}
			result_err = NewSqlError(strSql, result_err, ps...)
		}
	}()
	if p.tx == nil {
		result_err = internalBatchExec(p.connectionString, strSql, params...)
	} else {
		result_err = internalBatchExecTx(p.tx, strSql, params...)
	}
	return
}
func (p *PGHelp) GetDataTableBatch(strSql string, params ...[]interface{}) (table *DataTable, result_err error) {
	result_err = p.QueryBatch(func(rows *sql.Rows) (err error) {
		if table == nil {
			table, result_err = internalRows2DataTable(rows)
			if result_err != nil {
				return
			}
		} else {
			result_err = internalRowsFillTable(rows, table)
			if result_err != nil {
				return
			}
		}
		return
	}, strSql, params...)
	return
}
func (p *PGHelp) alterColumnDesc(tname, cname string, desc *ColumnDesc) error {
	if bys, err := json.Marshal(desc); err != nil {
		return err
	} else {
		return p.ExecuteSql(fmt.Sprintf(SQL_AlterColumnDesc, tname, cname, pqSignStr(string(bys))))
	}
}
func (p *PGHelp) dropConstraint(tname, cname string) error {
	return p.ExecuteSql(fmt.Sprintf(SQL_DropConstraint, tname, cname))
}
func (p *PGHelp) createColumn(tname, cname string, dt *PGType, def string) error {
	defstr := ""
	if len(def) > 0 {
		defstr = "DEFAULT " + def
	}
	return p.ExecuteSql(fmt.Sprintf(SQL_CreateColumn, tname, cname, dt.DBString(), defstr))
}
func (p *PGHelp) createTable(tname string) error {
	return p.ExecuteSql(fmt.Sprintf(SQL_CreateTable, tname))
}
func (p *PGHelp) alterIndexDesc(name string, desc *IndexDesc) error {
	if bys, err := json.Marshal(desc); err != nil {
		return err
	} else {
		return p.ExecuteSql(fmt.Sprintf(SQL_AlterIndexDesc, name, pqSignStr(string(bys))))
	}
}
func (p *PGHelp) createPrimaryKey(tname string, cname []string) error {
	return p.ExecuteSql(fmt.Sprintf(SQL_CreatePrimaryKey, tname, strings.Join(cname, ",")))
}
func (p *PGHelp) renameColumn(tname, oldName, newName string) error {
	return p.ExecuteSql(fmt.Sprintf(SQL_RenameColumn, tname, oldName, newName))
}
func (p *PGHelp) dropIndex(name string) error {
	return p.ExecuteSql(fmt.Sprintf(SQL_DropIndex, name))
}
func (p *PGHelp) dropColumns(tname string, columns ...string) error {
	for _, v := range columns {
		if err := p.ExecuteSql(fmt.Sprintf(SQL_DropColumn, tname, v)); err != nil {
			return err
		}
	}
	return nil
}

func (p *PGHelp) alterColumnType(tname, cname string, dt *PGType) error {
	return p.ExecuteSql(fmt.Sprintf(SQL_AlterColumnType, tname, cname, dt.DBString()))
}
func (p *PGHelp) dropColumnNotNull(tname, cname string) error {
	return p.ExecuteSql(fmt.Sprintf(SQL_DropColumnNotNull, tname, cname))
}
func (p *PGHelp) setColumnNotNull(tname, cname string) error {
	return p.ExecuteSql(fmt.Sprintf(SQL_SetColumnNotNull, tname, cname))
}
func (p *PGHelp) dropColumnDefault(tname, cname string) error {
	return p.ExecuteSql(fmt.Sprintf(SQL_DropColumnDefault, tname, cname))
}
func (p *PGHelp) setColumnDefault(tname, cname, def string) error {
	return p.ExecuteSql(fmt.Sprintf(SQL_SetColumnDefault, tname, cname, def))
}
func (p *PGHelp) UpdateStruct(newStruct *DataTable, grade string) error {
	log.SQLTRACE.Printf("start update %v's struct...\n", newStruct.TableName)
	if len(newStruct.TableName) == 0 {
		return ERROR_TableNameIsEmpty
	}
	tablename := newStruct.TableName
	oldStruct, err := p.Table(tablename)
	//如果表不存在，则创建空表
	if err != nil {
		if _, ok := err.(ERROR_NotFoundTable); !ok {
			return err
		} else {
			p.createTable(newStruct.TableName)
		}
	}
	//首先判断主关键字是否有变化
	bKeyChange := false
	if !reflect.DeepEqual(oldStruct.GetPK(), newStruct.GetPK()) {
		bKeyChange = true
	}
	if !bKeyChange {
		//判断主键的数据类型是否变化
		oldPks := oldStruct.PrimaryKeys()
		newPks := newStruct.PrimaryKeys()
		for i := 0; i < len(oldPks); i++ {
			if !reflect.DeepEqual(oldPks[i].DataType, newPks[i].DataType) {
				bKeyChange = true
				break
			}
		}
	}
	if bKeyChange && oldStruct.HasPrimaryKey() {
		//删除主键
		if err := p.dropConstraint(tablename, oldStruct.PKConstraintName); err != nil {
			return err
		}
	}
	//找出相对应的一对字段
	oldColumns := oldStruct.Columns()
	newColumns := []*DataColumn{}
	//仅读取上级或者本级的字段
	for _, v := range newStruct.Columns() {
		if strings.HasPrefix(grade, v.Desc.Grade) {
			newColumns = append(newColumns, v)
		}
	}
	type FoundColumn struct {
		OldColumn *DataColumn
		NewColumn *DataColumn
	}
	foundColumns := []FoundColumn{}

	for _, vNew := range newColumns {
		trueNewName := vNew.Name

		if vNew.Desc.OriginName != "" && vNew.Name != vNew.Desc.OriginName {
			trueNewName = vNew.Desc.OriginName
		}
		for _, vOld := range oldColumns {
			if vOld.Name == trueNewName {
				foundColumns = append(foundColumns, FoundColumn{vOld, vNew})
			}
		}
	}
	//删除字段
	for _, oldColumn := range oldColumns {
		bFound := false
		for _, foundColumn := range foundColumns {
			if oldColumn == foundColumn.OldColumn {
				bFound = true
				break
			}
		}
		//找不到的，且是本级或上级的字段，才需要删除
		if !bFound && strings.HasPrefix(grade, oldColumn.Desc.Grade) {
			if err := p.dropColumns(tablename, oldColumn.Name); err != nil {
				return err
			}
		}
	}

	//修改字段类型或者重命名
	for _, column := range foundColumns {
		//改名
		if column.NewColumn.Name != column.OldColumn.Name {
			if err := p.renameColumn(tablename, column.OldColumn.Name, column.NewColumn.Name); err != nil {
				return err
			}
		}
		//改类型
		if column.OldColumn.PGType.Type != column.NewColumn.PGType.Type ||
			column.OldColumn.PGType.MaxSize != column.NewColumn.PGType.MaxSize {
			if err := p.alterColumnType(tablename, column.NewColumn.Name, column.NewColumn.PGType); err != nil {
				return err
			}
		}
		//去除NotNull
		if column.OldColumn.PGType.NotNull && !column.NewColumn.PGType.NotNull {
			if err := p.dropColumnNotNull(tablename, column.NewColumn.Name); err != nil {
				return err
			}
		}
		//设置NotNull
		if !column.OldColumn.PGType.NotNull && column.NewColumn.PGType.NotNull {
			if err := p.setColumnNotNull(tablename, column.NewColumn.Name); err != nil {
				return err
			}
		}
		//去除Default
		if len(column.OldColumn.Default) > 0 && len(column.NewColumn.Default) == 0 {
			if err := p.dropColumnDefault(tablename, column.NewColumn.Name); err != nil {
				return err
			}
		}
		//设置Default
		if column.OldColumn.Default != column.NewColumn.Default && len(column.NewColumn.Default) > 0 {
			if err := p.setColumnDefault(tablename, column.NewColumn.Name, column.NewColumn.Default); err != nil {
				return err
			}
		}
		//改描述
		if *column.NewColumn.Desc != *column.OldColumn.Desc {
			if err := p.alterColumnDesc(tablename, column.NewColumn.Name, column.NewColumn.Desc); err != nil {
				return err
			}
		}
	}
	//新增字段
	for _, newColumn := range newColumns {
		bFound := false
		for _, foundColumn := range foundColumns {
			if newColumn == foundColumn.NewColumn {
				bFound = true
				break
			}
		}
		if !bFound {
			if err := p.createColumn(tablename, newColumn.Name, newColumn.PGType, newColumn.Default); err != nil {
				return err
			}
			//有必要就加描述
			if newColumn.Desc != nil {
				if err := p.alterColumnDesc(tablename, newColumn.Name, newColumn.Desc); err != nil {
					return err
				}
			}
		}
	}
	if bKeyChange && newStruct.HasPrimaryKey() {
		//创建主键
		if err := p.createPrimaryKey(tablename, newStruct.GetPK()); err != nil {
			return err
		}
	}
	//处理索引
	//删除不存在的,并修改存在的
	for idxName, oldIdx := range oldStruct.Indexes {
		//本级或上级的索引才纳入考察范围
		if newIdx, ok := newStruct.Indexes[idxName]; ok && strings.HasPrefix(grade, newIdx.Desc.Grade) {
			if oldIdx.Define != newIdx.Define {
				if err := p.dropIndex(idxName); err != nil {
					return err
				}
				if err := p.ExecuteSql(newIdx.Define, newStruct.TableName, idxName); err != nil {
					return err
				}
			}
			if !reflect.DeepEqual(oldIdx.Desc, newIdx.Desc) {
				if err := p.alterIndexDesc(idxName, newIdx.Desc); err != nil {
					return err
				}
			}
		} else if strings.HasPrefix(grade, oldIdx.Desc.Grade) { //仅本级或上级不存在的索引才删除
			if err := p.dropIndex(idxName); err != nil {
				return err
			}
		}
	}
	//新增索引
	for idxName, newIdx := range newStruct.Indexes {
		//仅取本级及上级的索引
		if strings.HasPrefix(grade, newIdx.Desc.Grade) {
			if _, ok := oldStruct.Indexes[idxName]; !ok {
				if err := p.ExecuteSql(newIdx.Define, newStruct.TableName, idxName); err != nil {
					return err
				}
				if err := p.alterIndexDesc(idxName, newIdx.Desc); err != nil {
					return err
				}
			}
		}
	}
	//处理表的描述
	if !reflect.DeepEqual(oldStruct.Desc, newStruct.Desc) {
		if err := p.alterTableDesc(newStruct.TableName, newStruct.Desc); err != nil {
			return err
		}
	}
	return nil
}
