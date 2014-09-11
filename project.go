package main

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/linlexing/dbgo/grade"
	"github.com/linlexing/dbgo/log"
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/robertkrimen/otto"
	"github.com/russross/blackfriday"
	"html/template"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	BEFORE int64 = iota

	//AFTER
	//PANIC
	//FINALLY

	NotFoundAction = "NotFoundAction"
)

type TranslateString map[string]string

type Project interface {
	Name() string
	DisplayLabel() TranslateString
	NewDBHelper() *grade.DBHelper

	DefaultAction() (string, string)
	ReverseUrl(args ...string) string
	ClearCache()
	ReloadRepository() error
	GetPackageNames(dirName string, gradestr grade.Grade) ([]string, error)
	Require(rm *otto.Otto, fileName, currentModuleDir string, gradestr grade.Grade) (*otto.Script, string, error)

	WSHub() *WSHub
	Version(grade grade.Grade) (int64, bool, error)
	//Model(mname string, gradestr grade.Grade) *grade.DataTable
	DBModel(gradestr grade.Grade, mnames ...string) []*grade.DBTable
	ExportData(dumpName string, expFile *os.File, gradestr grade.Grade) error
	ImportData(impPath string) error
	Checks(tablename string, gradestr grade.Grade) ([]*Check, error)
	TemplateSet(f template.FuncMap) (*template.Template, error)
	Object() map[string]interface{}
}
type EmptyPackageError struct {
	Message string
}

func (e *EmptyPackageError) Error() string {
	return e.Message
}

type project struct {
	name         string
	displayLabel TranslateString
	repository   string
	dburl        string
	wsHub        *WSHub

	lockTableDefine  *sync.Mutex
	lockVersion      *sync.Mutex
	lockTemplateSet  *sync.Mutex
	lockCheck        *sync.Mutex
	lockPackage      *sync.Mutex
	lockPackageNames *sync.Mutex

	cachePackage map[struct {
		Name  string
		Grade grade.Grade
	}]*otto.Script
	cacheVersion      map[grade.Grade]int64 //每级grade均对应一个最新的版本
	cacheTableDefine  map[string]*grade.DataTable
	cacheTemplateSet  *template.Template
	cacheCheck        map[string][]*Check
	cachePackageNames map[struct {
		FileName string
		Grade    grade.Grade
	}][]string
}

/*
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
}*/
func RepositoryPath(repo string) string {
	return filepath.Join(AppPath, "repository", repo, "_pd")
}

/*
func publicTables() []*grade.DataTable {
	return []*grade.DataTable{lx_dump(), lx_checkaddition(), lx_check()}

}*/
func NewProject(name string, label TranslateString, dburl, repository string) Project {

	return &project{
		name:         name,
		displayLabel: label,
		dburl:        dburl,
		repository:   repository,
		wsHub: &WSHub{
			broadcast:   make(chan SocketMessage),
			register:    make(chan *WSConn),
			unregister:  make(chan *WSConn),
			connections: make(map[string]map[*WSConn]bool),
		},
		lockTableDefine:  &sync.Mutex{},
		lockVersion:      &sync.Mutex{},
		lockTemplateSet:  &sync.Mutex{},
		lockCheck:        &sync.Mutex{},
		lockPackage:      &sync.Mutex{},
		lockPackageNames: &sync.Mutex{},

		cacheVersion:     map[grade.Grade]int64{},
		cacheTableDefine: map[string]*grade.DataTable{},
		cacheTemplateSet: nil,
		cacheCheck:       map[string][]*Check{},
		cachePackage: map[struct {
			Name  string
			Grade grade.Grade
		}]*otto.Script{},
		cachePackageNames: map[struct {
			FileName string
			Grade    grade.Grade
		}][]string{},
	}
	/*
		//先确保有公共的表
		for _, v := range publicTables() {
			if err := p.dbHelper.UpdateStruct(v); err != nil {
				return nil, err
			}
		}
		return p, nil*/
}
func (p *project) Model(mname string, gradestr grade.Grade) *grade.DataTable {
	return p.table(mname, gradestr)
}
func (p *project) DBModel(gradestr grade.Grade, mnames ...string) []*grade.DBTable {
	return p.dbTable(gradestr, mnames...)
}
func (p *project) loadVersion() (map[grade.Grade]int64, error) {
	AHelp := p.NewDBHelper()
	if err := AHelp.Open(); err != nil {
		return nil, err
	}
	defer AHelp.Close()
	tab, err := AHelp.GetData("select grade,max(verno) as verno from lx_version group by grade")
	if err != nil {
		return nil, err
	}
	rev := map[grade.Grade]int64{}
	for i := 0; i < tab.RowCount(); i++ {
		row := tab.Row(i)
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
		"markdown": func(value interface{}) template.HTML {
			str := ""
			switch tv := value.(type) {
			case string:
				str = tv
			case template.HTML:
				str = string(tv)
			default:
				str = fmt.Sprintf("the value:%v(%T) invalid", value, value)
			}
			rev := string(blackfriday.MarkdownCommon([]byte(str)))
			return template.HTML(rev)
		},
		"mkSlice": func(args ...interface{}) []interface{} {
			return args
		},
		"mkMap": func() map[string]interface{} {
			return map[string]interface{}{}
		},
		"static": func(filename string) string {
			return p.ReverseUrl("static", "file", filename)
		},
		"title": func(title, urlstr string) string {
			return fmt.Sprintf("%s?title=%s", urlstr, url.QueryEscape(title))
		},
		"HTMLAttr": func(src string) template.HTMLAttr {
			return template.HTMLAttr(src)
		},
		"JS": func(src string) template.JS {
			return template.JS(src)
		},
		"pathJoin": path.Join,
		/*"GradeCanUse": func(g1, g2 string) bool {
			return grade.Grade(g1).CanUse(grade.Grade(g2))
		},*/
		"tmpl": func(tmpl string, data map[string]interface{}) template.HTML {
			buf := &bytes.Buffer{}
			if err := rev.ExecuteTemplate(buf, tmpl, data); err != nil {
				return template.HTML(err.Error())
			} else {
				return template.HTML(buf.String())
			}
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
	AHelp := p.NewDBHelper()
	if err := AHelp.Open(); err != nil {
		return nil, err
	}
	defer AHelp.Close()
	tab, err := AHelp.GetData("select name,content,grade from lx_view order by name")
	if err != nil {
		return nil, err
	}
	for i := 0; i < tab.RowCount(); i++ {
		row := tab.Row(i)
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

func (p *project) DefaultAction() (string, string) {
	return "login", "show"
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

/*func (p *project) jsModel(call otto.FunctionCall) otto.Value {
	tablename := oftenfun.AssertString(call.Argument(0))
	gradestr := grade.Grade(oftenfun.AssertString(call.Argument(1)))
	return oftenfun.JSToValue(call.Otto, p.Model(tablename, gradestr).Object())
}*/
func (p *project) jsDBModel(call otto.FunctionCall) otto.Value {
	gradestr := grade.Grade(oftenfun.AssertString(call.Argument(0)))
	tnames := make([]string, len(call.ArgumentList)-1)
	for i, v := range call.ArgumentList[1:] {
		tnames[i] = oftenfun.AssertString(v)
	}
	rev := make([]interface{}, len(tnames))
	for i, v := range p.DBModel(gradestr, tnames...) {
		rev[i] = v.Object()
	}
	return oftenfun.JSToValue(call.Otto, rev)
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
		"NewDBHelper":      p.jsNewDBHelper,
		"DBModel":          p.jsDBModel,
		"Checks":           p.jsChecks,
		"ReloadRepository": p.jsReloadRepository,
	}
}
func (p *project) jsNewDBHelper(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, p.NewDBHelper().Object())
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
func (p *project) ClearPackageNamesCache() {
	p.lockPackageNames.Lock()
	defer p.lockPackageNames.Unlock()
	p.cachePackageNames = map[struct {
		FileName string
		Grade    grade.Grade
	}][]string{}
	return
}
func (p *project) ClearCache() {
	log.INFO.Printf("project %s clear cache.", p.Name())
	p.ClearCheckCache()
	p.ClearPackageCache()
	p.ClearTableDefineCache()
	p.ClearTemplateSetCache()
	p.ClearVersionCache()
	p.ClearPackageNamesCache()
}

func (p *project) getTableDefine(tablename string) (result *grade.DataTable, err error) {
	p.lockTableDefine.Lock()
	defer p.lockTableDefine.Unlock()
	var ok bool
	if result, ok = p.cacheTableDefine[tablename]; !ok {
		var tab *grade.DataTable
		help := p.NewDBHelper()
		if err := help.Open(); err != nil {
			return nil, err
		}
		defer help.Close()
		tab, err = help.Table(tablename)
		if err != nil {
			return
		}
		result = tab
		p.cacheTableDefine[tablename] = result
	}
	return
}
func (p *project) table(mname string, gradestr grade.Grade) *grade.DataTable {
	tab, err := p.getTableDefine(mname)
	if err != nil {
		panic(err)
	}
	result, ok := tab.Reduced(gradestr)
	if !ok {
		panic(fmt.Errorf("the table %q not exits at grade:%q", mname, gradestr))
	}
	return result
}
func (p *project) dbTable(gradestr grade.Grade, mnames ...string) []*grade.DBTable {
	rev := make([]*grade.DBTable, len(mnames))
	ahelp := p.NewDBHelper()
	for i, v := range mnames {
		tab, err := p.getTableDefine(v)
		if err != nil {
			panic(err)
		}
		tab1, ok := tab.Reduced(gradestr)
		if !ok {
			panic(fmt.Errorf("the table %q not exits at grade:%q", v, gradestr))
		}
		rev[i] = grade.NewDBTable(ahelp, tab1)
	}
	return rev
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
	h := p.NewDBHelper()
	if err := h.Open(); err != nil {
		return nil, err
	}
	defer h.Close()
	rows, err := h.Query(`
		select
			a.id,
			a.displaylabel_en,
			a.displaylabel_cn,
			a.level,
			a.fields,
			b.fields as fields_add,
			a.runatserver,
			b.runatserver as runatserver_add,
			a.script,
			b.script as script_add,
			a.sqlwhere,
			b.sqlwhere as sqlwhere_add,
			a.grade
		from lx_check a left join lx_checkaddition b on a.tablename=b.tablename and a.addition=b.addition
		where a.tablename={{ph}}`, tablename)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id, level int64
		var displaylabel_en, displaylabel_cn, grade string
		var runAtServer, runAtServer_add sql.NullBool
		var script, script_add, sqlWhere, sqlWhere_add, fields, fields_add sql.NullString
		if err := rows.Scan(
			&id,
			&displaylabel_en,
			&displaylabel_cn,
			&level,
			&fields,
			&fields_add,
			&runAtServer,
			&runAtServer_add,
			&script,
			&script_add,
			&sqlWhere,
			&sqlWhere_add,
			&grade); err != nil {
			return nil, err
		}
		c := &Check{
			ID:           id,
			DisplayLabel: TranslateString{"en": displaylabel_en, "cn": displaylabel_cn},
			Level:        level,
			Fields:       []string{},
			RunAtServer:  false,
			Script:       "",
			SqlWhere:     "",
			Grade:        grade,
		}
		if fields.Valid {
			c.Fields = strings.Split(fields.String, ",")
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
	dumpData := p.DBModel(gradestr, "lx_dump")[0]
	if err := dumpData.DBHelper().Open(); err != nil {
		return err
	}
	defer dumpData.DBHelper().Close()

	if err := dumpData.FillWhere("name={{ph}}", dumpName); err != nil {
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
		row := dumpData.Row(i)
		fileColumns := map[string]string{}
		if row["filecolumns"] != nil {
			if err := json.Unmarshal([]byte(row["filecolumns"].(string)), &fileColumns); err != nil {
				return err
			}
		}
		param := &grade.ExportParam{
			TableName:        row["tablename"].(string),
			CurrentGrade:     gradestr,
			PathName:         filepath.Join(tmpDir, row["tablename"].(string)),
			FileColumns:      fileColumns,
			SqlWhere:         oftenfun.SafeToString(row["sqlwhere"]),
			ImpAutoUpdate:    oftenfun.SafeToBool(row["impautoupdate"]),
			ImpAutoRemove:    oftenfun.SafeToBool(row["impautoremove"]),
			RunAtImport:      oftenfun.SafeToString(row["sqlrunatimport"]),
			ImpRefreshStruct: oftenfun.SafeToBool(row["imprefreshstruct"]),
			CheckVersion:     oftenfun.SafeToBool(row["checkversion"]),
		}
		if err := dumpData.DBHelper().Export(param); err != nil {
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
func (p *project) ImportData(impPathName string) (err error) {
	//sub directory is table
	dirs, err := ioutil.ReadDir(impPathName)
	if err != nil {
		return err
	}
	h := p.NewDBHelper()
	if err = h.Open(); err != nil {
		return
	}
	defer h.Close()
	if err = h.Begin(); err != nil {
		return
	}
	defer func() {
		if v := recover(); v != nil {
			switch tv := v.(type) {
			case error:
				err = tv
			case string:
				err = fmt.Errorf("%s", tv)
			default:
				err = fmt.Errorf("new error")
			}

		}
		if err == nil {
			err = h.Commit()
		} else {
			h.Rollback()
		}
	}()
	for _, file := range dirs {
		if file.IsDir() {
			err = h.Import(filepath.Join(impPathName, file.Name()))
			if err != nil {
				return
			}
		}
	}
	return nil
}
func (p *project) dumpStaticFile() error {
	h := p.NewDBHelper()
	if err := h.Open(); err != nil {
		return err
	}
	defer h.Close()
	if exists, err := h.TableExists("lx_static"); err != nil {
		return err
	} else if exists {
		rows, err := h.Query("select filename,content,lasttime from lx_static order by filename")
		if err != nil {
			return err
		}
		var filename string
		var filetime time.Time
		var content sql.RawBytes
		for rows.Next() {
			if err := rows.Scan(&filename, &content, &filetime); err != nil {
				return err
			}
			trueFileName := filepath.Join(AppPath, "static", p.Name(), filename)
			if err := os.MkdirAll(filepath.Dir(trueFileName), os.ModePerm); err != nil {
				return err
			}
			info, err := os.Stat(trueFileName)
			if os.IsNotExist(err) || (err == nil && filetime.After(info.ModTime())) {
				if err := ioutil.WriteFile(trueFileName, []byte(content), os.ModePerm); err != nil {
					return err
				}
				if err := os.Chtimes(trueFileName, filetime, filetime); err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
		}
	}
	return nil
}
func (p *project) ReloadRepository() error {
	repositoryPath := RepositoryPath(p.repository)
	log.TRACE.Printf("ReloadRepository,%s,db:%s", repositoryPath, p.dburl)
	if _, err := os.Stat(repositoryPath); err == os.ErrExist {
		return nil
	} else if err != nil {
		return err
	}
	p.lockCheck.Lock()
	defer p.lockCheck.Unlock()
	p.lockPackage.Lock()
	defer p.lockPackage.Unlock()
	p.lockPackageNames.Lock()
	defer p.lockPackageNames.Unlock()
	p.lockTableDefine.Lock()
	defer p.lockTableDefine.Unlock()
	p.lockTemplateSet.Lock()
	defer p.lockTemplateSet.Unlock()
	p.lockVersion.Lock()
	defer p.lockVersion.Unlock()
	if err := p.ImportData(repositoryPath); err != nil {
		return err
	}
	//fetch the static file to disk
	if err := p.dumpStaticFile(); err != nil {
		return err
	}
	return nil
}
func (p *project) loadPackage(rm *otto.Otto, fileName string, gradestr grade.Grade) (*otto.Script, error) {
	strTmp := []string{}
	h := p.NewDBHelper()
	if err := h.Open(); err != nil {
		return nil, err
	}
	defer h.Close()
	rows, err := h.Query(
		`select script from lx_package where filename like {{strcat ph "'%'"}} and {{reglike (printf "right(filename,length(filename)-length(%s))" ph) (str "^[^/]*$")}} and grade_canuse({{ph}},grade) order by filename`,
		fileName, fileName, gradestr.String())
	if err != nil {
		return nil, err
	}
	var script string
	for rows.Next() {
		if err := rows.Scan(&script); err != nil {
			return nil, err
		}
		strTmp = append(strTmp, script)
	}
	str := strings.Join(strTmp, "\n")
	if len(strTmp) == 0 {
		return nil, &EmptyPackageError{fmt.Sprintf("empty package:%s(%s)", fileName, gradestr)}
	}
	rev := "(function(module) {var require = module.require;var safeRequire = module.safeRequire;var exports = module.exports;" + str + ";module.exports=exports;})"
	src, err := rm.Compile(fileName, rev)
	if err != nil {
		src := []string{}
		for i, v := range strings.Split(str, "\n") {
			src = append(src, fmt.Sprintf("%d\t%s", i+1, v))
		}
		return nil, fmt.Errorf("error:%s\nfile:%s\nsource:\n%s", err, fileName, strings.Join(src, "\n"))
	}
	return src, nil
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
func (p *project) GetPackageNames(dirNameStr string, gradestr grade.Grade) ([]string, error) {
	p.lockPackageNames.Lock()
	defer p.lockPackageNames.Unlock()
	dirName := struct {
		FileName string
		Grade    grade.Grade
	}{dirNameStr, gradestr}
	if rev, ok := p.cachePackageNames[dirName]; ok {
		return rev, nil
	} else {
		rev := []string{}
		h := p.NewDBHelper()
		if err := h.Open(); err != nil {
			return nil, err
		}
		defer h.Close()
		rows, err := h.Query(`
			select filename
			from lx_package
			where filename like {{"'%'"|strcat ph}} and
				{{with $A:=(ph|printf "right(filename,length(filename)-length(%s))")}}
				{{with $B:=("^/?[^/]+\\.js$"|str)}}
					{{reglike $A $B}} and
				{{end}}
				{{end}}
				grade_canuse({{ph}},grade)
			order by filename`,
			dirNameStr, dirNameStr, gradestr.String())
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return nil, err
			}
			rev = append(rev, name)
		}
		p.cachePackageNames[dirName] = rev
		return rev, nil
	}

}
func (p *project) DisplayLabel() TranslateString {
	return p.displayLabel
}
func (p *project) NewDBHelper() *grade.DBHelper {
	return grade.NewDBHelper(p.dburl)
}
func (p *project) WSHub() *WSHub {
	return p.wsHub
}
