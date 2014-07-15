package main

import (
	"fmt"
	"github.com/linlexing/dbgo/grade"
	"github.com/linlexing/dbgo/jsmvcerror"
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/robertkrimen/otto"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	SessionAuthUrlPrex = "auth."
)

type ControllerAgent struct {
	Request          *http.Request
	Response         http.ResponseWriter
	Session          *Session
	Project          Project
	JsonBody         map[string]interface{}
	UserName         string
	ControllerName   string
	ActionName       string
	TagPath          string
	Title            string
	Public           bool        //Indicate whether you can access without authentication
	CurrentGrade     grade.Grade //When this call, the user's Grade properties. After obtaining the value from the Session, and not to change it, even if the user changes their Grade other requests
	Result           Result
	TemplateFun      template.FuncMap
	Object           otto.Value
	ControllerScript otto.Value
	jsRuntime        *otto.Otto
}

func (c *ControllerAgent) Require(fileName, currentModuleDir string) otto.Value {

	script, moduleFileName, err := c.Project.Require(c.jsRuntime, fileName, currentModuleDir, c.CurrentGrade)
	if err != nil {
		panic(err)
	}
	jsfun, err := c.jsRuntime.Run(script)
	if err != nil {
		panic(err)
	}

	tmpValue, err := c.jsRuntime.ToValue(map[string]interface{}{
		"exports":    map[string]interface{}{},
		"__dirname":  path.Dir(moduleFileName),
		"__filename": path.Base(moduleFileName),
		"require": func(call otto.FunctionCall) otto.Value {
			return c.Require(oftenfun.AssertString(call.Argument(0)), path.Dir(moduleFileName))
		},
	})
	if err != nil {
		panic(err)
	}
	jsModule := tmpValue.Object()
	exports, _ := jsModule.Get("exports")
	rev, err := jsfun.Call(exports, jsModule)
	if err != nil {
		panic(err)
	}
	if rev.IsUndefined() {
		exports, _ := jsModule.Get("exports")
		return exports
	} else {
		return rev
	}
}
func (c *ControllerAgent) QueryValues() url.Values {
	return c.Request.URL.Query()
}

func NewAgent(w http.ResponseWriter, r *http.Request) *ControllerAgent {

	c := &ControllerAgent{
		Request:  r,
		Response: w,
	}
	c.TemplateFun = template.FuncMap{
		"url":     c.Url,
		"authurl": c.AuthUrl,
	}

	return c
}
func (c *ControllerAgent) RenderError(err error) Result {
	if err == jsmvcerror.ForbiddenError {
		c.Result = &ForbiddenResult{}
		return c.Result
	}
	c.Result = &ErrorResult{
		Error: err,
	}
	return c.Result
}
func (c *ControllerAgent) userStaticFileName(fileName string) string {
	return filepath.Join(AppPath, "userstatic", c.Project.Name(), c.UserName, fileName)
}
func (c *ControllerAgent) tempUserStaticFile(prefix string) (*os.File, error) {
	return ioutil.TempFile(c.userStaticFileName("temp"), prefix)
}
func (c *ControllerAgent) tempUserStaticDir(prefix string) (string, error) {
	return ioutil.TempDir(c.userStaticFileName("temp"), prefix)
}
func (c *ControllerAgent) ExportData(dumpName string) (string, error) {
	expFile, err := c.tempUserStaticFile("exp_")
	if err != nil {
		return "", err
	}
	rev, err := filepath.Rel(c.userStaticFileName(""), expFile.Name())
	if err != nil {
		return "", err
	}
	defer expFile.Close()
	err = c.Project.ExportData(dumpName, expFile, c.CurrentGrade)
	if err != nil {
		return "", err
	}
	return rev, nil
}
func (c *ControllerAgent) RenderJson(data map[string]interface{}) Result {
	c.Result = &RenderJsonResult{data}
	return c.Result
}
func (c *ControllerAgent) RenderTemplate(tname string, args map[string]interface{}) Result {
	data := map[string]interface{}{}
	for k, v := range args {
		data[k] = v
	}
	data["c"] = map[string]interface{}{
		"ControllerName": c.ControllerName,
		"ActionName":     c.ActionName,
		"ViewName":       tname,
		"UserName":       c.UserName,
		"Session":        c.Session,
		"CurrentGrade":   c.CurrentGrade,
		"Project": map[string]interface{}{
			"Name":         c.Project.Name(),
			"DisplayLabel": c.Project.DisplayLabel(),
		},
	}
	if c.Title != "" {
		data["Title"] = c.Title
	}
	t, err := c.Project.TemplateSet(c.TemplateFun)
	if err != nil {
		return c.RenderError(err)
	}
	c.Result = &RenderTemplateResult{
		TemplateSet:  t,
		TemplateName: tname,
		RenderArgs:   data,
	}
	return c.Result
}
func (c *ControllerAgent) RenderStaticFile(filename string) Result {
	c.Result = &RenderStaticFileResult{
		ProjectName: c.Project.Name(),
		FileName:    filename,
	}
	return c.Result
}
func (c *ControllerAgent) RenderUserStaticFile(filename string) Result {
	if c.UserName == "" {
		panic(fmt.Errorf("not logged"))
	}
	c.Result = &RenderUserStaticFileResult{
		ProjectName: c.Project.Name(),
		UserName:    c.UserName,
		FileName:    filename,
	}
	return c.Result
}

//Rendering template corresponding to the current action
func (c *ControllerAgent) Render(args map[string]interface{}) Result {
	return c.RenderTemplate(fmt.Sprintf("%s/%s", c.ControllerName, c.ActionName), args)
}

//Combination of URL and authentication
// can:"controller.action" or "controller action"
func (c *ControllerAgent) AuthUrl(args ...string) string {
	url := c.Url(args...)
	if err := c.Session.Set(SessionAuthUrlPrex+url, "1"); err != nil {
		panic(err)
	}
	return url
}
func (c *ControllerAgent) Url(args ...string) string {
	v := []string{}
	if len(args) > 0 && strings.Contains(args[0], ".") {
		v = strings.Split(args[0], ".")
		v = append(v, args[1:]...)
	} else {
		v = args
	}
	url := c.Project.ReverseUrl(v...)
	return url
}

//Check whether the controller is authed or is public
func (c *ControllerAgent) UrlAuthed() bool {
	if c.Public {
		return true
	}
	url := c.Request.URL.Path
	return c.Session.Get(SessionAuthUrlPrex+url) == "1"
}
func (c *ControllerAgent) jsQueryValues(call otto.FunctionCall) otto.Value {
	v, err := c.jsRuntime.ToValue(c.QueryValues())
	if err != nil {
		return otto.NullValue()
	} else {
		return v
	}
}
func (c *ControllerAgent) jsHasResult(call otto.FunctionCall) otto.Value {
	r, _ := otto.ToValue(c.Result != nil)
	return r
}
func (c *ControllerAgent) jsRender(call otto.FunctionCall) otto.Value {
	params := map[string]interface{}{}
	if len(call.ArgumentList) > 0 && call.Argument(0).Class() == "Object" {
		v, err := call.Argument(0).Export()
		if err == nil {
			params = v.(map[string]interface{})
		}
	}
	c.Render(params)
	return otto.NullValue()
}
func (c *ControllerAgent) jsRenderTemplate(call otto.FunctionCall) otto.Value {
	templateName := oftenfun.AssertString(call.Argument(0))
	params := map[string]interface{}{}
	if len(call.ArgumentList) > 1 && call.Argument(1).Class() == "Object" {
		v, err := call.Argument(1).Export()
		if err == nil {
			params = v.(map[string]interface{})
		}
	}
	c.RenderTemplate(templateName, params)
	return otto.NullValue()
}

func (c *ControllerAgent) jsRenderStaticFile(call otto.FunctionCall) otto.Value {
	c.RenderStaticFile(call.Argument(0).String())
	return otto.NullValue()
}
func (c *ControllerAgent) jsRenderJson(call otto.FunctionCall) otto.Value {
	v := oftenfun.AssertObject(call.Argument(0))
	c.RenderJson(v)
	return otto.NullValue()
}
func (c *ControllerAgent) jsUrlAuthed(call otto.FunctionCall) otto.Value {
	v, _ := otto.ToValue(c.UrlAuthed())
	return v
}
func (c *ControllerAgent) jsModel(call otto.FunctionCall) otto.Value {
	mname := oftenfun.AssertString(call.Argument(0))
	return oftenfun.JSToValue(call.Otto, c.Project.Model(mname, c.CurrentGrade).Object())
}
func (c *ControllerAgent) jsGradeCanUse(call otto.FunctionCall) otto.Value {
	byUseGrade := grade.Grade(oftenfun.AssertString(call.Argument(0)))
	return oftenfun.JSToValue(call.Otto, c.CurrentGrade.CanUse(byUseGrade))
}
func (c *ControllerAgent) jsModelChecks(call otto.FunctionCall) otto.Value {
	mname := oftenfun.AssertString(call.Argument(0))
	chks, err := c.Project.Checks(mname, c.CurrentGrade)
	if err != nil {
		panic(err)
	}
	return oftenfun.JSToValue(call.Otto, chks)
}
func (c *ControllerAgent) jsAuthUrl(call otto.FunctionCall) otto.Value {
	strs := make([]string, len(call.ArgumentList))
	for i, v := range call.ArgumentList {
		strs[i] = oftenfun.AssertString(v)
	}
	return oftenfun.JSToValue(call.Otto, c.AuthUrl(strs...))
}
func (c *ControllerAgent) object() map[string]interface{} {

	return map[string]interface{}{
		"ActionName":       c.ActionName,
		"UrlAuthed":        c.jsUrlAuthed,
		"AuthUrl":          c.jsAuthUrl,
		"GradeCanUse":      c.jsGradeCanUse,
		"CurrentGrade":     c.CurrentGrade.String(),
		"ControllerName":   c.ControllerName,
		"JsonBody":         c.JsonBody,
		"TagPath":          c.TagPath,
		"Render":           c.jsRender,
		"RenderTemplate":   c.jsRenderTemplate,
		"RenderStaticFile": c.jsRenderStaticFile,
		"RenderJson":       c.jsRenderJson,
		"HasResult":        c.jsHasResult,
		"Session":          c.Session.Object(),
		"Project":          c.Project.Object(),
		"Method":           c.Request.Method,
		"Model":            c.jsModel,
		"ModelChecks":      c.jsModelChecks,
		"TemplateFunc":     c.jsTemplateFunc,
	}

}

func (c *ControllerAgent) jsTemplateFunc(call otto.FunctionCall) otto.Value {
	if !call.Argument(0).IsObject() {
		panic(jsmvcerror.JSNotIsObject)
	}
	o := call.Argument(0).Object()
	f := c.TemplateFun
	for _, key := range o.Keys() {
		oneFun, err := o.Get(key)
		if err != nil {
			panic(err)
		}
		if !oneFun.IsFunction() {
			continue
		}
		f[key] = func(args ...interface{}) interface{} {
			v, err := oneFun.Call(oneFun, args...)
			if err != nil {
				return err
			}
			var result interface{}
			if v.IsObject() {
				rfmt, err1 := v.Object().Get("fmt")
				rdata, err2 := v.Object().Get("data")

				if err1 != nil || err2 != nil {
					result = template.HTML("")
				} else {
					switch rfmt.String() {
					case "html":
						result = template.HTML(rdata.String())
					case "attr":
						result = template.HTMLAttr(rdata.String())
					case "css":
						result = template.CSS(rdata.String())
					case "js":
						result = template.JS(rdata.String())
					case "jsstr":
						result = template.JSStr(rdata.String())
					default:
						result = template.HTML("")
					}
				}
			} else {
				r, err := v.Export()
				if err != nil {
					return err
				}
				result = r
			}
			return result
		}
		c.TemplateFun = f
	}
	return otto.NullValue()
}
