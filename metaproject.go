package main

import (
	"github.com/linlexing/dbgo/grade"
	"github.com/linlexing/dbgo/jsmvcerror"
	"github.com/linlexing/dbgo/oftenfun"
)

type MetaProject interface {
	Project
	NewProject(name string) (Project, error)
}
type metaProject struct {
	*project
}

func NewMetaProject(dburl string) (result MetaProject) {
	var err error
	var p Project
	if p, err = NewProject("meta", "meta project", dburl, grade.GRADE_ROOT.Child("meta").String()); err != nil {
		panic(err)
	}
	result = &metaProject{project: p.(*project)}
	if err = result.ReloadRepository(); err != nil {
		panic(err)
	}
	return
}
func (p *metaProject) NewProject(name string) (result Project, err error) {
	table := p.Model("lx_project", grade.GRADE_ROOT.Child("meta"))
	if err = table.FillByID(name); err != nil {
		return
	}
	if table.RowCount() == 0 {
		err = jsmvcerror.NotFoundProject
		return
	}
	row := table.Row(0)
	result, err = NewProject(name, oftenfun.SafeToString(row["displaylabel"]), row["dburl"].(string), oftenfun.SafeToString(row["repository"]))
	if err != nil {
		return
	}
	err = result.ReloadRepository()
	return
}
