package grade

import (
	"crypto/rand"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/linlexing/datatable.go"
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/linlexing/dbhelper"
	_ "github.com/linlexing/myhelper"
	_ "github.com/linlexing/pghelper"
	"github.com/robertkrimen/otto"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DBHelper struct {
	*dbhelper.DBHelper
}

func NewDBHelper(dburl string) *DBHelper {
	urls := strings.SplitN(dburl, " ", 2)
	return newDBHelperT(dbhelper.NewDBHelper(urls[0], urls[1]))
}
func newDBHelperT(ahelp *dbhelper.DBHelper) *DBHelper {
	return &DBHelper{ahelp}
}
func (p *DBHelper) GetData(strSql string, params ...interface{}) (*DataTable, error) {
	tab, err := p.DBHelper.GetData(strSql, params...)
	if err != nil {
		return nil, err
	}
	return NewDataTableT(tab), nil

}
func (p *DBHelper) SelectLimit(srcSql string, pkFields []string, startKeyValue map[string]interface{}, selectCols []string, where string, orderby []string, limit int) (*DataTable, error) {
	tab, err := p.DBHelper.SelectLimit(srcSql, pkFields, startKeyValue, selectCols, where, orderby, limit)
	if err != nil {
		return nil, err
	}
	return NewDataTableT(tab), nil
}
func (p *DBHelper) SelectLimitT(srcSql string, templateParam map[string]interface{}, pkFields []string, startKeyValue map[string]interface{}, selectCols []string, where string, orderby []string, limit int) (*DataTable, error) {
	tab, err := p.DBHelper.SelectLimitT(srcSql, templateParam, pkFields, startKeyValue, selectCols, where, orderby, limit)
	if err != nil {
		return nil, err
	}
	return NewDataTableT(tab), nil
}
func (p *DBHelper) GetDataT(strSql string, templateParams map[string]interface{}, params ...interface{}) (*DataTable, error) {
	tab, err := p.DBHelper.GetDataT(strSql, templateParams, params...)
	if err != nil {
		return nil, err
	}
	return NewDataTableT(tab), nil

}
func (p *DBHelper) Table(tablename string) (*DataTable, error) {
	tab, err := p.DBHelper.Table(tablename)
	if err != nil {
		return nil, err
	}
	return NewDataTableT(tab), nil
}
func (p *DBHelper) DBTable(tablename string) (*DBTable, error) {
	tab, err := p.DBHelper.Table(tablename)
	if err != nil {
		return nil, err
	}
	t := NewDBTable(p, NewDataTableT(tab))
	return t, nil
}

func (p *DBHelper) UpdateStruct(newStruct *DataTable) error {
	var oldStruct *DataTable
	if exists, err := p.TableExists(newStruct.TableName); err != nil {
		return err
	} else {
		if exists {
			oldStruct, err = p.Table(newStruct.TableName)
			if err != nil {
				return err
			}
		}
	}
	if oldStruct == nil {
		return p.DBHelper.UpdateStruct(nil, newStruct.DataTable, nil)
	}
	trueOld, ok := oldStruct.Reduced(newStruct.Grade())
	if !ok {
		return fmt.Errorf("the oldStruct's grade is %q,newStruct can't use it", oldStruct.Grade())
	}

	return p.DBHelper.UpdateStruct(trueOld.DataTable, newStruct.DataTable, oldStruct.ColumnNames())

}
func (p *DBHelper) jsOpen(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, p.Open())
}
func (p *DBHelper) jsClose(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, p.Close())
}
func (p *DBHelper) jsBegin(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, p.Begin())
}
func (p *DBHelper) jsCommit(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, p.Commit())
}
func (p *DBHelper) jsRollback(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, p.Rollback())
}

func (p *DBHelper) jsSelectLimit(call otto.FunctionCall) otto.Value {
	srcSql := oftenfun.AssertString(call.Argument(0))
	pkFields := oftenfun.AssertStringArray(call.Argument(1))
	startKeyValue := oftenfun.AssertObject(call.Argument(2))
	selectCols := oftenfun.AssertStringArray(call.Argument(3))
	where := oftenfun.AssertString(call.Argument(4))
	orderby := oftenfun.AssertStringArray(call.Argument(5))
	limit := oftenfun.AssertInteger(call.Argument(6))
	result, err := p.SelectLimit(srcSql, pkFields, startKeyValue, selectCols, where, orderby, limit)
	if err != nil {
		panic(err)
	}
	return oftenfun.JSToValue(call.Otto, result.Object())
}

func (p *DBHelper) jsSelectLimitT(call otto.FunctionCall) otto.Value {
	srcSql := oftenfun.AssertString(call.Argument(0))
	templateParam := oftenfun.AssertObject(call.Argument(1))
	pkFields := oftenfun.AssertStringArray(call.Argument(2))
	startKeyValue := oftenfun.AssertObject(call.Argument(3))
	selectCols := oftenfun.AssertStringArray(call.Argument(4))
	where := oftenfun.AssertString(call.Argument(5))
	orderby := oftenfun.AssertStringArray(call.Argument(6))
	limit := oftenfun.AssertInteger(call.Argument(7))
	result, err := p.SelectLimitT(srcSql, templateParam, pkFields, startKeyValue, selectCols, where, orderby, limit)
	if err != nil {
		panic(err)
	}
	return oftenfun.JSToValue(call.Otto, result.Object())
}
func (p *DBHelper) jsExecT(call otto.FunctionCall) otto.Value {
	sql := oftenfun.AssertString(call.Argument(0))
	templateParam := oftenfun.AssertObject(call.Argument(1))
	args := []interface{}{}
	if len(call.ArgumentList) > 2 {
		args = oftenfun.AssertValue(call.ArgumentList[2:]...)
	}
	_, err := p.ExecT(sql, templateParam, args...)
	if err != nil {
		panic(err)
	}
	return otto.UndefinedValue()
}

func (p *DBHelper) jsExec(call otto.FunctionCall) otto.Value {
	sql := oftenfun.AssertString(call.Argument(0))
	args := []interface{}{}
	if len(call.ArgumentList) > 1 {
		args = oftenfun.AssertValue(call.ArgumentList[1:]...)
	}
	_, err := p.Exec(sql, args...)
	if err != nil {
		panic(err)
	}
	return otto.UndefinedValue()
}
func (p *DBHelper) jsGoExecT(call otto.FunctionCall) otto.Value {
	sql := oftenfun.AssertString(call.Argument(0))
	templateParam := oftenfun.AssertObject(call.Argument(1))
	err := p.GoExecT(sql, templateParam)
	if err != nil {
		panic(err)
	}
	return otto.UndefinedValue()
}
func (p *DBHelper) jsTable(call otto.FunctionCall) otto.Value {
	tablename := oftenfun.AssertString(call.Argument(0))
	tab, err := p.Table(tablename)
	if err != nil {
		panic(err)
	}
	return oftenfun.JSToValue(call.Otto, tab.Object())
}
func (p *DBHelper) jsGetData(call otto.FunctionCall) otto.Value {

	strSql := oftenfun.AssertString(call.Argument(0))
	params := oftenfun.AssertValue(call.ArgumentList[1:]...)
	result, err := p.GetData(strSql, params...)
	if err != nil {
		panic(err)
	}
	return oftenfun.JSToValue(call.Otto, result.Object())
}
func (p *DBHelper) jsGetDataT(call otto.FunctionCall) otto.Value {

	strSql := oftenfun.AssertString(call.Argument(0))
	templateParam := oftenfun.AssertObject(call.Argument(1))
	params := oftenfun.AssertValue(call.ArgumentList[2:]...)
	result, err := p.GetDataT(strSql, templateParam, params...)
	if err != nil {
		panic(err)
	}
	return oftenfun.JSToValue(call.Otto, result.Object())
}
func (p *DBHelper) jsQueryOne(call otto.FunctionCall) otto.Value {
	strSql := oftenfun.AssertString(call.Argument(0))
	params := oftenfun.AssertValue(call.ArgumentList[1:]...)
	rev, err := p.QueryOne(strSql, params...)
	if err != nil {
		panic(err)
	}
	return oftenfun.JSToValue(call.Otto, rev)
}
func (p *DBHelper) jsQStr(call otto.FunctionCall) otto.Value {
	strSql := oftenfun.AssertString(call.Argument(0))
	params := oftenfun.AssertValue(call.ArgumentList[1:]...)
	rev, err := p.QueryOne(strSql, params...)
	if err != nil {
		panic(err)
	}
	return oftenfun.JSToValue(call.Otto, oftenfun.SafeToString(rev))
}
func (p *DBHelper) Object() map[string]interface{} {
	return map[string]interface{}{
		"GetData":      p.jsGetData,
		"GetDataT":     p.jsGetDataT,
		"Table":        p.jsTable,
		"Exec":         p.jsExec,
		"ExecT":        p.jsExecT,
		"GoExecT":      p.jsGoExecT,
		"QueryOne":     p.jsQueryOne,
		"QStr":         p.jsQStr,
		"Open":         p.jsOpen,
		"Close":        p.jsClose,
		"Begin":        p.jsBegin,
		"Commit":       p.jsCommit,
		"Rollback":     p.jsRollback,
		"SelectLimit":  p.jsSelectLimit,
		"SelectLimitT": p.jsSelectLimitT,
	}
}

func (ahelp *DBHelper) Import(pathName string) error {
	configFileName := filepath.Join(pathName, "config.json")
	defineFileName := filepath.Join(pathName, "define.json")
	dataCsvFileName := filepath.Join(pathName, "data.csv")
	runAtImportFileName := filepath.Join(pathName, "runatimport.js")
	if _, err := os.Stat(configFileName); err == os.ErrNotExist {
		return fmt.Errorf("the dir %q invalid,can't foud the config.json", pathName)
	}
	buf, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return err
	}
	config := dumpConfig{}
	err = json.Unmarshal(buf, &config)
	if err != nil {
		return err
	}
	if config.CheckVersion {
		current_version, err := ahelp.Version(config.CurrentGrade)
		if err != nil {
			return err
		}
		if !config.Version.GE(current_version) {
			return fmt.Errorf("the table's version %v not ge current version %v", config.Version, current_version)
		}
	}
	buf, err = ioutil.ReadFile(runAtImportFileName)
	if os.IsNotExist(err) {
		buf = nil
	} else if err != nil {
		return err
	}

	runAtImport := string(buf)

	buf, err = ioutil.ReadFile(defineFileName)
	if err != nil {
		return err
	}
	tmptable, err := NewDataTableJSON(buf)
	if err != nil {
		return err
	}
	tmpTableName := ""
	{
		b := make([]byte, 8)
		_, err := rand.Read(b)
		if err != nil {
			return err
		}
		tmpTableName = fmt.Sprintf("tmp_%x", b)
	}

	trueTableName := tmptable.TableName

	table := NewDBTable(ahelp, tmptable)
	table.TableName = tmpTableName
	table.Temporary = true
	//create table
	if err := table.UpdateStruct(); err != nil {
		return err
	}
	defer ahelp.DropTable(tmpTableName)
	dataCSV, err := os.Open(dataCsvFileName)
	if err != nil && !os.IsNotExist(err) {
		return err
	} else if err == nil {
		// close fo on exit and check for its returned error
		defer func() {
			if err := dataCSV.Close(); err != nil {
				panic(err)
			}
		}()
		// make a reader buffer
		rCSV := csv.NewReader(dataCSV)
		if err := importFromCsv(pathName, rCSV, table, &config); err != nil {
			return err
		}
	}
	//update true table struct if the ImpRefreshStruct is true
	table.TableName = trueTableName
	table.Temporary = false
	if config.ImpRefreshStruct {
		if err := table.UpdateStruct(); err != nil {
			return err
		}
	}
	//run import sql
	if runAtImport != "" {
		js := fmt.Sprintf("(%s)", runAtImport)
		rm := otto.New()
		jsfun, err := rm.Run(js)
		if err != nil {
			return err
		}
		h, err := rm.ToValue(ahelp.Object())
		if err != nil {
			return err
		}
		_, err = jsfun.Call(jsfun, h, tmpTableName, config.CurrentGrade.String(), config.SqlWhere)
		if err != nil {
			return err
		}
	}
	fmt.Println("merge table", trueTableName, tmpTableName)
	if err := ahelp.Merge(trueTableName, tmpTableName, table.ColumnNames(), table.PK, config.ImpAutoUpdate, config.ImpAutoRemove, config.SqlWhere); err != nil {
		return err
	}
	return nil
}
func importFromCsv(pathName string, rCSV *csv.Reader, table *DBTable, config *dumpConfig) error {
	var colNames []string
	var err error
	colNames, err = rCSV.Read()
	if err != nil && err != io.EOF {
		return err
	}
	if err == nil {
		//map text column index to table column index
		columnIndexes := make([]int, len(colNames))
		for i, v := range colNames {
			bFound := false
			for _, col := range table.Columns {
				if col.Name == v {
					columnIndexes[i] = col.Index()
					bFound = true
					break
				}
			}
			if !bFound {
				return fmt.Errorf("the column %q not exits at table", v)
			}
		}

		fileColumnIndexes := map[string]int{}
		for k, _ := range config.FileColumns {
			fileColumnIndexes[k] = table.ColumnIndex(k)
			if fileColumnIndexes[k] < 0 {
				return fmt.Errorf("the column %q not exits at table", k)
			}
		}

		fileTimeColumnIndexes := map[string]int{}
		for k, _ := range config.FileTimeColumns {
			fileTimeColumnIndexes[k] = table.ColumnIndex(k)
			if fileTimeColumnIndexes[k] < 0 {
				return fmt.Errorf("the column %q not exits at table", k)
			}
		}

		var line []string
		for line, err = rCSV.Read(); err == nil; line, err = rCSV.Read() {
			addValues := make([]interface{}, table.ColumnCount())
			//process the csv data
			for i, v := range line {
				icolIndex := columnIndexes[i]
				tv, err := table.Columns[icolIndex].DecodeString(v)
				if err != nil {
					return err
				}
				addValues[icolIndex] = tv
			}
			//get primary keys
			keys := make([]interface{}, len(table.PK))
			for i, v := range table.PK {
				keys[i] = addValues[table.ColumnIndex(v)]
				if keys[i] == nil {
					return fmt.Errorf("the primary key is null,column name is %q", v)
				}
			}
			//process the file column data
			for i, v := range config.FileColumns {
				icolIndex := fileColumnIndexes[i]
				tv, err := readFileColumn(pathName, table.Columns[icolIndex].Name, v, keys)
				if err != nil && !os.IsNotExist(err) {
					return err
				}
				if tv == nil {
					addValues[icolIndex] = tv

				} else {
					switch table.Columns[icolIndex].DataType {
					case datatable.Bytea:
						addValues[icolIndex] = tv
					case datatable.String:
						addValues[icolIndex] = string(tv)
					default:
						return fmt.Errorf("the column %q 's type %#v invalid", i, table.Columns[icolIndex].DataType)
					}
				}
			}
			//process the file time column data
			for i, flColumnName := range config.FileTimeColumns {
				if ext, ok := config.FileColumns[flColumnName]; ok {
					icolIndex := fileTimeColumnIndexes[i]
					tv, err := readFileTimeColumn(pathName, flColumnName, ext, keys)
					if err != nil {
						return err
					}
					addValues[icolIndex] = tv
				} else {
					return fmt.Errorf("the column %s not is file column", flColumnName)
				}
			}
			if err := table.AddValues(addValues...); err != nil {
				return err
			}
			if table.RowCount() >= ImportBatch {
				if _, err := table.Save(); err != nil {
					return err
				}
				table.Clear()
			}
		}
		if table.RowCount() >= 0 {
			if _, err := table.Save(); err != nil {
				return err
			}
			table.Clear()
		}
		if err != nil && err != io.EOF {
			return fmt.Errorf("file %s line:%#v\n%s", pathName, line, err)
		}
	}
	return nil
}
func readFileColumn(pathName, columnName, ext string, primaryKey []interface{}) ([]byte, error) {
	fileName := filepath.Join(pathName, columnName, primaryKeys2FileName(primaryKey, ext))
	return ioutil.ReadFile(fileName)
}
func readFileTimeColumn(pathName, columnName, ext string, primaryKey []interface{}) (time.Time, error) {
	fileName := filepath.Join(pathName, columnName, primaryKeys2FileName(primaryKey, ext))
	if info, err := os.Stat(fileName); err != nil {
		return time.Time{}, err
	} else {
		return info.ModTime(), nil
	}
}
func primaryKeys2FileName(values []interface{}, ext string) string {
	strs := oftenfun.SafeToStrings(values)
	fileName := strings.Join(strs, ",") + ext
	if len(fileName) > 255 {
		panic(fmt.Errorf("the filename:%q too long", fileName))
	}
	return fileName
}
func writeFileColumn(pathName, columnName, ext string, primaryKey []interface{}, value interface{}) error {

	fileName := filepath.Join(pathName, columnName, primaryKeys2FileName(primaryKey, ext))
	filePath := filepath.Dir(fileName)
	if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
		return err
	}
	var buf []byte
	switch tv := value.(type) {
	case []byte:
		buf = tv
	case string:
		buf = []byte(tv)
	default:
		panic(fmt.Errorf("the type %T invalid", value))
	}
	if err := ioutil.WriteFile(fileName, buf, os.ModePerm); err != nil {
		return err
	}
	return nil
}
func writeFileTimeColumn(pathName, columnName, ext string, primaryKey []interface{}, value interface{}) error {

	fileName := filepath.Join(pathName, columnName, primaryKeys2FileName(primaryKey, ext))
	if _, err := os.Stat(fileName); err != nil {
		return err
	}
	var t time.Time
	switch tv := value.(type) {
	case time.Time:
		t = tv
	case nil:
		t = time.Now()
	default:
		panic(fmt.Errorf("the type %T invalid", value))
	}
	if err := os.Chtimes(fileName, t, t); err != nil {
		return err
	}
	return nil
}

type ExportParam struct {
	TableName        string
	CurrentGrade     Grade
	PathName         string
	FileColumns      map[string]string
	FileTimeColumns  map[string]string
	SqlWhere         string
	ImpAutoUpdate    bool
	ImpAutoRemove    bool
	RunAtImport      string
	ImpRefreshStruct bool
	CheckVersion     bool
}
type dumpConfig struct {
	Version          GradeVersion
	CurrentGrade     Grade
	RowCount         int64
	FileColumns      map[string]string
	FileTimeColumns  map[string]string
	SqlWhere         string
	ImpAutoUpdate    bool
	ImpAutoRemove    bool
	ImpRefreshStruct bool
	CheckVersion     bool
}

func (p *DBHelper) Export(param *ExportParam) error {
	var t *DBTable
	if tab, err := p.Table(param.TableName); err != nil {
		return err
	} else {
		t = NewDBTable(p, tab)
	}

	for c, _ := range param.FileColumns {
		if t.ColumnIndex(c) < 0 {
			return fmt.Errorf("the filecolumns has invalid column:%q", c)
		}
	}
	pathName := param.PathName
	if err := os.MkdirAll(pathName, os.ModePerm); err != nil {
		return err
	}
	//write the runatimport sql
	if param.RunAtImport != "" {
		if err := ioutil.WriteFile(filepath.Join(pathName, "runatimport.js"), []byte(param.RunAtImport), os.ModePerm); err != nil {
			return err
		}
	}
	//write the table struct define
	if bys, err := json.MarshalIndent(t.DataTable, "", "\t"); err != nil {
		return err
	} else {
		err = ioutil.WriteFile(filepath.Join(pathName, "define.json"), bys, os.ModePerm)
		if err != nil {
			return err
		}
	}
	sqlWhere := p.ConvertSql(param.SqlWhere, map[string]interface{}{"CurrentGrade": string(param.CurrentGrade)})
	//write the config.json
	count, err := t.Count(sqlWhere)
	if err != nil {
		return err
	}
	version, err := p.Version(param.CurrentGrade)
	if err != nil {
		return err
	}
	config := dumpConfig{
		Version:          version,
		CurrentGrade:     param.CurrentGrade,
		RowCount:         count,
		FileColumns:      param.FileColumns,
		FileTimeColumns:  param.FileTimeColumns,
		SqlWhere:         sqlWhere,
		ImpAutoUpdate:    param.ImpAutoUpdate,
		ImpAutoRemove:    param.ImpAutoRemove,
		ImpRefreshStruct: param.ImpRefreshStruct,
		CheckVersion:     param.CheckVersion,
	}
	if bys, err := json.MarshalIndent(config, "", "\t"); err != nil {
		return err
	} else {
		err = ioutil.WriteFile(filepath.Join(pathName, "config.json"), bys, os.ModePerm)
		if err != nil {
			return err
		}
	}
	//process the record
	// open output file
	dataCSV, err := os.Create(filepath.Join(pathName, "data.csv"))
	if err != nil {
		return err
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := dataCSV.Close(); err != nil {
			panic(err)
		}
	}()
	// make a write buffer
	wCSV := csv.NewWriter(dataCSV)
	colNames := []string{}
	for _, col := range t.Columns {
		if _, ok := param.FileColumns[col.Name]; !ok {
			colNames = append(colNames, col.Name)
		}
	}
	err = wCSV.Write(colNames)
	if err != nil {
		return err
	}
	stepTable, err := p.StepTable(t.DataTable.DataTable, ExportBatch, t.SelectAllByWhere(sqlWhere))
	if err != nil {
		return err
	}
	defer func() {
		if err := stepTable.Close(); err != nil {
			panic(err)
		}
	}()
	for {
		table, eof, err := stepTable.Step()
		if err != nil {
			return err
		}

		for i := 0; i < table.RowCount(); i++ {
			row := table.GetValues(i)
			line := []string{}
			for colIdx, col := range table.Columns {
				if ext, ok := param.FileColumns[col.Name]; ok {
					if err := writeFileColumn(pathName, col.Name, ext, table.KeyValues(i), row[colIdx]); err != nil {
						return err
					}
				} else if flColumnName, ok := param.FileTimeColumns[col.Name]; ok {
					if ext, ok := param.FileColumns[flColumnName]; ok {
						if err := writeFileTimeColumn(pathName, flColumnName, ext, table.KeyValues(i), row[colIdx]); err != nil {
							return err
						}
					} else {
						return fmt.Errorf("the column %s not is file column", flColumnName)
					}
				} else {
					line = append(line, col.EncodeString(row[colIdx]))
				}
			}
			if err := wCSV.Write(line); err != nil {
				return err
			}
		}
		if eof {
			break
		}
	}
	wCSV.Flush()
	if err = wCSV.Error(); err != nil {
		return err
	}
	return nil
}
func (p *DBHelper) Version(gradestr Grade) (GradeVersion, error) {
	var tab *DataTable
	if exists, err := p.TableExists("lx_version"); err != nil {
		return nil, err
	} else if exists {
		if gradestr == GRADE_TAG {
			tab, err = p.GetData("select verno from lx_version")
		} else {
			tab, err = p.GetData("select verno from lx_version where {{ph}} like concat(grade,'%')", gradestr.String())
		}
		if err != nil {
			return nil, err
		}
		rev := make([]string, tab.RowCount())
		for i, v := range tab.GetColumnValues(0) {
			rev[i] = fmt.Sprintf("%d", v)
		}
		return ParseGradeVersion(strings.Join(rev, "."))
	} else {
		return GradeVersion{}, nil
	}
}
