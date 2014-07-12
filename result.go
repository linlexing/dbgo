package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

type Result interface {
	Apply(req *http.Request, w http.ResponseWriter)
}
type NotFoundResult struct{}

func (e *NotFoundResult) Apply(req *http.Request, w http.ResponseWriter) {
	http.Error(w, "can't found the resource.", http.StatusNotFound)
}

type ForbiddenResult struct{}

func (e *ForbiddenResult) Apply(r *http.Request, w http.ResponseWriter) {
	http.Error(w, "can't access the resource,you need login.", http.StatusForbidden)
}

type ErrorResult struct {
	Error error
}

func (e *ErrorResult) Apply(r *http.Request, w http.ResponseWriter) {
	w.Write([]byte(e.Error.Error()))

}

/*
type RenderBill struct {
	Title       string
	BillName    string
	OperateType BillOperateType
	Keys        []interface{}
	AutoFill    map[string]interface{}
	FieldCtrl   map[string]interface{}
	TemplateSet *template.Template
}

func (r *RenderBill) Apply(req *http.Request, w http.ResponseWriter) {
	render := &RenderTemplateResult{
		TemplateSet:  r.TemplateSet,
		TemplateName: "bill",
		RenderArgs:   r.RenerArgs,
	}
	render.Apply(req, w)
}
*/
type RenderHtmlResult struct {
	html string
}

func (r *RenderHtmlResult) Apply(req *http.Request, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(r.html))
}

// Action methods return this result to request a template be rendered.
type RenderTemplateResult struct {
	TemplateSet  *template.Template
	TemplateName string
	RenderArgs   map[string]interface{}
}

func (r *RenderTemplateResult) Apply(req *http.Request, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.TemplateSet.Lookup(r.TemplateName) == nil {
		w.Write([]byte(fmt.Sprintf("can't found template:%s\n",
			r.TemplateName)))
		return
	}
	err := r.TemplateSet.ExecuteTemplate(w, r.TemplateName, r.RenderArgs)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("render template:%s error:%s\n",
			r.TemplateName, err)))
	}
}

type RenderStaticFileResult struct {
	ProjectName string
	FileName    string
}

func (r *RenderStaticFileResult) Apply(req *http.Request, w http.ResponseWriter) {
	http.ServeFile(w, req, filepath.Join(AppPath, "static", r.ProjectName, r.FileName))
}

type RenderUserStaticFileResult struct {
	ProjectName string
	UserName    string
	FileName    string
}

func (r *RenderUserStaticFileResult) Apply(req *http.Request, w http.ResponseWriter) {
	http.ServeFile(w, req, filepath.Join(AppPath, "userstatic", r.ProjectName, r.UserName, r.FileName))
}

type RenderJsonResult struct {
	obj map[string]interface{}
}

func (r *RenderJsonResult) Apply(req *http.Request, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	encoder := json.NewEncoder(w)

	if err := encoder.Encode(r.obj); err != nil {
		(&ErrorResult{Error: err}).Apply(req, w)
		return
	}
}
