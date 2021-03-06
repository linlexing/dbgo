package main

import (
	"encoding/json"
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"github.com/linlexing/dbgo/grade"
	"github.com/linlexing/dbgo/jsmvcerror"
	"github.com/linlexing/dbgo/log"
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/robertkrimen/otto"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"time"
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
	Public           bool        //Indicate whether you can access without authentication
	CurrentGrade     grade.Grade //When this call, the user's Grade properties. After obtaining the value from the Session, and not to change it, even if the user changes their Grade other requests
	Result           Result
	TemplateFun      template.FuncMap
	Object           otto.Value
	ControllerScript otto.Value
	Tag              map[string]interface{}
	ws               *WSAgent
	jsRuntime        *otto.Otto
}

func (c *ControllerAgent) MapFileName(url string) string {
	if !strings.HasPrefix(url, "/") {
		url = path.Join("/", url)
	}
	paths := strings.Split(url, "/")[1:]
	projectName := paths[0]
	if projectName == "public" {
		return filepath.Join(AppPath, url)
	} else {
		controlName := paths[1]
		actionName := paths[2]
		tagPath := strings.Join(paths[3:], "/")
		if controlName == "static" && actionName == "file" {
			return filepath.Join(AppPath, "static", projectName, tagPath)
		}
		if controlName == "userfile" && actionName == "file" {
			return filepath.Join(AppPath, "userfile", projectName, c.UserName, tagPath)
		}
		panic(fmt.Errorf("can't map the file:%s", url))
	}
}
func (c *ControllerAgent) Require(fileName, currentModuleDir string) (otto.Value, error) {

	script, moduleFileName, err := c.Project.Require(c.jsRuntime, fileName, currentModuleDir, c.CurrentGrade)
	if err != nil {
		return otto.UndefinedValue(), err
	}
	jsfun, err := c.jsRuntime.Run(script)
	if err != nil {
		return otto.UndefinedValue(), err
	}

	tmpValue, err := c.jsRuntime.ToValue(map[string]interface{}{
		"exports":    map[string]interface{}{},
		"__dirname":  path.Dir(moduleFileName),
		"__filename": path.Base(moduleFileName),
		"require": func(call otto.FunctionCall) otto.Value {
			rev, err := c.Require(oftenfun.AssertString(call.Argument(0)), path.Dir(moduleFileName))
			if err != nil {
				panic(err)
			}
			return rev
		},
		"safeRequire": func(call otto.FunctionCall) otto.Value {
			rev, err := c.Require(oftenfun.AssertString(call.Argument(0)), path.Dir(moduleFileName))
			switch err.(type) {
			case nil:
				return rev
			case *EmptyPackageError:
				return otto.NullValue()
			default:
				panic(err)
			}
		},
	})
	if err != nil {
		return otto.UndefinedValue(), err
	}
	jsModule := tmpValue.Object()
	exports, _ := jsModule.Get("exports")
	rev, err := jsfun.Call(exports, jsModule)
	if err != nil {
		return otto.UndefinedValue(), err
	}
	if rev.IsUndefined() {
		exports, _ := jsModule.Get("exports")
		return exports, nil
	} else {
		return rev, nil
	}
}
func (c *ControllerAgent) QueryValues() url.Values {
	if c.ws == nil {
		return c.Request.URL.Query()
	} else {
		return c.ws.conn.RequestUrl.Query()
	}
}

type packConfig struct {
	Files []string
	Times []time.Time
}

func NewAgentWS(ws *WSAgent) *ControllerAgent {
	return &ControllerAgent{
		ws:  ws,
		Tag: map[string]interface{}{},
	}
}
func NewAgent(w http.ResponseWriter, r *http.Request) *ControllerAgent {

	c := &ControllerAgent{
		Request:  r,
		Response: w,
		Tag:      map[string]interface{}{},
	}
	c.TemplateFun = template.FuncMap{
		"url":     c.Url,
		"authUrl": c.AuthUrl,
		"packFiles": func(dest string, srcFiles []interface{}) string {

			//read config
			config := packConfig{}
			configFileName := c.MapFileName(dest) + ".cfg"
			bys, err := ioutil.ReadFile(configFileName)
			if err != nil && !os.IsNotExist(err) {
				panic(err)
			} else if err == nil && len(bys) > 0 {

				err = json.Unmarshal(bys, &config)
				if err != nil {
					panic(err)
				}
			}

			//build the Config
			newConfig := packConfig{}
			newConfig.Files = make([]string, len(srcFiles))
			newConfig.Times = make([]time.Time, len(srcFiles))
			for i, v := range srcFiles {
				fileInfo, err := os.Stat(c.MapFileName(v.(string)))
				if err != nil {
					panic(err)
				}
				newConfig.Files[i] = v.(string)
				newConfig.Times[i] = fileInfo.ModTime()
			}
			//if modify,pack it
			if !reflect.DeepEqual(config, newConfig) {
				file, err := os.Create(c.MapFileName(dest))
				if err != nil {
					panic(err)
				}
				defer file.Close()
				for _, v := range srcFiles {
					buf, err := ioutil.ReadFile(c.MapFileName(v.(string)))
					if err != nil {
						panic(err)
					}
					_, err = file.Write(buf)
					if err != nil {
						panic(err)
					}
					_, err = file.WriteString("\n")
					if err != nil {
						panic(err)
					}
				}

				buffer, err := json.Marshal(newConfig)
				if err != nil {
					panic(err)
				}
				err = ioutil.WriteFile(configFileName, buffer, os.ModePerm)
				if err != nil {
					panic(err)
				}
			}
			return dest
		},
		"userFile": func(filename string) string {
			return c.AuthUrl("userfile.file", filename)
		},
	}

	return c
}

/*func (c *ControllerAgent) ConvertFillValue(fillValue string, args map[string]interface{}) string {
	if fillValue == "" {
		return ""
	}
	tmpl, err := template.New("fill").Parse(fillValue)
	if err != nil {
		panic(fmt.Errorf("convert fill value %q error:%s", fillValue, err))
	}
	rev := &bytes.Buffer{}
	if args == nil {
		args = map[string]interface{}{}
	}
	args["Project"] = c.Project
	args["c"] = c
	if err = tmpl.Execute(rev, args); err != nil {
		panic(fmt.Errorf("convert fill value %q error:%s", fillValue, err))
	}
	return rev.String()
}*/
func (c *ControllerAgent) RenderForbidden() Result {
	c.Result = &ForbiddenResult{}
	return c.Result
}

func (c *ControllerAgent) RenderRedirection(url string) Result {
	c.Result = &RedirectionResult{url}
	return c.Result
}
func (c *ControllerAgent) RenderError(err error) Result {
	if err == jsmvcerror.ForbiddenError {
		c.Result = &ForbiddenResult{c.ws}
		return c.Result
	}
	c.Result = &ErrorResult{
		Error: err,
		ws:    c.ws,
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
func (c *ControllerAgent) TemplateExists(tname string) bool {
	t, err := c.Project.TemplateSet(c.TemplateFun)
	if err != nil {
		panic(err)
	}
	return t.Lookup(tname) != nil
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
		"Tag":            c.Tag,
		"Project": map[string]interface{}{
			"Name":         c.Project.Name(),
			"DisplayLabel": c.Project.DisplayLabel(),
		},
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
func (c *ControllerAgent) RenderUserFile(filename string) Result {
	if c.UserName == "" {
		panic(fmt.Errorf("not logged"))
	}
	c.Result = &RenderUserFileResult{
		ProjectName: c.Project.Name(),
		UserName:    c.UserName,
		FileName:    filename,
	}
	return c.Result
}

//Rendering template corresponding to the current action
func (c *ControllerAgent) Render(args map[string]interface{}) Result {
	return c.RenderTemplate(fmt.Sprintf("%s/%s.html", c.ControllerName, c.ActionName), args)
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
	url := c.Request.URL.String()
	return c.Session.Get(SessionAuthUrlPrex+url) == "1"
}
func (c *ControllerAgent) jsQueryValues(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) > 0 {
		vName := oftenfun.AssertString(call.Argument(0))
		return oftenfun.JSToValue(call.Otto, c.QueryValues().Get(vName))
	} else {
		v, err := c.jsRuntime.ToValue(c.QueryValues())
		if err != nil {
			return otto.NullValue()
		} else {
			return v
		}
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
	return otto.UndefinedValue()
}

/*func (c *ControllerAgent) jsConvertFillValue(call otto.FunctionCall) otto.Value {
	fillValue := oftenfun.AssertString(call.Argument(0))
	args := oftenfun.AssertObject(call.Argument(1))
	return oftenfun.JSToValue(call.Otto, c.ConvertFillValue(fillValue, args))
}*/

func (c *ControllerAgent) jsRenderTemplate(call otto.FunctionCall) otto.Value {
	templateName := oftenfun.AssertString(call.Argument(0))
	params := map[string]interface{}{}
	if len(call.ArgumentList) > 1 {
		params = oftenfun.AssertObject(call.Argument(1))
	}
	c.RenderTemplate(templateName, params)
	return otto.UndefinedValue()
}
func (c *ControllerAgent) jsRenderRedirection(call otto.FunctionCall) otto.Value {
	c.RenderRedirection(call.Argument(0).String())
	return otto.UndefinedValue()
}

func (c *ControllerAgent) jsRenderStaticFile(call otto.FunctionCall) otto.Value {
	c.RenderStaticFile(call.Argument(0).String())
	return otto.UndefinedValue()
}
func (c *ControllerAgent) jsRenderUserFile(call otto.FunctionCall) otto.Value {
	c.RenderUserFile(call.Argument(0).String())
	return otto.UndefinedValue()
}
func (c *ControllerAgent) jsRenderError(call otto.FunctionCall) otto.Value {
	log.WARN.Println("js render error:", call.Argument(0).String())
	c.RenderError(fmt.Errorf(call.Argument(0).String()))
	return otto.UndefinedValue()
}
func (c *ControllerAgent) jsRenderJson(call otto.FunctionCall) otto.Value {
	v := oftenfun.AssertObject(call.Argument(0))
	c.RenderJson(v)
	return otto.UndefinedValue()
}
func (c *ControllerAgent) jsUrlAuthed(call otto.FunctionCall) otto.Value {
	v, _ := otto.ToValue(c.UrlAuthed())
	return v
}
func (c *ControllerAgent) jsUrl(call otto.FunctionCall) otto.Value {
	strs := make([]string, len(call.ArgumentList))
	for i, v := range call.ArgumentList {
		strs[i] = oftenfun.AssertString(v)
	}
	return oftenfun.JSToValue(call.Otto, c.Url(strs...))
}
func (c *ControllerAgent) jsDBModel(call otto.FunctionCall) otto.Value {
	gradestr := c.CurrentGrade
	tnames := []string{}
	for _, v := range call.ArgumentList {
		switch v.Class() {
		case "GoArray", "Array":
			tnames = append(tnames, oftenfun.AssertStringArray(v)...)
		default:
			tnames = append(tnames, oftenfun.AssertString(v))
		}
	}
	rev := make([]interface{}, len(tnames))
	for i, v := range c.Project.DBModel(gradestr, tnames...) {
		rev[i] = v.Object()
	}
	return oftenfun.JSToValue(call.Otto, rev)
}
func (c *ControllerAgent) jsTemplateExists(call otto.FunctionCall) otto.Value {
	tname := oftenfun.AssertString(call.Argument(0))
	return oftenfun.JSToValue(call.Otto, c.TemplateExists(tname))
}
func (c *ControllerAgent) jsGradeCanUse(call otto.FunctionCall) otto.Value {
	byUseGrade := grade.Grade(oftenfun.AssertString(call.Argument(0)))
	return oftenfun.JSToValue(call.Otto, c.CurrentGrade.CanUse(byUseGrade))
}
func (c *ControllerAgent) jsSetTag(call otto.FunctionCall) otto.Value {
	vname := oftenfun.AssertString(call.Argument(0))
	vv := oftenfun.AssertValue(call.Argument(1))[0]
	c.Tag[vname] = vv
	return call.Argument(1)
}
func (c *ControllerAgent) jsGetTag(call otto.FunctionCall) otto.Value {
	vname := oftenfun.AssertString(call.Argument(0))

	return oftenfun.JSToValue(call.Otto, c.Tag[vname])
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
func (c *ControllerAgent) jsUserName(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) > 0 {
		c.UserName = oftenfun.AssertString(call.Argument(0))
		if err := c.Session.Set("user.name", c.UserName); err != nil {
			panic(err)
		}
	}
	return oftenfun.JSToValue(call.Otto, c.UserName)
}
func (c *ControllerAgent) jsBroadcast(call otto.FunctionCall) otto.Value {
	mes := SocketMessage{
		oftenfun.AssertString(call.Argument(0)),
		oftenfun.AssertString(call.Argument(1)),
	}
	SocketHub.broadcast <- mes
	return otto.UndefinedValue()
}
func (c *ControllerAgent) jsChannelExists(call otto.FunctionCall) otto.Value {
	url := oftenfun.AssertString(call.Argument(0))
	_, ok := SocketHub.connections[url]
	return oftenfun.JSToValue(call.Otto, ok)
}
func (c *ControllerAgent) package_UserFile() map[string]interface{} {
	return map[string]interface{}{
		"FileExists": func(call otto.FunctionCall) otto.Value {
			fileName := oftenfun.AssertString(call.Argument(0))
			_, err := os.Stat(filepath.Join(AppPath, "userfile", c.Project.Name(), c.UserName, fileName))
			rev := err == nil
			return oftenfun.JSToValue(call.Otto, rev)
		},
		"ReadFile": func(call otto.FunctionCall) otto.Value {
			fileName := oftenfun.AssertString(call.Argument(0))
			bys, err := ioutil.ReadFile(filepath.Join(AppPath, "userfile", c.Project.Name(), c.UserName, fileName))
			if err != nil {
				panic(err)
			}
			return oftenfun.JSToValue(call.Otto, bys)
		},
		"ReadFileStr": func(call otto.FunctionCall) otto.Value {
			fileName := oftenfun.AssertString(call.Argument(0))
			bys, err := ioutil.ReadFile(filepath.Join(AppPath, "userfile", c.Project.Name(), c.UserName, fileName))
			if err != nil {
				panic(err)
			}
			return oftenfun.JSToValue(call.Otto, string(bys))
		},
		"WriteFile": func(call otto.FunctionCall) otto.Value {
			fileName := oftenfun.AssertString(call.Argument(0))
			bys := oftenfun.AssertByteArray(call.Argument(1))
			fileName = filepath.Join(AppPath, "userfile", c.Project.Name(), c.UserName, fileName)
			err := os.MkdirAll(filepath.Dir(fileName), os.ModePerm)
			if err != nil {
				panic(err)
			}
			err = ioutil.WriteFile(fileName, bys, os.ModePerm)
			if err != nil {
				panic(err)
			}
			return otto.UndefinedValue()
		},
		"WriteFileStr": func(call otto.FunctionCall) otto.Value {
			fileName := oftenfun.AssertString(call.Argument(0))
			bys := oftenfun.AssertString(call.Argument(1))
			fileName = filepath.Join(AppPath, "userfile", c.Project.Name(), c.UserName, fileName)
			err := os.MkdirAll(filepath.Dir(fileName), os.ModePerm)
			if err != nil {
				panic(err)
			}
			err = ioutil.WriteFile(fileName, []byte(bys), os.ModePerm)
			if err != nil {
				panic(err)
			}
			return otto.UndefinedValue()
		},
	}
}
func (c *ControllerAgent) object() map[string]interface{} {
	var remoteAddr string
	if c.ws == nil {
		remoteAddr = c.Request.RemoteAddr
	} else {
		remoteAddr = c.ws.conn.ws.RemoteAddr().String()
	}
	rev := map[string]interface{}{
		"ActionName":     c.ActionName,
		"AuthUrl":        c.jsAuthUrl,
		"Broadcast":      c.jsBroadcast,
		"ChannelExists":  c.jsChannelExists,
		"CurrentGrade":   c.CurrentGrade.String(),
		"ControllerName": c.ControllerName,
		"ClientAddr":     remoteAddr,
		//"ConvertFillValue":  c.jsConvertFillValue,
		"UrlAuthed":         c.jsUrlAuthed,
		"GradeCanUse":       c.jsGradeCanUse,
		"GetTag":            c.jsGetTag,
		"JsonBody":          c.JsonBody,
		"TagPath":           c.TagPath,
		"QueryValues":       c.jsQueryValues,
		"Render":            c.jsRender,
		"RenderError":       c.jsRenderError,
		"RenderTemplate":    c.jsRenderTemplate,
		"RenderStaticFile":  c.jsRenderStaticFile,
		"RenderJson":        c.jsRenderJson,
		"RenderUserFile":    c.jsRenderUserFile,
		"RenderRedirection": c.jsRenderRedirection,
		"HasResult":         c.jsHasResult,
		"Session":           c.Session.Object(),
		"SetTag":            c.jsSetTag,
		"Project":           c.Project.Object(),
		//"Model":            c.jsModel,
		"DBModel":        c.jsDBModel,
		"ModelChecks":    c.jsModelChecks,
		"TemplateFunc":   c.jsTemplateFunc,
		"TemplateExists": c.jsTemplateExists,
		"UserFile":       c.package_UserFile(),
		"UserName":       c.jsUserName,
		"Url":            c.jsUrl,
	}
	if c.ws != nil {
		rev["ws"] = c.ws.Object()
	} else {
		rev["Method"] = c.Request.Method

	}
	return rev
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
