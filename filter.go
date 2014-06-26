package main

import (
	"code.google.com/p/go-uuid/uuid"
	"encoding/base64"
	"github.com/linlexing/dbgo/jsmvcerror"
	"github.com/linlexing/dbgo/log"
	"net/http"
	"path"
	"strings"
)

var ()

type Filter func(c *ControllerAgent, filterChain []Filter)

func RouteFilter(c *ControllerAgent, f []Filter) {
	var pname string
	var err error
	if c.Request.URL.Path == "/" {
		pname = DefaultProject
	} else {
		s := strings.Split(c.Request.URL.Path, "/")
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
	c.Title = c.Request.URL.Query().Get("title")
	c.Project, err = MetaCache.Project(pname)
	if err != nil {
		c.RenderError(err)
	}
	if c.Result == nil {
		if c.ActionName == "" {
			c.ControllerName, c.ActionName = c.Project.DefaultAction()
		}
	}
	var ctrl *Controller
	var srcIntercept string
	if c.Result == nil {
		ctrl, err = c.Project.Controller(c.ControllerName, c.ActionName)
		if err != nil {
			c.RenderError(err)
		}
	}
	if c.Result == nil {
		srcIntercept, err = c.Project.InterceptScript(c.CurrentGrade, BEFORE)
		if err != nil {
			c.RenderError(err)
		}
	}
	if c.Result == nil {
		c.script = ctrl.GetScript(srcIntercept, c.CurrentGrade)
		c.Public = ctrl.Public
	}
	if len(f) > 0 && c.Result == nil {
		f[0](c, f[1:])
	}
}

func SessionFilter(c *ControllerAgent, f []Filter) {
	var sid string
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
	c.Session = NewSession(sid, c.Project.Name())
	//取出特殊的Grade属性，该属性用于确定后续的所有拦截器、控制器的可见范围
	//对于未登录用户，赋予特殊GRADE_TAG
	c.CurrentGrade = c.Session.Get("Grade")
	if c.CurrentGrade == "" {
		c.CurrentGrade = GRADE_TAG
		if err := c.Session.Set("Grade", c.CurrentGrade); err != nil {
			c.RenderError(err)
		}
	}
	if len(f) > 0 && c.Result == nil {
		f[0](c, f[1:])
	}
}
func AuthFilter(c *ControllerAgent, f []Filter) {
	if !c.Authed() {
		log.WARN.Printf("Forbidden url:%s\n", c.Request.URL.String())
		c.RenderError(jsmvcerror.ForbiddenError)
	}
	if len(f) > 0 && c.Result == nil {
		f[0](c, f[1:])
	}
}
func ActionFilter(c *ControllerAgent, f []Filter) {
	if c.script != "" {
		args, err := c.jsRuntime.ToValue(c.Object())
		if err != nil {
			c.RenderError(err)
			return
		}
		if _, err := c.jsRuntime.Call(c.script, nil, args, c.ActionName); err != nil {
			if err.Error() == NotFoundAction {
				c.RenderError(err)
			} else {
				c.RenderError(jsmvcerror.NewJavascriptError(c.script, err))
			}
		}
	}
	if len(f) > 0 && c.Result == nil {
		f[0](c, f[1:])
	}
}
