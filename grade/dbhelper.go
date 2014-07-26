package grade

import (
	"bytes"
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
	"text/template"
	"time"
)

type DBHelper struct {
	*dbhelper.DBHelper
}

func NewDBHelper(dbDriver, dburl string) *DBHelper {
	return NewDBHelperT(dbhelper.NewDBHelper(dbDriver, dburl))
}
func NewDBHelperT(ahelp *dbhelper.DBHelper) *DBHelper {
	return &DBHelper{ahelp}
}
func (p *DBHelper) GetData(strSql string, params ...interface{}) (*DataTable, error) {
	tab, err := p.DBHelper.GetData(strSql, params...)
	if err != nil {
		return nil, err
	}
	return NewDataTableT(tab), nil

}
func (p *DBHelper) Table(tablename string) (*DBTable, error) {
	tab, err := p.DBHelper.Table(tablename)
	if err != nil {
		return nil, err
	}
	t := NewDBTable(p, NewDataTableT(tab))
	return t, nil
}
func (p *DBHelper) UpdateStruct(newStruct *DataTable) error {
	var oldStruct *DBTable
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
		return p.DBHelper.UpdateStruct(nil, newStruct.DataTable)
	}
	trueOld, ok := oldStruct.Reduced(newStruct.Grade())
	if !ok {
		return fmt.Errorf("the oldStruct's grade is %q,newStruct can't use it", oldStruct.DataTable.Grade())
	}

	return p.DBHelper.UpdateStruct(trueOld.DataTable, newStruct.DataTable)

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
	params := make([]interface{}, len(call.ArgumentList)-1)
	for i := 1; i < len(call.ArgumentList); i++ {
		var err error
		params[i-1], err = call.ArgumentList[i].Export()
		if err != nil {
			panic(err)
		}
	}
	result, err := p.GetData(strSql, params...)
	if err != nil {
		panic(err)
	}
	return oftenfun.JSToValue(call.Otto, result.Object())
}
func (p *DBHelper) Object() map[string]interface{} {
	return map[string]interface{}{
		"GetData": p.jsGetData,
		"Table":   p.jsTable,
	}
}

func (ahelp *DBHelper) Import(pathName string) error {
	configFileName := filepath.Join(pathName, "config.json")
	defineFileName := filepath.Join(pathName, "define.json")
	dataCsvFileName := filepath.Join(pathName, "data.csv")
	runAtImportFileName := filepath.Join(pathName, "runatimport.sql")
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

	runAtImportSql := string(buf)

	buf, err = ioutil.ReadFile(defineFileName)
	if err != nil {
		return err
	}
	tmptable, err := NewDataTableJSON(buf)
	if err != nil {
		return err
	}
	tmpTableName := "import_tmp_lx"
	trueTableName := tmptable.TableName

	runAtImportSql, err = ahelp.tmplExecute(runAtImportSql, map[string]interface{}{
		"TempTableName": tmpTableName,
		"ExportGrade":   string(config.CurrentGrade),
		"SqlWhere":      config.SqlWhere,
	})
	if err != nil {
		return err
	}
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
	if runAtImportSql != "" {
		if _, err := ahelp.Exec(runAtImportSql); err != nil {
			return err
		}
	}
	if err := ahelp.Merge(trueTableName, tmpTableName, table.ColumnNames(), table.PK, config.ImpAutoRemove, config.SqlWhere); err != nil {
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
				if err != nil {
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
			return err
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
	ImpAutoRemove    bool
	SqlRunAtImport   string
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
	ImpAutoRemove    bool
	ImpRefreshStruct bool
	CheckVersion     bool
}

func (d *DBHelper) tmplExecute(tmplSrc string, param interface{}) (string, error) {
	tmpl := template.New("expimp_template")
	tmpl.Delims("<#", "#>")
	tmpl.Funcs(template.FuncMap{
		"StringExpress": d.StringExpress,
	})

	tmpl, err := tmpl.Parse(tmplSrc)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	if err = tmpl.Execute(buf, param); err != nil {
		return "", err
	}
	return buf.String(), nil
}
func (p *DBHelper) Export(param *ExportParam) error {
	const Import_Desc = "/*************************************************************/\n" +
		"/*When the data imported temporary table, run the import SQL,*/\n" +
		"/*finally merge data into actual table from temporary table. */\n" +
		"/*The sql can use  following macro:                          */\n" +
		"/*                                                           */\n" +
		"/*<#.TempTableName#>  The temp table name                    */\n" +
		"/*<#.Grade#>          The export table's grade               */\n" +
		"/*************************************************************/\n"
	const Struct_Desc = "/*************************************************************/\n" +
		"/*The DataType can use following value:                      */\n" +
		"/*                                                           */\n" +
		"/* 0  TypeString                                             */\n" +
		"/* 1  TypeBool                                               */\n" +
		"/* 2  TypeInt64                                              */\n" +
		"/* 3  TypeFloat64                                            */\n" +
		"/* 4  TypeTime                                               */\n" +
		"/* 5  TypeBytea                                              */\n" +
		"/* 6  TypeStringSlice                                        */\n" +
		"/* 7  TypeBoolSlice                                          */\n" +
		"/* 8  TypeInt64Slice                                         */\n" +
		"/* 9  TypeFloat64Slice                                       */\n" +
		"/* 10 TypeTimeSlice                                          */\n" +
		"/* 11 TypeJSON                                               */\n" +
		"/* 12 TypeJSONSlice                                          */\n" +
		"/*************************************************************/\n"
	t, err := p.Table(param.TableName)
	if err != nil {
		return err
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
	if param.SqlRunAtImport != "" {
		if err := ioutil.WriteFile(filepath.Join(pathName, "runatimport.sql"), []byte(param.SqlRunAtImport), os.ModePerm); err != nil {
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
	sqlWhere, err := p.tmplExecute(param.SqlWhere, map[string]interface{}{"CurrentGrade": string(param.CurrentGrade)})
	if err != nil {
		return err
	}
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
	var err error
	if gradestr == GRADE_TAG {
		tab, err = p.GetData("select verno from lx_version")
	} else {
		tab, err = p.GetData("select verno from lx_version where $1 like concat(grade,'%')", gradestr.String())
	}
	if err != nil {
		return nil, err
	}
	rev := make([]string, tab.RowCount())
	for i, v := range tab.GetColumnValues(0) {
		rev[i] = fmt.Sprintf("%d", v)
	}
	return ParseGradeVersion(strings.Join(rev, "."))
}
