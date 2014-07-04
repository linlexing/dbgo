package main

import (
	"github.com/linlexing/dbgo/grade"
	"github.com/linlexing/dbgo/jsmvcerror"
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/linlexing/pghelper"
)

type MetaProject interface {
	Project
	NewProject(name string) (Project, error)
}
type metaProject struct {
	*project
}

func lx_project() *grade.DataTable {
	table := grade.NewDataTable("lx_project", grade.GRADE_ROOT)
	table.AddColumn(grade.NewColumn("name", grade.GRADE_ROOT, pghelper.TypeString, true, 50))
	table.AddColumn(grade.NewColumn("dburl", grade.GRADE_ROOT, pghelper.TypeString, true))
	table.AddColumn(grade.NewColumn("owner", grade.GRADE_ROOT, pghelper.TypeString, true))
	table.AddColumn(grade.NewColumn("repository", grade.GRADE_ROOT, pghelper.TypeString, false))
	table.SetPK("name")
	return table
}
func NewMetaProject(dburl, repository string) (result MetaProject) {
	var err error
	var p Project
	if p, err = NewProject("meta", dburl, repository); err != nil {
		panic(err)
	}

	if err = p.DBHelper().UpdateStruct(lx_project()); err != nil {
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
		"name":       "meta",
		"dburl":      dburl,
		"owner":      "(system)",
		"repository": "root/meta",
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
	table := p.Model("lx_project", grade.GRADE_ROOT)
	if err = table.FillByID(name); err != nil {
		return
	}
	if table.RowCount() == 0 {
		err = jsmvcerror.NotFoundProject
		return
	}
	row := table.GetRow(0)
	result, err = NewProject(name, row["dburl"].(string), oftenfun.SafeToString(row["repository"]))
	if err != nil {
		return
	}
	result.ReloadRepository()
	return
}
