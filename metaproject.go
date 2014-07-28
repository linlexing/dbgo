package main

import (
	"github.com/linlexing/dbgo/grade"
	"github.com/linlexing/dbgo/jsmvcerror"
	"github.com/linlexing/dbgo/oftenfun"
	"sync"
)

type MetaProject interface {
	Project
	Project(name string) (Project, error)
}
type metaProject struct {
	*project
	lock         *sync.Mutex
	projectCache map[string]Project
}

func NewMetaProject(dburl string) (result MetaProject) {
	var err error
	p := NewProject("meta", TranslateString{"en": "meta project", "zh_CN": "元数据库"}, dburl, grade.GRADE_ROOT.Child("meta").String())
	result = &metaProject{project: p.(*project)}
	if err = result.ReloadRepository(); err != nil {
		panic(err)
	}
	return
}

func (p *metaProject) loadProject(name string) (result Project, err error) {
	table := p.DBModel("lx_project", grade.GRADE_ROOT.Child("meta"))
	if err := table.DBHelper().Open(); err != nil {
		return nil, err
	}
	defer table.DBHelper().Close()

	if err = table.FillByID(name); err != nil {
		return
	}
	if table.RowCount() == 0 {
		err = jsmvcerror.NotFoundProject
		return
	}
	row := table.Row(0)
	label := TranslateString{}
	if row["label_en"] != nil {
		label["en"] = row["label_en"].(string)
	}
	if row["label_zh_cn"] != nil {
		label["zh_CN"] = row["label_zh_cn"].(string)
	}
	result = NewProject(name, label, row["dburl"].(string), oftenfun.SafeToString(row["repository"]))
	err = result.ReloadRepository()
	return
}
func (m *metaProject) Project(name string) (result Project, err error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	var ok bool
	if result, ok = m.projectCache[name]; ok {
		return
	}
	if result, err = m.loadProject(name); err != nil {
		return
	}
	m.projectCache[name] = result
	return

}
