package main

import (
	"github.com/linlexing/dbgo/grade"
	"github.com/linlexing/dbgo/jsmvcerror"
	"github.com/linlexing/pghelper"
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

func t_project() *grade.DataTable {
	table := grade.NewDataTable("lx_project", grade.GRADE_ROOT)
	table.AddColumn(grade.NewColumn("name", grade.GRADE_ROOT, pghelper.TypeString, true, 50))
	table.AddColumn(grade.NewColumn("dburl", grade.GRADE_ROOT, pghelper.TypeString, true))
	table.AddColumn(grade.NewColumn("owner", grade.GRADE_ROOT, pghelper.TypeString, true))
	table.SetPK("name")
	return table
}
func NewMetaProject(dburl string) (result MetaProject) {
	var err error
	var p Project
	if p, err = NewProject("meta", dburl); err != nil {
		panic(err)
	}

	if err = p.DBHelper().UpdateStruct(t_project()); err != nil {
		panic(err)
	}
	var pBill *grade.DBTable
	if pBill, err = p.DBHelper().Table("lx_project"); err != nil {
		panic(err)
	}

	if err = pBill.FillByID("meta"); err != nil {
		panic(err)
	}
	row := map[string]interface{}{
		"name":  "meta",
		"dburl": dburl,
		"owner": "(system)",
	}
	if pBill.RowCount() == 0 {
		err = pBill.AddRow(row)
	} else {
		err = pBill.UpdateRow(0, row)
	}
	if err != nil {
		panic(err)
	}
	if _, err = pBill.Save(); err != nil {
		panic(err)
	}
	result = &metaProject{project: p.(*project)}
	return
}
func (p *metaProject) NewProject(name string) (result Project, err error) {
	var tab *grade.DataTable
	if tab, err = p.dbHelper.GetDataTable(SQL_GetProject, name); err != nil {
		return
	}
	if tab.RowCount() == 0 {
		err = jsmvcerror.NotFoundProject
		return
	}
	result, err = NewProject(name, tab.GetValue(0, 0).(string))
	return
}
