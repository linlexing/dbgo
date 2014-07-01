package grade

import (
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

type dumpConfig struct {
	CurrentGrade Grade
	RowCount     int64
	FileColumns  []struct{ ColumnName, FileExt string }
}

func (t *DBTable) Export(currentGrade Grade, pathName string, fileColumns []struct{ ColumnName, FileExt string }, sqlRunAtImport string, sqlwhere string, params ...interface{}) error {
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

	for _, c := range fileColumns {
		if t.ColumnIndex(c.ColumnName) < 0 {
			return fmt.Errorf("the filecolumns has invalid column:%q", c.ColumnName)
		}
	}
	pathName = filepath.Join(pathName, t.TableName)
	if err := os.MkdirAll(pathName, os.ModePerm); err != nil {
		return err
	}
	//write the runatimport sql
	if err := ioutil.WriteFile(filepath.Join(pathName, "runatimport.sql"), []byte(sqlRunAtImport), os.ModePerm); err != nil {
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
	//write the config.json
	count, err := t.Count(sqlwhere, params...)
	if err != nil {
		return err
	}

	config := dumpConfig{
		CurrentGrade: currentGrade,
		RowCount:     count,
		FileColumns:  fileColumns,
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
		bFound := false
		for _, fileColumn := range fileColumns {
			if fileColumn.ColumnName == col.Name {
				bFound = true
				break
			}
		}
		if !bFound {
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
				bFound := false
				for _, fileColumn := range fileColumns {
					if fileColumn.ColumnName == col.Name {
						if err := writeFileColumn(pathName, col.Name, fileColumn.FileExt, table.KeyValues(i), row[colIdx]); err != nil {
							return err
						}
						bFound = true
						break
					}
				}
				if !bFound {
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
	}, ExportBatch, sqlwhere, params...)
	if err != nil {
		return err
	}
	wCSV.Flush()
	if err = wCSV.Error(); err != nil {
		return err
	}
	return nil
}
func Import(ahelp *PGHelper, pathName string) error {
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

	runAtImportSql = strings.Replace(runAtImportSql, "<#.TempTableName#>", tmpTableName, -1)
	runAtImportSql = strings.Replace(runAtImportSql, "<#.Grade#>", string(config.CurrentGrade), -1)

	table := NewDBTable(ahelp, tmptable)
	table.TableName = tmpTableName
	table.Temp = true
	//create table
	if err := table.UpdateStruct(); err != nil {
		return err
	}
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
		fileColumnIndexes := make([]int, len(config.FileColumns))
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
		for i, v := range config.FileColumns {
			bFound := false
			for _, col := range table.Columns {
				if col.Name == v.ColumnName {
					fileColumnIndexes[i] = col.Index()
					bFound = true
					break
				}
			}
			if !bFound {
				return fmt.Errorf("the column %q not exits at table", v)
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
				tv, err := readFileColumn(pathName, table.Columns[icolIndex].Name, v.FileExt, keys)
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
						return fmt.Errorf("the column %q 's type %#v invalid", v.ColumnName, table.Columns[icolIndex].PGType)
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
		if table.RowCount() >= ImportBatch {
			if _, err := table.Save(); err != nil {
				return err
			}
			table.Clear()
		}
		if err != nil && err != io.EOF {
			return err
		}
	}
	//update true table struct
	table.TableName = trueTableName
	table.Temp = false
	if err := table.UpdateStruct(); err != nil {
		return err
	}
	//run import sql
	if err := ahelp.ExecuteSql(runAtImportSql); err != nil {
		return err
	}
	if err := ahelp.Merge(trueTableName, tmpTableName, table.DataTable); err != nil {
		return err
	}
	return nil
}

func readFileColumn(pathName, columnName, ext string, primaryKey []interface{}) ([]byte, error) {
	fileName := primaryKeys2FileName(primaryKey, ext)
	filePath := filepath.Join(pathName, columnName)
	return ioutil.ReadFile(filepath.Join(filePath, fileName))
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
	fileName := primaryKeys2FileName(primaryKey, ext)
	filePath := filepath.Join(pathName, columnName)
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
	if err := ioutil.WriteFile(filepath.Join(filePath, fileName), buf, os.ModePerm); err != nil {
		return err
	}
	return nil
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
