package grade

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/linlexing/pghelper"
	"github.com/robertkrimen/otto"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
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
	t := NewDBTable(p, NewDataTableT(tab.DataTable))
	return t, nil
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
func (ahelp *PGHelper) Import(pathName string) error {
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
	if err != nil {
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

	runAtImportSql, err = tmplExecute(runAtImportSql, map[string]interface{}{
		"TempTableName": tmpTableName,
		"ExportGrade":   string(config.CurrentGrade),
		"SqlWhere":      config.SqlWhere,
	})
	if err != nil {
		return err
	}
	table := NewDBTable(ahelp, tmptable)
	table.TableName = tmpTableName
	table.Temp = true
	//create table
	if err := table.UpdateStruct(); err != nil {
		return err
	}
	defer ahelp.DropTable(tmpTableName)
	dataCSV, err := os.Open(dataCsvFileName)
	if err != nil {
		return err
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := dataCSV.Close(); err != nil {
			panic(err)
		}
	}()
	// make a reader buffer
	rCSV := csv.NewReader(dataCSV)
	var colNames []string
	colNames, err = rCSV.Read()
	if err != nil && err != io.EOF {
		return err
	}
	if err == nil {
		//map text column index to table column index
		columnIndexes := make([]int, len(colNames))
		fileColumnIndexes := map[string]int{}
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
		for k, _ := range config.FileColumns {
			fileColumnIndexes[k] = table.ColumnIndex(k)
			if fileColumnIndexes[k] < 0 {
				return fmt.Errorf("the column %q not exits at table", k)
			}
		}

		var line []string
		for line, err = rCSV.Read(); err == nil; line, err = rCSV.Read() {
			addValues := make([]interface{}, table.ColumnCount())
			//process the csv data
			for i, v := range line {
				icolIndex := columnIndexes[i]
				tv, err := table.Columns[icolIndex].Parse(v)
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
					switch table.Columns[icolIndex].PGType.Type {
					case pghelper.TypeBytea:
						addValues[icolIndex] = tv
					case pghelper.TypeString:
						addValues[icolIndex] = string(tv)
					default:
						return fmt.Errorf("the column %q 's type %#v invalid", i, table.Columns[icolIndex].PGType)
					}
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
	//update true table struct if the ImpRefreshStruct is true
	table.TableName = trueTableName
	table.Temp = false
	if config.ImpRefreshStruct {
		if err := table.UpdateStruct(); err != nil {
			return err
		}
	}
	//run import sql
	if err := ahelp.ExecuteSql(runAtImportSql); err != nil {
		return err
	}
	if err := ahelp.Merge(trueTableName, tmpTableName, table.ColumnNames(), table.PK, config.ImpAutoRemove, config.SqlWhere); err != nil {
		return err
	}
	return nil
}

func readFileColumn(pathName, columnName, ext string, primaryKey []interface{}) ([]byte, error) {
	fileName := filepath.Join(pathName, columnName, primaryKeys2FileName(primaryKey, ext))
	return ioutil.ReadFile(fileName)
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

type ExportParam struct {
	TableName        string
	CurrentGrade     Grade
	PathName         string
	FileColumns      map[string]string
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
	SqlWhere         string
	ImpAutoRemove    bool
	ImpRefreshStruct bool
	CheckVersion     bool
}

func tmplExecute(tmplSrc string, param interface{}) (string, error) {
	tmpl := template.New("expimp_template")
	tmpl.Delims("<#", "#>")
	tmpl.Funcs(template.FuncMap{
		"PGSignStr": pghelper.PGSignStr,
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
func (p *PGHelper) Export(param *ExportParam) error {
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
	if err := ioutil.WriteFile(filepath.Join(pathName, "runatimport.sql"), []byte(param.SqlRunAtImport), os.ModePerm); err != nil {
		return err
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
	sqlWhere, err := tmplExecute(param.SqlWhere, map[string]interface{}{"CurrentGrade": string(param.CurrentGrade)})
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
	err = t.BatchFillWhere(func(table *DBTable, eof bool) error {
		for i := 0; i < table.RowCount(); i++ {
			row := table.GetValues(i)
			line := []string{}
			for colIdx, col := range table.Columns {
				if ext, ok := param.FileColumns[col.Name]; ok {
					if err := writeFileColumn(pathName, col.Name, ext, table.KeyValues(i), row[colIdx]); err != nil {
						return err
					}

				} else {
					if str, err := col.String(row[colIdx]); err != nil {
						return err
					} else {
						line = append(line, str)
					}

				}
			}
			if err := wCSV.Write(line); err != nil {
				return err
			}
		}
		return nil
	}, ExportBatch, sqlWhere)
	if err != nil {
		return err
	}
	wCSV.Flush()
	if err = wCSV.Error(); err != nil {
		return err
	}
	return nil
}
func (p *PGHelper) Version(gradestr Grade) (GradeVersion, error) {
	str, err := p.GetString("select string_agg(cast(verno as text),'.') from lx_version where $1 like grade||'%'", string(gradestr))
	if err != nil {
		return nil, err
	}
	return ParseGradeVersion(str)
}
