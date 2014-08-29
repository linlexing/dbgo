package main

import (
	"code.google.com/p/go-uuid/uuid"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/linlexing/dbgo/grade"
	"github.com/linlexing/dbgo/log"
	"github.com/robertkrimen/otto"
	"net/http"
	"path"
	"strings"
)

type Filter func(c *ControllerAgent, filterChain []Filter)

func RouteFilter(c *ControllerAgent, f []Filter) {
	var pname string
	var err error
	var Url string
	if c.ws != nil {
		Url = c.ws.conn.RequestUrl.Path
	} else {
		Url = c.Request.URL.Path
	}
	if Url == "/" {
		pname = DefaultProject
	} else {
		s := strings.Split(Url, "/")
		pname = s[1]
		if len(s) >= 4 {
			c.ControllerName, c.ActionName = s[2], s[3]
		}
		if len(s) > 4 {
			c.TagPath = path.Join(s[4:]...)
			//If the end there is a slash, then remove
			if strings.HasSuffix(c.TagPath, "/") {
				c.TagPath = c.TagPath[:len(c.TagPath)-1]
			}
		}
	}
	c.Project, err = Meta.Project(pname)
	if err != nil {
		c.RenderError(err)
	}
	if c.Result == nil {
		if c.ActionName == "" {
			c.ControllerName, c.ActionName = c.Project.DefaultAction()
		}
	}

	if len(f) > 0 && c.Result == nil {
		f[0](c, f[1:])
	}
}
func ParseJsonFilter(c *ControllerAgent, f []Filter) {
	if c.ws == nil && strings.Contains(c.Request.Header.Get("Content-Type"), "application/json") {
		decoder := json.NewDecoder(c.Request.Body)
		t := map[string]interface{}{}
		decoder.UseNumber()
		err := decoder.Decode(&t)
		if err != nil {
			c.RenderError(err)
		} else {
			c.JsonBody = t
		}
	}
	if len(f) > 0 && c.Result == nil {
		f[0](c, f[1:])
	}
}
func SessionFilter(c *ControllerAgent, f []Filter) {
	var sid string
	if c.ws == nil {
		if cookie, err := c.Request.Cookie("sid"); err != nil {
			if http.ErrNoCookie == err {
				sid = base64.StdEncoding.EncodeToString(uuid.NewRandom())
				cok := &http.Cookie{Name: "sid", Value: sid, Path: "/"}
				http.SetCookie(c.Response, cok)
			} else {
				c.RenderError(err)
			}
		} else {
			sid = cookie.Value

		}
	} else {
		sid = c.ws.conn.SessionID
	}
	c.Session = NewSession(sid, c.Project.Name())

	//取出特殊的Grade属性，该属性用于确定后续的所有拦截器、控制器的可见范围
	//对于未登录用户，赋予特殊GRADE_TAG
	svalue := c.Session.Get("user.dept")
	if svalue != nil {
		dept := svalue.(map[string]interface{})
		c.CurrentGrade = grade.Grade(dept["grade"].(string))
	}
	if c.CurrentGrade == "" {
		c.CurrentGrade = grade.GRADE_TAG
	}
	if len(f) > 0 && c.Result == nil {
		f[0](c, f[1:])
	}
}
func BuildObjectFilter(c *ControllerAgent, f []Filter) {
	var err error
	c.Object, err = c.jsRuntime.ToValue(c.object())
	if err != nil {
		c.RenderError(err)
	}
	if len(f) > 0 && c.Result == nil {
		f[0](c, f[1:])
	}
}
func InterceptFilter(c *ControllerAgent, f []Filter) {
	var filterFuncs []otto.Value
	names, err := c.Project.GetPackageNames("/intercept", c.CurrentGrade)

	if err != nil {
		c.RenderError(err)
		goto end
	}
	filterFuncs = make([]otto.Value, len(names))
	for i, v := range names {
		jsValue := c.Require(v, "/")
		oneIntercept := jsValue.Object()
		whenValue, err := oneIntercept.Get("When")
		if err != nil {
			c.RenderError(err)
			goto end
		}
		when, err := whenValue.ToInteger()
		if err != nil {
			c.RenderError(err)
			goto end
		}
		if when == BEFORE {
			interceptFun, err := oneIntercept.Get("Intercept")
			if err != nil {
				c.RenderError(err)
				goto end
			}
			filterFuncs[i] = interceptFun
		}
	}
	if c.Result == nil && len(filterFuncs) > 0 {

		jsFilterFuncs, err := c.jsRuntime.ToValue(filterFuncs)
		if err != nil {
			c.RenderError(err)
			goto end
		}
		_, err = filterFuncs[0].Call(filterFuncs[0], c.Object, jsFilterFuncs)
		if err != nil {
			c.RenderError(err)
			goto end
		}
	}
end:
	if len(f) > 0 && c.Result == nil {
		f[0](c, f[1:])
	}
}
func LoadControlFilter(c *ControllerAgent, f []Filter) {
	c.ControllerScript = c.Require(path.Join("/controller", c.ControllerName), "/")
	pub, err := c.ControllerScript.Object().Get("Public")
	if err != nil {
		c.Public = false
		goto end
	}
	c.Public, err = pub.ToBoolean()
	if err != nil {
		c.Public = false
	}
end:
	if len(f) > 0 && c.Result == nil {
		f[0](c, f[1:])
	}
}
func UrlAuthFilter(c *ControllerAgent, f []Filter) {
	if c.ws == nil && !c.UrlAuthed() {
		log.WARN.Printf("Forbidden url:%s,sid:%s", c.Request.URL.String(), c.Session.SessionID)
		c.RenderRedirection(c.Project.ReverseUrl(c.Project.DefaultAction()))
	}
	if len(f) > 0 && c.Result == nil {
		f[0](c, f[1:])
	}
}
func UserFilter(c *ControllerAgent, f []Filter) {
	uName := c.Session.Get("user.name")
	if uName != nil {
		c.UserName = uName.(string)
	}

	if len(f) > 0 && c.Result == nil {
		f[0](c, f[1:])
	}
}
func ActionFilter(c *ControllerAgent, f []Filter) {
	actionFunc, err := c.ControllerScript.Object().Get(c.ActionName)
	if err != nil {
		c.RenderError(err)
		goto end
	}
	if actionFunc.IsFunction() {
		_, err = actionFunc.Call(actionFunc, c.Object)
		if err != nil {
			c.RenderError(err)
			goto end
		}
	} else {
		c.RenderError(fmt.Errorf("the action %q not exists", c.ActionName))
		goto end
	}

end:
	if len(f) > 0 && c.Result == nil {
		f[0](c, f[1:])
	}
}
