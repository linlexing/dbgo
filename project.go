package main

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"fmt"
	"github.com/linlexing/dbgo/grade"
	"github.com/linlexing/dbgo/log"
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/linlexing/pghelper"
	"github.com/robertkrimen/otto"
	"html/template"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

const (
	BEFORE int64 = iota

	//AFTER
	//PANIC
	//FINALLY

	NotFoundAction = "NotFoundAction"
)

type Intercept struct {
	ID            string
	InterceptWhen int64
	Script        string
	Grade         string
}

type Project interface {
	Name() string

	DBHelper() *grade.PGHelper
	DefaultAction() (string, string)
	ReverseUrl(args ...string) string
	ClearCache()
	ReloadRepository() error
	GetPackageNames(dirName string, gradestr grade.Grade) ([]string, error)
	Require(rm *otto.Otto, fileName, currentModuleDir string, gradestr grade.Grade) (*otto.Script, string, error)

	Version(grade grade.Grade) (int64, bool, error)
	Model(mname string, gradestr grade.Grade) *grade.DBTable
	Table(mname string, gradestr grade.Grade) *grade.DBTable
	ExportData(dumpName string, expFile *os.File, gradestr grade.Grade) error
	ImportData(impPath string) error
	Checks(tablename string, gradestr grade.Grade) ([]*Check, error)
	TemplateSet(f template.FuncMap) (*template.Template, error)
	Object() map[string]interface{}
}

type project struct {
	name        string
	dbHelper    *grade.PGHelper
	metaProject string
	repository  string

	lockTableDefine *sync.Mutex
	lockVersion     *sync.Mutex
	lockTemplateSet *sync.Mutex
	lockCheck       *sync.Mutex
	lockPackage     *sync.Mutex

	cachePackage map[struct {
		Name  string
		Grade grade.Grade
	}]*otto.Script
	cacheVersion     map[grade.Grade]int64 //每级grade均对应一个最新的版本
	cacheTableDefine map[string]*grade.DataTable
	cacheTemplateSet *template.Template
	cacheCheck       map[string][]*Check
}
type Action struct {
	ID     string
	Script string
	Grade  string
}

func lx_dump() *grade.DataTable {
	table := grade.NewDataTable("lx_dump", grade.GRADE_ROOT)
	table.AddColumn(grade.NewColumn("name", grade.GRADE_ROOT, pghelper.TypeString, true))
	table.AddColumn(grade.NewColumnT("id", grade.GRADE_ROOT, pghelper.NewPGType(pghelper.TypeInt64, 0, true), "0"))
	table.AddColumn(grade.NewColumn("grade", grade.GRADE_ROOT, pghelper.TypeString, true))
	table.AddColumn(grade.NewColumnT("tablename", grade.GRADE_ROOT, pghelper.NewPGType(pghelper.TypeString, 64, true), ""))
	table.AddColumn(grade.NewColumn("sqlwhere", grade.GRADE_ROOT, pghelper.TypeString, false))
	table.AddColumn(grade.NewColumn("filecolumns", grade.GRADE_ROOT, pghelper.TypeJSON, false))
	table.AddColumn(grade.NewColumn("impautoremove", grade.GRADE_ROOT, pghelper.TypeBool, false))
	table.AddColumn(grade.NewColumn("sqlrunatimport", grade.GRADE_ROOT, pghelper.TypeString, false))
	table.AddColumn(grade.NewColumn("imprefreshstruct", grade.GRADE_ROOT, pghelper.TypeBool, false))
	table.AddColumn(grade.NewColumn("checkversion", grade.GRADE_ROOT, pghelper.TypeBool, false))
	table.SetPK("name", "id")
	return table
}

func lx_version() *grade.DataTable {
	table := grade.NewDataTable("lx_version", grade.GRADE_ROOT)
	table.AddColumn(grade.NewColumn("grade", grade.GRADE_ROOT, pghelper.TypeString, true))
	table.AddColumn(grade.NewColumn("verno", grade.GRADE_ROOT, pghelper.TypeInt64, true))
	table.AddColumn(grade.NewColumn("verlabel", grade.GRADE_ROOT, pghelper.TypeString, false))
	table.AddColumn(grade.NewColumn("changelog", grade.GRADE_ROOT, pghelper.TypeString, false))
	table.AddColumn(grade.NewColumn("releasetime", grade.GRADE_ROOT, pghelper.TypeTime, false))
	table.SetPK("grade", "verno")
	return table
}

func lx_view() *grade.DataTable {
	table := grade.NewDataTable("lx_view", grade.GRADE_ROOT)
	table.AddColumn(grade.NewColumnT("name", grade.GRADE_ROOT, pghelper.NewPGType(pghelper.TypeString, 0, true), ""))
	table.AddColumn(grade.NewColumnT("grade", grade.GRADE_ROOT, pghelper.NewPGType(pghelper.TypeString, 0, true), ""))
	table.AddColumn(grade.NewColumnT("content", grade.GRADE_ROOT, pghelper.NewPGType(pghelper.TypeString, 0, false), ""))
	table.SetPK("name")
	return table
}

func lx_checkaddition() *grade.DataTable {

	table := grade.NewDataTable("lx_checkaddition", grade.GRADE_ROOT)
	table.AddColumn(grade.NewColumnT("tablename", grade.GRADE_ROOT, pghelper.NewPGType(pghelper.TypeString, 64, true), ""))
	table.AddColumn(grade.NewColumnT("addition", grade.GRADE_ROOT, pghelper.NewPGType(pghelper.TypeString, 64, true), ""))
	table.AddColumn(grade.NewColumnT("fields", grade.GRADE_ROOT, pghelper.NewPGType(pghelper.TypeStringSlice, 0, false), ""))
	table.AddColumn(grade.NewColumnT("runatserver", grade.GRADE_ROOT, pghelper.NewPGType(pghelper.TypeBool, 0, false), ""))
	table.AddColumn(grade.NewColumnT("script", grade.GRADE_ROOT, pghelper.NewPGType(pghelper.TypeString, 0, false), ""))
	table.AddColumn(grade.NewColumnT("sqlwhere", grade.GRADE_ROOT, pghelper.NewPGType(pghelper.TypeString, 0, false), ""))
	table.SetPK("tablename", "addition")
	return table
}
func lx_check() *grade.DataTable {

	table := grade.NewDataTable("lx_check", grade.GRADE_ROOT)
	table.AddColumn(grade.NewColumnT("tablename", grade.GRADE_ROOT, pghelper.NewPGType(pghelper.TypeString, 64, true), ""))
	table.AddColumn(grade.NewColumnT("id", grade.GRADE_ROOT, pghelper.NewPGType(pghelper.TypeInt64, 0, true), ""))
	table.AddColumn(grade.NewColumnT("displaylabel", grade.GRADE_ROOT, pghelper.NewPGType(pghelper.TypeString, 0, true), ""))
	table.AddColumn(grade.NewColumnT("level", grade.GRADE_ROOT, pghelper.NewPGType(pghelper.TypeInt64, 0, true), ""))
	table.AddColumn(grade.NewColumnT("fields", grade.GRADE_ROOT, pghelper.NewPGType(pghelper.TypeStringSlice, 0, false), ""))
	table.AddColumn(grade.NewColumnT("runatserver", grade.GRADE_ROOT, pghelper.NewPGType(pghelper.TypeBool, 0, false), ""))
	table.AddColumn(grade.NewColumnT("addition", grade.GRADE_ROOT, pghelper.NewPGType(pghelper.TypeString, 64, false), ""))
	table.AddColumn(grade.NewColumnT("script", grade.GRADE_ROOT, pghelper.NewPGType(pghelper.TypeString, 0, false), ""))
	table.AddColumn(grade.NewColumnT("sqlwhere", grade.GRADE_ROOT, pghelper.NewPGType(pghelper.TypeString, 0, false), ""))
	table.AddColumn(grade.NewColumnT("grade", grade.GRADE_ROOT, pghelper.NewPGType(pghelper.TypeString, 0, true), ""))
	table.SetPK("tablename", "id")
	return table
}
func RepositoryPath(repo string) string {
	return filepath.Join(AppPath, "repository", repo, "_pd")
}
func publicTables() []*grade.DataTable {
	return []*grade.DataTable{lx_dump(), lx_version(), lx_view(), lx_checkaddition(), lx_check()}

}
func NewProject(name, dburl, repository string) (Project, error) {
	p := &project{
		name:       name,
		dbHelper:   grade.NewPGHelper(dburl),
		repository: repository,

		lockTableDefine: &sync.Mutex{},
		lockVersion:     &sync.Mutex{},
		lockTemplateSet: &sync.Mutex{},
		lockCheck:       &sync.Mutex{},
		lockPackage:     &sync.Mutex{},

		cacheVersion:     map[grade.Grade]int64{},
		cacheTableDefine: map[string]*grade.DataTable{},
		cacheTemplateSet: nil,
		cacheCheck:       map[string][]*Check{},
		cachePackage: map[struct {
			Name  string
			Grade grade.Grade
		}]*otto.Script{},
	}

	//先确保有公共的表
	for _, v := range publicTables() {
		if err := p.dbHelper.UpdateStruct(v); err != nil {
			return nil, err
		}
	}
	return p, nil
}
func (p *project) Model(mname string, gradestr grade.Grade) *grade.DBTable {
	return p.Table(mname, gradestr)
}

func (p *project) loadVersion() (map[grade.Grade]int64, error) {
	tab, err := p.dbHelper.GetDataTable(SQL_GerVersion)
	if err != nil {
		return nil, err
	}
	rev := map[grade.Grade]int64{}
	for i := 0; i < tab.RowCount(); i++ {
		row := tab.GetRow(i)
		rev[grade.Grade(row["grade"].(string))] = row["verno"].(int64)
	}
	return rev, nil
}
func (p *project) ReverseUrl(args ...string) string {
	return path.Join(append([]string{"/", p.Name()}, args...)...)
}

//Check the project template is ready, if not ready then loaded
func (p *project) loadTemplate(f template.FuncMap) (*template.Template, error) {
	rev := template.New("")
	rev.Delims("<#", "#>")
	rev.Funcs(template.FuncMap{
		"static": func(filename string) string {
			return p.ReverseUrl("static", "file", filename)
		},
		"title": func(title, urlstr string) string {
			return fmt.Sprintf("%s?title=%s", urlstr, url.QueryEscape(title))
		},
		"JS": func(src string) template.JS {
			return template.JS(src)
		},
		"PathJoin": path.Join,
		/*"GradeCanUse": func(g1, g2 string) bool {
			return grade.Grade(g1).CanUse(grade.Grade(g2))
		},*/
		"tmpl": func(tmpl string, data map[string]interface{}) template.HTML {
			buf := &bytes.Buffer{}
			rev.ExecuteTemplate(buf, tmpl, data)
			return template.HTML(buf.String())
		},
		"css": func(tmpl string, data map[string]interface{}) template.CSS {
			buf := &bytes.Buffer{}
			rev.ExecuteTemplate(buf, tmpl, data)
			return template.CSS(buf.String())
		},
		"set": func(renderArgs map[string]interface{}, key string, value interface{}) template.HTML {
			renderArgs[key] = value
			return template.HTML("")
		},
		"append": func(renderArgs map[string]interface{}, key string, value interface{}) template.HTML {
			if renderArgs[key] == nil {
				renderArgs[key] = []interface{}{value}
			} else {
				renderArgs[key] = append(renderArgs[key].([]interface{}), value)
			}
			return template.HTML("")
		},
	})
	rev.Funcs(f)
	//get template
	tab, err := p.dbHelper.GetDataTable(SQL_GetView)
	if err != nil {
		return nil, err
	}
	for i := 0; i < tab.RowCount(); i++ {
		row := tab.GetRow(i)
		content := fmt.Sprintf(`
			<#if .c.CurrentGrade.CanUse %q#>
			%s
			<#end#>`, row["grade"], row["content"])
		_, err := rev.New(row["name"].(string)).Parse(content)
		if err != nil {
			return nil, err
		}
	}

	return rev, nil
}
func (p *project) Name() string {
	return p.name
}

func (p *project) Version(gradestr grade.Grade) (int64, bool, error) {
	p.lockVersion.Lock()
	defer p.lockVersion.Unlock()
	if len(p.cacheVersion) == 0 {
		ver, err := p.loadVersion()
		if err != nil {
			return 0, false, err
		}
		p.cacheVersion = ver
	}
	v, ok := p.cacheVersion[gradestr]
	return v, ok, nil
}

func (p *project) DBHelper() *grade.PGHelper {
	return p.dbHelper
}
func (p *project) DefaultAction() (string, string) {
	return "login", "login"
}

func (p *project) TemplateSet(f template.FuncMap) (*template.Template, error) {
	p.lockTemplateSet.Lock()
	defer p.lockTemplateSet.Unlock()
	if p.cacheTemplateSet == nil {
		t, err := p.loadTemplate(f)
		if err != nil {
			return nil, err
		}
		p.cacheTemplateSet = t
	} else {
		p.cacheTemplateSet.Funcs(f)
	}

	return p.cacheTemplateSet, nil
}

func (p *project) jsReverseUrl(call otto.FunctionCall) otto.Value {
	params := make([]string, len(call.ArgumentList))
	for i, v := range call.ArgumentList {
		params[i] = v.String()
	}
	v, _ := otto.ToValue(p.ReverseUrl(params...))
	return v
}
func (p *project) jsModel(call otto.FunctionCall) otto.Value {
	tablename := oftenfun.AssertString(call.Argument(0))
	gradestr := grade.Grade(oftenfun.AssertString(call.Argument(1)))
	return oftenfun.JSToValue(call.Otto, p.Model(tablename, gradestr).Object())
}
func (p *project) jsChecks(call otto.FunctionCall) otto.Value {
	tablename := oftenfun.AssertString(call.Argument(0))
	gradestr := grade.Grade(oftenfun.AssertString(call.Argument(1)))
	chks, err := p.Checks(tablename, gradestr)
	if err != nil {
		panic(err)
	}
	return oftenfun.JSToValue(call.Otto, chks)
}
func (p *project) jsReloadRepository(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, p.ReloadRepository())
}

func (p *project) Object() map[string]interface{} {
	return map[string]interface{}{
		"ReverseUrl":       p.jsReverseUrl,
		"Name":             p.Name(),
		"ClearCache":       p.jsClearCache,
		"DBHelper":         p.DBHelper().Object(),
		"Model":            p.jsModel,
		"Checks":           p.jsChecks,
		"ReloadRepository": p.jsReloadRepository,
	}
}
func (p *project) jsClearCache(call otto.FunctionCall) otto.Value {
	p.ClearCache()
	return otto.NullValue()
}
func (p *project) ClearCheckCache() {
	p.lockCheck.Lock()
	defer p.lockCheck.Unlock()
	p.cacheCheck = map[string][]*Check{}
	return
}
func (p *project) ClearTableDefineCache() {
	p.lockTableDefine.Lock()
	defer p.lockTableDefine.Unlock()
	p.cacheTableDefine = map[string]*grade.DataTable{}
	return
}
func (p *project) ClearTemplateSetCache() {
	p.lockTemplateSet.Lock()
	defer p.lockTemplateSet.Unlock()
	p.cacheTemplateSet = nil
	return
}
func (p *project) ClearVersionCache() {
	p.lockVersion.Lock()
	defer p.lockVersion.Unlock()
	p.cacheVersion = map[grade.Grade]int64{}
	return
}
func (p *project) ClearPackageCache() {
	p.lockPackage.Lock()
	defer p.lockPackage.Unlock()
	p.cachePackage = map[struct {
		Name  string
		Grade grade.Grade
	}]*otto.Script{}
	return
}
func (p *project) ClearCache() {
	log.INFO.Printf("project %s clear cache.", p.Name())
	p.ClearCheckCache()
	p.ClearPackageCache()
	p.ClearTableDefineCache()
	p.ClearTemplateSetCache()
	p.ClearVersionCache()
}

func (p *project) getTableDefine(tablename string) (result *grade.DataTable, err error) {
	p.lockTableDefine.Lock()
	defer p.lockTableDefine.Unlock()
	var ok bool
	if result, ok = p.cacheTableDefine[tablename]; !ok {
		var tab *grade.DBTable
		tab, err = p.DBHelper().Table(tablename)
		if err != nil {
			return
		}
		result = tab.DataTable
		p.cacheTableDefine[tablename] = result
	}
	return
}
func (p *project) Table(mname string, gradestr grade.Grade) *grade.DBTable {
	tab, err := p.getTableDefine(mname)
	if err != nil {
		panic(err)
	}
	result, ok := tab.Reduced(gradestr)
	if !ok {
		panic(fmt.Errorf("the table %q not exits at grade:%q", mname, gradestr))
	}
	return grade.NewDBTable(p.DBHelper(), result)
}
func termCat(str1, str2 string) string {
	if len(str1) == 0 {
		return str2
	}
	if len(str2) == 0 {
		return str1
	}
	return "(" + str1 + ") and (" + str2 + ")"
}
func (p *project) loadCheck(tablename string) ([]*Check, error) {
	rev := []*Check{}
	err := p.DBHelper().Query(func(rows *sql.Rows) error {
		for rows.Next() {
			var id, level int64
			var displaylabel, grade string
			var runAtServer pghelper.NullBool
			var script, sqlWhere pghelper.NullString
			var fields pghelper.NullStringSlice
			//a.id,
			//a.displaylabel,
			//a.level,
			//a.fields||b.fields as fields,
			//a.runatserver or b.runatserver as runatserver,
			//array_to_string(array['('||b.script||')','('||a.script||')'],'&&') as script,
			//array_to_string(array['('||b.sqlwhere||')','('||a.sqlwhere||')'],' AND ') as sqlwhere,
			//a.grade
			if err := rows.Scan(&id, &displaylabel, &level, &fields, &runAtServer, &script, &sqlWhere, &grade); err != nil {
				return err
			}
			c := &Check{
				ID:           id,
				DisplayLabel: displaylabel,
				Level:        level,
				Fields:       []string{},
				RunAtServer:  false,
				Script:       "",
				SqlWhere:     "",
				Grade:        grade,
			}
			if fields.Valid {
				c.Fields = fields.Slice
			}
			if runAtServer.Valid {
				c.RunAtServer = runAtServer.Bool
			}
			if script.Valid {
				c.Script = script.String
			}
			if sqlWhere.Valid {
				c.SqlWhere = sqlWhere.String
			}
			rev = append(rev, c)
		}
		return nil
	}, SQL_GetCheck, tablename)
	if err != nil {
		return nil, err
	}
	return rev, nil
}
func (p *project) Checks(tablename string, gradestr grade.Grade) ([]*Check, error) {
	p.lockCheck.Lock()
	defer p.lockCheck.Unlock()
	if _, ok := p.cacheCheck[tablename]; !ok {
		chk, err := p.loadCheck(tablename)
		if err != nil {
			return nil, err
		}
		p.cacheCheck[tablename] = chk
	}
	rev := []*Check{}
	for _, chk := range p.cacheCheck[tablename] {
		if gradestr.CanUse(chk.Grade) {
			rev = append(rev, chk)
		}
	}
	return rev, nil
}
func (p *project) ExportData(dumpName string, expFile *os.File, gradestr grade.Grade) error {
	dumpData := p.Model("lx_dump", gradestr)
	if err := dumpData.FillWhere("name=$1", dumpName); err != nil {
		return err
	}
	if dumpData.RowCount() == 0 {
		return fmt.Errorf("lx_dump can't find record,the name is %q", dumpName)
	}
	tmpDir, err := ioutil.TempDir("", "dbgo_exp")
	if err != nil {
		return err
	}
	defer func() {
		os.RemoveAll(tmpDir)
	}()
	for i := 0; i < dumpData.RowCount(); i++ {
		row := dumpData.GetRow(i)
		fileColumns := map[string]string{}
		for k, v := range row["filecolumns"].(pghelper.JSON) {
			fileColumns[k] = oftenfun.SafeToString(v)
		}
		param := &grade.ExportParam{
			TableName:        row["tablename"].(string),
			CurrentGrade:     gradestr,
			PathName:         filepath.Join(tmpDir, row["tablename"].(string)),
			FileColumns:      fileColumns,
			SqlWhere:         oftenfun.SafeToString(row["sqlwhere"]),
			ImpAutoRemove:    oftenfun.SafeToBool(row["impautoremove"]),
			SqlRunAtImport:   oftenfun.SafeToString(row["sqlrunatimport"]),
			ImpRefreshStruct: oftenfun.SafeToBool(row["imprefreshstruct"]),
			CheckVersion:     oftenfun.SafeToBool(row["checkversion"]),
		}
		if err := p.DBHelper().Export(param); err != nil {
			return err
		}
	}
	if err := zipDir(tmpDir, expFile); err != nil {
		return err
	}
	return nil
}
func zipDir(src string, dest *os.File) error {
	destW := zip.NewWriter(dest)
	filepath.Walk(src, func(filename string, info os.FileInfo, orgerr error) error {
		if filename == src {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		trueZipFileName, err := filepath.Rel(src, filename)
		if err != nil {
			return err
		}
		w, err := destW.Create(trueZipFileName)
		if err != nil {
			return err
		}
		openF, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer openF.Close()
		_, err = io.Copy(w, openF)
		return err
	})
	if err := destW.Close(); err != nil {
		return err
	}
	return nil
}
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			err = os.MkdirAll(fpath, f.Mode())
			if err != nil {
				return err
			}
		} else {
			err = os.MkdirAll(filepath.Dir(fpath), f.Mode())
			if err != nil {
				return err
			}
			f, err := os.OpenFile(
				fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (p *project) ImportData(impPathName string) error {
	//sub directory is table
	dirs, err := ioutil.ReadDir(impPathName)
	if err != nil {
		return err
	}
	return grade.RunAtTrans(p.DBHelper().DbUrl(), func(helper *grade.PGHelper) error {
		for _, file := range dirs {
			if file.IsDir() {
				err := helper.Import(filepath.Join(impPathName, file.Name()))
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (p *project) ReloadRepository() error {
	repositoryPath := RepositoryPath(p.repository)
	log.TRACE.Printf("ReloadRepository,%q", repositoryPath)
	if _, err := os.Stat(repositoryPath); err == os.ErrExist {
		return nil
	} else if err != nil {
		return err
	}
	return p.ImportData(repositoryPath)
}
func (p *project) loadPackage(rm *otto.Otto, fileName string, gradestr grade.Grade) (*otto.Script, error) {
	rev := ""
	if err := p.DBHelper().Query(func(rows *sql.Rows) error {
		var script string
		for rows.Next() {
			if err := rows.Scan(&script); err != nil {
				return err
			}
			rev += "\n" + script
		}
		return nil
	}, SQL_GetPackage, fileName, gradestr.String()); err != nil {
		return nil, err
	}
	if rev == "" {
		return nil, fmt.Errorf("empty package:%s(%s)", fileName, gradestr)
	}
	rev = "(function(module) {var require = module.require;var exports = module.exports;\n" + rev + "\n})"
	return rm.Compile(fileName, rev)
}
func (p *project) Require(rm *otto.Otto, fileName, currentModuleDir string, gradestr grade.Grade) (*otto.Script, string, error) {
	var moduleFileName string
	//if is root path,then cut the /
	if strings.HasPrefix(fileName, "/") {
		moduleFileName = fileName
	} else {
		moduleFileName = path.Join(currentModuleDir, fileName)
	}
	moduleName := struct {
		Name  string
		Grade grade.Grade
	}{moduleFileName, gradestr}
	p.lockPackage.Lock()
	defer p.lockPackage.Unlock()
	if rev, ok := p.cachePackage[moduleName]; !ok {
		m, err := p.loadPackage(rm, moduleFileName, gradestr)
		if err != nil {
			return nil, "", err
		}
		p.cachePackage[moduleName] = m
		return m, moduleFileName, nil
	} else {
		return rev, moduleFileName, nil
	}
}
func (p *project) GetPackageNames(dirName string, gradestr grade.Grade) ([]string, error) {
	rev := []string{}
	if err := p.DBHelper().Query(func(rows *sql.Rows) error {
		var name string
		for rows.Next() {
			if err := rows.Scan(&name); err != nil {
				return err
			}
			rev = append(rev, name)
		}
		return nil
	}, SQL_GetPackageNames, dirName, string(gradestr)); err != nil {
		return nil, err
	}
	return rev, nil
}
