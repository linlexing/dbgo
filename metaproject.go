package main

import (
	"github.com/linlexing/dbgo/jsmvcerror"
	"github.com/linlexing/dbgo/pghelp"
)

const (
	SQL_GetProject = "select dburl from lx_project where name=$1"
)

type MetaProject interface {
	Project
	NewProject(name string) (Project, error)
}
type metaProject struct {
	*project
}

func t_project() *pghelp.DataTable {
	table := pghelp.NewDataTable("lx_project")
	table.Desc.Grade = GRADE_ROOT
	table.AddColumn(pghelp.NewColumn("name", pghelp.TypeString, true, 50)).Desc.Grade = GRADE_ROOT
	table.AddColumn(pghelp.NewColumn("dburl", pghelp.TypeString, true)).Desc.Grade = GRADE_ROOT
	table.AddColumn(pghelp.NewColumn("owner", pghelp.TypeString, true)).Desc.Grade = GRADE_ROOT
	table.SetPK("name")
	return table
}
func NewMetaProject(dburl string) (result MetaProject) {
	var err error
	var p Project
	if p, err = NewProject("meta", dburl); err != nil {
		panic(err)
	}

	if err = p.DBHelp().UpdateStruct(t_project(), GRADE_ROOT); err != nil {
		panic(err)
	}
	var pBill *Bill
	if pBill, err = p.Bill("lx_project", GRADE_ROOT); err != nil {
		panic(err)
	}
	if err = pBill.FillByID("meta"); err != nil {
		panic(err)
	}
	if err = pBill.UpdateMainRow(map[string]interface{}{
		"name":  "meta",
		"dburl": dburl,
		"owner": "(system)",
	}); err != nil {
		panic(err)
	}
	if err = pBill.Save(); err != nil {
		panic(err)
	}
	result = &metaProject{project: p.(*project)}
	return
}
func (p *metaProject) NewProject(name string) (result Project, err error) {
	var tab *pghelp.DataTable
	if tab, err = p.dbHelp.GetDataTable(SQL_GetProject, name); err != nil {
		return
	}
	if tab.RowCount() == 0 {
		err = jsmvcerror.NotFoundProject
		return
	}
	result, err = NewProject(name, tab.GetValue(0, 0).(string))
	return
}
