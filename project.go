package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/linlexing/dbgo/grade"
	"github.com/linlexing/dbgo/jsmvcerror"
	"github.com/linlexing/dbgo/log"
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/linlexing/dbgo/pghelp"
	"github.com/robertkrimen/otto"
	"html/template"
	"net/url"
	"path"
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
	MetaProject() string
	Grade() string
	DBHelp() *pghelp.PGHelp
	DefaultAction() (string, string)
	ReverseUrl(args ...string) string
	ClearCache()

	Version(grade string) (int64, bool, error)
	InterceptScript(grade string, when int64) (string, error)
	Controller(ctrlname, gradestr string) (*Controller, error)
	Table(tablename, gradestr string) *pghelp.DBTable
	Checks(tablename, gradestr string) ([]*Check, error)
	TemplateSet(f template.FuncMap) (*template.Template, error)
	Object() map[string]interface{}
}

type project struct {
	name          string
	dbHelp        *pghelp.PGHelp
	metaProject   string
	grade         string
	defaultAction string

	lockController  *sync.Mutex
	lockTableDefine *sync.Mutex
	lockVersion     *sync.Mutex
	lockIntercept   *sync.Mutex
	lockTemplateSet *sync.Mutex
	lockCheck       *sync.Mutex

	cacheVersion     map[string]int64 //每级grade均对应一个最新的版本
	cacheIntercept   []*Intercept
	cacheTableDefine map[string]*pghelp.DataTable
	cacheController  map[string]*Controller
	cacheTemplateSet *template.Template
	cacheCheck       map[string][]*Check
}
type Action struct {
	ID     string
	Script string
	Grade  string
}

type ViewRegion struct {
	Content string
	Grade   string
}
type View struct {
	Name    string
	Regions []*ViewRegion
	Grade   string
}

func lx_version() *pghelp.DataTable {
	table := pghelp.NewDataTable("lx_version")
	table.Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumn("grade", pghelp.TypeString, true)).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumn("verno", pghelp.TypeInt64, true)).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumn("verlabel", pghelp.TypeString, true)).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumn("changelog", pghelp.TypeString, true)).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumn("releasetime", pghelp.TypeTime, true)).Desc.Grade = grade.GRADE_ROOT
	table.SetPK("grade", "verno")
	return table
}
func lx_intercept() *pghelp.DataTable {
	table := pghelp.NewDataTable("lx_intercept")
	table.Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumn("id", pghelp.TypeString, true)).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumn("script", pghelp.TypeString, true)).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumn("whenint", pghelp.TypeInt64, true)).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumn("grade", pghelp.TypeString, true)).Desc.Grade = grade.GRADE_ROOT
	table.SetPK("id")
	return table
}
func lx_controller() *pghelp.DataTable {

	table := pghelp.NewDataTable("lx_controller")
	table.Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("name", pghelp.NewPGType(pghelp.TypeString, 0, true), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("script", pghelp.NewPGType(pghelp.TypeString, 0, false), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("public", pghelp.NewPGType(pghelp.TypeBool, 0, true), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("grade", pghelp.NewPGType(pghelp.TypeString, 0, true), "")).Desc.Grade = grade.GRADE_ROOT
	table.SetPK("name")

	return table
}
func lx_action() *pghelp.DataTable {
	table := pghelp.NewDataTable("lx_action")
	table.Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("ctrlname", pghelp.NewPGType(pghelp.TypeString, 0, true), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("id", pghelp.NewPGType(pghelp.TypeString, 0, true), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("script", pghelp.NewPGType(pghelp.TypeString, 0, false), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("grade", pghelp.NewPGType(pghelp.TypeString, 0, true), "")).Desc.Grade = grade.GRADE_ROOT
	table.SetPK("ctrlname", "id")
	return table
}
func lx_view() *pghelp.DataTable {
	table := pghelp.NewDataTable("lx_view")
	table.Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("name", pghelp.NewPGType(pghelp.TypeString, 0, true), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("grade", pghelp.NewPGType(pghelp.TypeString, 0, true), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("content", pghelp.NewPGType(pghelp.TypeString, 0, false), "")).Desc.Grade = grade.GRADE_ROOT
	table.SetPK("name")
	return table
}

func lx_checkaddition() *pghelp.DataTable {

	table := pghelp.NewDataTable("lx_checkaddition")
	table.Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("tablename", pghelp.NewPGType(pghelp.TypeString, 64, true), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("addition", pghelp.NewPGType(pghelp.TypeString, 64, true), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("fields", pghelp.NewPGType(pghelp.TypeStringSlice, 0, false), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("runatserver", pghelp.NewPGType(pghelp.TypeBool, 0, false), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("script", pghelp.NewPGType(pghelp.TypeString, 0, false), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("sqlwhere", pghelp.NewPGType(pghelp.TypeString, 0, false), "")).Desc.Grade = grade.GRADE_ROOT
	table.SetPK("tablename", "addition")
	return table
}
func lx_check() *pghelp.DataTable {

	table := pghelp.NewDataTable("lx_check")
	table.Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("tablename", pghelp.NewPGType(pghelp.TypeString, 64, true), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("id", pghelp.NewPGType(pghelp.TypeInt64, 0, true), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("displaylabel", pghelp.NewPGType(pghelp.TypeString, 0, true), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("level", pghelp.NewPGType(pghelp.TypeInt64, 0, true), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("fields", pghelp.NewPGType(pghelp.TypeStringSlice, 0, false), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("runatserver", pghelp.NewPGType(pghelp.TypeBool, 0, false), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("addition", pghelp.NewPGType(pghelp.TypeString, 64, false), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("script", pghelp.NewPGType(pghelp.TypeString, 0, false), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("sqlwhere", pghelp.NewPGType(pghelp.TypeString, 0, false), "")).Desc.Grade = grade.GRADE_ROOT
	table.AddColumn(pghelp.NewColumnT("grade", pghelp.NewPGType(pghelp.TypeString, 0, true), "")).Desc.Grade = grade.GRADE_ROOT
	table.SetPK("tablename", "id")
	return table
}

func publicTables() []*pghelp.DataTable {
	return []*pghelp.DataTable{lx_version(), lx_intercept(), lx_controller(), lx_action(), lx_view(), lx_checkaddition(), lx_check()}

}
func NewProject(name, dburl string) (Project, error) {
	p := &project{
		name:   name,
		dbHelp: pghelp.NewPGHelp(dburl),

		lockController:  &sync.Mutex{},
		lockTableDefine: &sync.Mutex{},
		lockVersion:     &sync.Mutex{},
		lockIntercept:   &sync.Mutex{},
		lockTemplateSet: &sync.Mutex{},
		lockCheck:       &sync.Mutex{},

		cacheVersion:     map[string]int64{},
		cacheIntercept:   []*Intercept{},
		cacheTableDefine: map[string]*pghelp.DataTable{},
		cacheController:  map[string]*Controller{},
		cacheTemplateSet: nil,
		cacheCheck:       map[string][]*Check{},
	}

	schema, err := p.dbHelp.Schema()
	if err != nil {
		return nil, err
	}
	p.grade = schema.Desc.Grade
	p.metaProject = schema.Desc.MetaProject
	p.defaultAction = schema.Desc.DefaultAction
	//先确保有公共的表
	for _, v := range publicTables() {
		if err := p.dbHelp.UpdateStruct(v, grade.GRADE_ROOT); err != nil {
			return nil, err
		}
	}
	return p, nil
}

func (p *project) loadVersion() (map[string]int64, error) {
	tab, err := p.dbHelp.GetDataTable(SQL_GerVersion)
	if err != nil {
		return nil, err
	}
	rev := map[string]int64{}
	for i := 0; i < tab.RowCount(); i++ {
		row := tab.GetRow(i)
		rev[row["grade"].(string)] = row["verno"].(int64)
	}
	return rev, nil
}
func (p *project) loadIntercept() ([]*Intercept, error) {
	var tab *pghelp.DataTable
	//获取Intercept
	tab, err := p.dbHelp.GetDataTable(SQL_GetIntercept)
	if err != nil {
		return nil, err
	}
	rev := make([]*Intercept, tab.RowCount(), tab.RowCount())
	for i := 0; i < tab.RowCount(); i++ {
		row := tab.GetRow(i)
		rev[i] = &Intercept{
			ID:            row["id"].(string),
			InterceptWhen: row["whenint"].(int64),
			Script:        row["script"].(string),
			Grade:         row["grade"].(string),
		}
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
		"PathJoin":    path.Join,
		"GradeCanUse": grade.GradeCanUse,
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
	tab, err := p.dbHelp.GetDataTable(SQL_GetView)
	if err != nil {
		return nil, err
	}
	for i := 0; i < tab.RowCount(); i++ {
		row := tab.GetRow(i)
		content := fmt.Sprintf(`
			<#if GradeCanUse .c.CurrentGrade %q#>
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
func (p *project) Grade() string {
	return p.grade
}
func (p *project) MetaProject() string {
	return p.metaProject
}
func (p *project) Version(grade string) (int64, bool, error) {
	p.lockVersion.Lock()
	defer p.lockVersion.Unlock()
	if len(p.cacheVersion) == 0 {
		ver, err := p.loadVersion()
		if err != nil {
			return 0, false, err
		}
		p.cacheVersion = ver
	}
	v, ok := p.cacheVersion[grade]
	return v, ok, nil
}

func (p *project) DBHelp() *pghelp.PGHelp {
	return p.dbHelp
}
func (p *project) DefaultAction() (string, string) {
	s := strings.Split(p.defaultAction, ".")
	if len(s) == 2 {
		return s[0], s[1]
	} else {
		return "login", "login"
	}
}

func (p *project) InterceptScript(gradestr string, when int64) (string, error) {
	p.lockIntercept.Lock()
	defer p.lockIntercept.Unlock()
	if len(p.cacheIntercept) == 0 {
		inter, err := p.loadIntercept()
		if err != nil {
			return "", err
		}
		p.cacheIntercept = inter
	}
	result := []string{}

	for _, inp := range p.cacheIntercept {
		if grade.GradeCanUse(gradestr, inp.Grade) && inp.InterceptWhen == when {
			result = append(result, inp.Script)
		}
	}
	if len(result) > 0 {
		return fmt.Sprintf(`
		(function Intercept(c){
			var filter = [%s];
			var GRADE = %q;
			filter[0](c,filter.slice(1));
		})`, strings.Join(result, ","), gradestr), nil
	} else {
		return "", nil
	}

}

func (p *project) loadController(ctrlname string) (*Controller, error) {
	//从数据库取出定义
	ctrlTab, err := p.dbHelp.GetDataTable(SQL_GetController, ctrlname)

	if err != nil {
		return nil, err
	}
	if ctrlTab.RowCount() == 0 {
		return nil, jsmvcerror.NotFoundControl
	}
	ctrlRow := ctrlTab.GetRow(0)
	ctrl := &Controller{
		Name:   ctrlname,
		Script: oftenfun.SafeToString(ctrlRow["script"]),
		Public: oftenfun.SafeToBool(ctrlRow["public"]),
		Grade:  oftenfun.SafeToString(ctrlRow["grade"]),
	}
	actionTab, err := p.dbHelp.GetDataTable(SQL_GetAction, ctrlname)
	if err != nil {
		return nil, err
	}
	for i := 0; i < actionTab.RowCount(); i++ {
		actionRow := actionTab.GetRow(i)
		ctrl.Actions = append(ctrl.Actions, &Action{
			ID:     actionRow["id"].(string),
			Script: actionRow["script"].(string),
			Grade:  actionRow["grade"].(string),
		})
	}
	return ctrl, nil
}
func (p *project) Controller(ctrlname, grade string) (*Controller, error) {
	p.lockController.Lock()
	defer p.lockController.Unlock()
	if _, ok := p.cacheController[ctrlname]; !ok {
		if ctrl, err := p.loadController(ctrlname); err != nil {
			return nil, err
		} else {
			p.cacheController[ctrlname] = ctrl
		}
	}
	return p.cacheController[ctrlname], nil
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
func (p *project) jsTable(call otto.FunctionCall) otto.Value {
	tablename := oftenfun.AssertString(call.Argument(0))
	gradestr := oftenfun.AssertString(call.Argument(1))
	return oftenfun.JSToValue(call.Otto, p.Table(tablename, gradestr).Object())
}
func (p *project) jsChecks(call otto.FunctionCall) otto.Value {
	tablename := oftenfun.AssertString(call.Argument(0))
	gradestr := oftenfun.AssertString(call.Argument(1))
	chks, err := p.Checks(tablename, gradestr)
	if err != nil {
		panic(err)
	}
	return oftenfun.JSToValue(call.Otto, chks)
}

func (p *project) Object() map[string]interface{} {
	return map[string]interface{}{
		"ReverseUrl": p.jsReverseUrl,
		"Name":       p.Name(),
		"ClearCache": p.jsClearCache,
		"DBHelp":     p.DBHelp().Object(),
		"Table":      p.jsTable,
		"Checks":     p.jsChecks,
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
func (p *project) ClearControllerCache() {
	p.lockController.Lock()
	defer p.lockController.Unlock()
	p.cacheController = map[string]*Controller{}
	return
}
func (p *project) ClearInterceptCache() {
	p.lockIntercept.Lock()
	defer p.lockIntercept.Unlock()
	p.cacheIntercept = nil
	return
}
func (p *project) ClearTableDefineCache() {
	p.lockTableDefine.Lock()
	defer p.lockTableDefine.Unlock()
	p.cacheTableDefine = map[string]*pghelp.DataTable{}
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
	p.cacheVersion = map[string]int64{}
	return
}
func (p *project) ClearCache() {
	log.INFO.Printf("project %s clear cache.", p.Name())
	p.ClearCheckCache()
	p.ClearControllerCache()
	p.ClearInterceptCache()
	p.ClearTableDefineCache()
	p.ClearTemplateSetCache()
	p.ClearVersionCache()
}

func (p *project) getTableDefine(tablename string) (result *pghelp.DataTable, err error) {
	p.lockTableDefine.Lock()
	defer p.lockTableDefine.Unlock()
	var ok bool
	if result, ok = p.cacheTableDefine[tablename]; !ok {
		var tab *pghelp.DBTable
		tab, err = p.DBHelp().Table(tablename)
		if err != nil {
			return
		}
		result = tab.DataTable
		p.cacheTableDefine[tablename] = result
	}
	return
}
func (p *project) Table(tablename, gradestr string) *pghelp.DBTable {
	tab, err := p.getTableDefine(tablename)
	if err != nil {
		panic(err)
	}
	result, ok := reducedTableDefine(tab, gradestr)
	if !ok {
		panic(fmt.Errorf("the table %q not exits at grade:%q", tablename, gradestr))
	}
	return pghelp.NewDBTable(p.DBHelp(), result)
}
func reducedTableDefine(tab *pghelp.DataTable, gradestr string) (*pghelp.DataTable, bool) {
	if !grade.GradeCanUse(gradestr, tab.Desc.Grade) {
		return nil, false
	}
	result := pghelp.NewDataTable(tab.TableName)
	result.Desc.Grade = tab.Desc.Grade
	//process the columns
	for _, col := range tab.Columns() {
		//the key column must to be add
		if grade.GradeCanUse(gradestr, col.Desc.Grade) || tab.IsPrimaryKey(col.Name) {
			result.AddColumn(col.Clone())
		}
	}
	result.SetPK(tab.GetPK()...)
	//process the indexes
	for idxname, idx := range tab.Indexes {
		if grade.GradeCanUse(gradestr, idx.Desc.Grade) {
			result.Indexes[idxname] = idx.Clone()
		}
	}
	return result, true
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
	err := p.DBHelp().Query(func(rows *sql.Rows) error {
		for rows.Next() {
			var id, level int64
			var displaylabel, grade string
			var runAtServer pghelp.NullBool
			var script, sqlWhere pghelp.NullString
			var fields pghelp.NullStringSlice
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
func (p *project) Checks(tablename, gradestr string) ([]*Check, error) {
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
		if grade.GradeCanUse(gradestr, chk.Grade) {
			rev = append(rev, chk)
		}
	}
	return rev, nil
}
