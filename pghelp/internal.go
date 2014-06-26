package pghelp

import (
	"database/sql"

	"errors"
	"fmt"
	"github.com/linlexing/dbgo/oftenfun"
	"reflect"
	"strconv"
	"strings"
)

var (
	ERROR_ColumnNumberError = errors.New("the table column number <> scan column number!")
)

func ERROR_ColumnNotFound(tabColName string) error {
	return fmt.Errorf("the column [%s] not found", tabColName)
}
func ERROR_Convert(src, dest interface{}) error {
	return fmt.Errorf("can't convert [%T] to [%T]", src, dest)
}
func internalExec(dburl, strSql string, params ...interface{}) (result_err error) {

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
	_, result_err = db.Exec(strSql, params...)
	return
}
func internalExecTx(tx *sql.Tx, strSql string, params ...interface{}) (result_err error) {
	_, result_err = tx.Exec(strSql, params...)
	return
}
func internalBatchExec(dburl, strSql string, params ...[]interface{}) (result_err error) {
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
	result_err = transact(db, func(tx *sql.Tx) error {
		return internalBatchExecTx(tx, strSql, params...)
	})
	return
}
func internalBatchExecTx(tx *sql.Tx, strSql string, params ...[]interface{}) error {
	for _, v := range params {
		stmt, err := tx.Prepare(strSql)
		if err != nil {
			return err
		}
		if _, err = stmt.Exec(v...); err != nil {
			return err
		}
	}
	return nil

}
func internalUpdateTable(dburl string, table *DataTable) (rcount int64, result_err error) {
	db, result_err := sql.Open("postgres", dburl)
	if result_err != nil {
		return
	}
	defer func() {
		if result_err == nil {
			result_err = db.Close()
		} else {
			db.Close()
		}

	}()
	result_err = transact(db, func(tx *sql.Tx) error {
		rcount, result_err = internalUpdateTableTx(tx, table)
		return result_err
	})
	return
}
func internalUpdateTableTx(tx *sql.Tx, table *DataTable) (rcount int64, result_err error) {
	changes := table.GetChange()
	if changes.RowCount == 0 {
		return
	}
	var stmt *sql.Stmt
	var result sql.Result
	var iCount int64
	if len(changes.DeleteRows) > 0 {
		strSql := buildDeleteSql(table)
		if stmt, result_err = tx.Prepare(strSql); result_err != nil {
			result_err = NewSqlError(strSql, result_err)
			return
		}
		for _, r := range changes.DeleteRows {
			if result, result_err = stmt.Exec(r.OriginData...); result_err != nil {
				result_err = NewSqlError(strSql, result_err, r.OriginData...)
				return
			}
			if iCount, result_err = result.RowsAffected(); result_err != nil {
				return
			}
			rcount += iCount

		}
	}
	if len(changes.UpdateRows) > 0 {
		strSql := buildUpdateSql(table)
		if stmt, result_err = tx.Prepare(strSql); result_err != nil {
			result_err = NewSqlError(strSql, result_err)
			return
		}
		for _, r := range changes.UpdateRows {
			if result, result_err = stmt.Exec(append(r.Data, r.OriginData...)...); result_err != nil {
				result_err = NewSqlError(strSql, result_err, append(r.Data, r.OriginData...)...)
				return
			}
			if iCount, result_err = result.RowsAffected(); result_err != nil {
				return
			}
			rcount += iCount
		}
	}

	if len(changes.InsertRows) > 0 {
		strSql := buildInsertSql(table)
		if stmt, result_err = tx.Prepare(strSql); result_err != nil {
			result_err = NewSqlError(strSql, result_err)
			return
		}
		for _, r := range changes.InsertRows {
			if _, result_err = stmt.Exec(r.Data...); result_err != nil {
				result_err = NewSqlError(strSql, result_err, r.Data...)
				return
			}
			rcount += 1
		}
	}
	return
}
func toStringSlice(v interface{}) []string {
	result := []string{}
	if arr, ok := v.([]interface{}); ok {
		result = make([]string, len(arr), len(arr))
		for i, one := range arr {
			if str, ok := one.(string); ok {
				result[i] = str
			} else {
				result[i] = ""
			}
		}
	}
	return result
}

func internalQuery(dburl string, callBack func(rows *sql.Rows) error, strSql string, params ...interface{}) (result_err error) {
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
	rows, err := db.Query(strSql, params...)
	if err != nil {
		result_err = err
		return
	}
	defer rows.Close()
	result_err = callBack(rows)
	return
}
func internalQueryTx(tx *sql.Tx, callBack func(rows *sql.Rows) error, strSql string, params ...interface{}) (result_err error) {
	rows, err := tx.Query(strSql, params...)

	if err != nil {
		result_err = err
		return
	}
	defer rows.Close()
	result_err = callBack(rows)
	return
}
func internalQueryBatch(dburl string, callBack func(rows *sql.Rows) error, strSql string, params ...[]interface{}) (result_err error) {
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
	stmt, err := db.Prepare(strSql)
	if err != nil {
		result_err = err
		return
	}
	for _, v := range params {
		rows, err := stmt.Query(v...)
		if err != nil {
			result_err = err
			return
		}
		defer rows.Close()
		result_err = callBack(rows)
		if result_err != nil {
			return
		}
	}
	return
}
func internalQueryBatchTx(tx *sql.Tx, callBack func(rows *sql.Rows) error, strSql string, params ...[]interface{}) (result_err error) {
	stmt, err := tx.Prepare(strSql)
	if err != nil {
		result_err = err
		return
	}
	for _, v := range params {
		rows, err := stmt.Query(v...)

		if err != nil {
			result_err = err
			return
		}
		defer rows.Close()
		result_err = callBack(rows)
		if result_err != nil {
			return
		}
	}
	return
}

func internalRowsFillTable(rows *sql.Rows, table *DataTable) (err error) {
	//先建立实际字段与扫描字段的顺序对应关系
	var cols []string

	if cols, err = rows.Columns(); err != nil {
		return
	}
	if len(cols) != table.ColumnCount() {
		err = ERROR_ColumnNumberError
		return
	}
	//scan index --> table column index
	trueIndex := make([]int, table.ColumnCount())
	for tabColIdx, tabColName := range table.ColumnNames() {
		bfound := false
		for scanColIdx, scanColName := range cols {
			if tabColName == scanColName {
				bfound = true
				trueIndex[scanColIdx] = tabColIdx
				break
			}
		}
		if !bfound {
			return ERROR_ColumnNotFound(tabColName)
		}
	}
	for rows.Next() {
		tabVals := table.NewPtrValues()
		//reorder vals
		vals := make([]interface{}, len(tabVals))
		for i, _ := range tabVals {
			vals[i] = tabVals[trueIndex[i]]
		}
		if err = rows.Scan(vals...); err != nil {
			return
		}
		valsToAdd := make([]interface{}, len(vals))
		for scanColIdx, tabColIdx := range trueIndex {
			if vals[scanColIdx] == nil {
				valsToAdd[tabColIdx] = nil
			} else {
				valsToAdd[tabColIdx] = reflect.ValueOf(vals[scanColIdx]).Elem().Interface()
			}
		}

		if err = table.AddValues(valsToAdd...); err != nil {
			return
		}
	}
	table.AcceptChange()
	return
}
func autoCreateColumn(cname string, value interface{}) (*DataColumn, error) {
	t := PGType{}
	t.NotNull = false
	t.MaxSize = 0
	switch value.(type) {
	case nil, []byte:
		t.Type = TypeString
	default:
		if err := t.SetReflectType(value); err != nil {
			return nil, err
		}
	}
	return NewColumnT(cname, &t, ""), nil
}
func internalRows2DataTable(rows *sql.Rows) (*DataTable, error) {
	result := NewDataTable("table1")
	var err error
	bFirst := true
	for rows.Next() {
		var vals []interface{}
		if bFirst {
			bFirst = false

			//创建表结构
			var cols []string
			if cols, err = rows.Columns(); err != nil {
				return nil, err
			}
			if vals, err = scanValues(rows, len(cols)); err != nil {
				return nil, err
			}
			for i, v := range vals {
				col, err := autoCreateColumn(cols[i], v)
				if err != nil {
					return nil, err
				}
				result.AddColumn(col)
			}
		} else {
			if vals, err = scanValues(rows, result.ColumnCount()); err != nil {
				return nil, err
			}
		}
		for i := 0; i < result.ColumnCount(); i++ {
			if _, ok := vals[i].(string); !ok && result.Columns()[i].DataType.Kind() == reflect.String {
				vals[i] = oftenfun.SafeToString(vals[i])
			}
		}
		if err := result.AddValues(vals...); err != nil {
			return nil, err
		}
	}

	result.AcceptChange()
	return result, err
}
func scanValues(r *sql.Rows, num int) ([]interface{}, error) {
	vals := make([]interface{}, num, num)
	pvals := make([]interface{}, num, num)
	for i, _ := range vals {
		pvals[i] = &vals[i]
	}
	if err := r.Scan(pvals...); err != nil {
		return nil, err
	}
	return vals, nil
}
func buildInsertSql(table *DataTable) string {
	cols := []string{}
	params := []string{}
	for i := 0; i < table.ColumnCount(); i++ {
		cols = append(cols, table.Columns()[i].Name)
		params = append(params, fmt.Sprintf("$%v", i+1))
	}
	return fmt.Sprintf("INSERT INTO %v(%v)VALUES(%v)", table.TableName, strings.Join(cols, ","), strings.Join(params, ","))
}
func buildUpdateSql(table *DataTable) string {
	sets := []string{}
	wheres := []string{}
	for i := 0; i < table.ColumnCount(); i++ {
		sets = append(sets, fmt.Sprintf("%v = $%v", table.Columns()[i].Name, i+1))
		wheres = append(wheres, fmt.Sprintf("%v = $%v", table.Columns()[i].Name, table.ColumnCount()+i+1))
	}
	return fmt.Sprintf("UPDATE %v SET %v WHERE %v", table.TableName, strings.Join(sets, ","), strings.Join(wheres, " AND "))

}
func buildDeleteSql(table *DataTable) string {
	params := []string{}
	for i, c := range table.PrimaryKeys() {
		params = append(params, fmt.Sprintf("%v = $%v", c.Name, i+1))
	}
	return fmt.Sprintf("DELETE %v WHERE %v", table.TableName, strings.Join(params, " AND "))

}
func buildSelectSql(table *DataTable) string {
	params := []string{}
	for i, c := range table.PrimaryKeys() {
		params = append(params, fmt.Sprintf("%v = $%v", c.Name, i+1))
	}
	return fmt.Sprintf("SELECT %s from %s WHERE %v", strings.Join(table.ColumnNames(), ","), table.TableName, strings.Join(params, " AND "))

}
func transact(db *sql.DB, txFunc func(*sql.Tx) error) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if p := recover(); p != nil {
			switch p := p.(type) {
			case error:
				err = p
			default:
				err = fmt.Errorf("%s", p)
			}
		}
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	return txFunc(tx)
}
func pqSignStr(str string) string {
	result := strings.Replace(strings.Replace(strconv.Quote(str), `\"`, `"`, -1), "'", "''", -1)
	return "'" + result[1:len(result)-1] + "'"
}
