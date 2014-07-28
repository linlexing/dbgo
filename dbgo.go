package main

/*
import (
	"github.com/linlexing/pghelper"
	"sync"
)

const ()

type DBGo struct {
	lock         *sync.Mutex
	dburl        pghelper.PGUrl
	dbhelper     *pghelper.PGHelper
	projectCache map[string]Project
}

func NewDBGo(dburl string) *DBGo {
	rev := &DBGo{
		lock:         &sync.Mutex{},
		dburl:        pghelper.NewPGUrl(dburl),
		dbhelper:     pghelper.NewPGHelper(dburl),
		projectCache: map[string]Project{},
	}
	return rev
}
func (d *DBGo) CreateProject(name, pwd string) (string, error) {
	err := d.dbhelper.CreateSchema(name, pwd)
	if err != nil {
		return "", err
	}
	tmps := pghelper.PGUrl{}
	tmps.Parse(d.dburl.String())
	tmps["user"] = name
	tmps["password"] = pwd
	return tmps.String(), nil
}
func (d *DBGo) ReadyMetaProject(dburl string) {
	d.projectCache["meta"] = NewMetaProject(dburl)
	return
}
func (c *DBGo) Project(name string) (result Project, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	var ok bool
	if result, ok = c.projectCache[name]; ok {
		return
	}
	if result, err = c.projectCache["meta"].(MetaProject).NewProject(name); err != nil {
		return
	}
	c.projectCache[name] = result
	return

}
*/
