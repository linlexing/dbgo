package main

import (
	"sync"
)

const (
	META            = "meta"
	META_Repository = "root/meta"
)

type Cache struct {
	lock         *sync.Mutex
	projectCache map[string]Project
}

func NewCache(dburl string) *Cache {
	return &Cache{
		lock:         &sync.Mutex{},
		projectCache: map[string]Project{META: NewMetaProject(dburl, META_Repository)}}

}
func (c *Cache) Project(name string) (result Project, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	var ok bool
	if result, ok = c.projectCache[name]; ok {
		return
	}

	meta := c.projectCache[META]
	if result, err = meta.(MetaProject).NewProject(name); err != nil {
		return
	}
	c.projectCache[name] = result
	return

}
