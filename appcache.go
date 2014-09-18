package main

import (
	"bytes"
	"encoding/gob"
	"github.com/linlexing/dbgo/log"
	"github.com/linlexing/dbgo/oftenfun"
	"path/filepath"

	"github.com/robertkrimen/otto"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Iterator struct {
	dbIterator iterator.Iterator
}

func (i *Iterator) jsNext(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, i.dbIterator.Next())
}
func (i *Iterator) jsKey(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, string(i.dbIterator.Key()))
}
func (i *Iterator) jsValue(call otto.FunctionCall) otto.Value {
	buf := i.dbIterator.Value()
	dec := gob.NewDecoder(bytes.NewBuffer(buf))
	var rev interface{}
	if err := dec.Decode(&rev); err != nil {
		panic(err)
	}
	return oftenfun.JSToValue(call.Otto, rev)
}
func (i *Iterator) Object() map[string]interface{} {
	return map[string]interface{}{
		"Next":  i.jsNext,
		"Key":   i.jsKey,
		"Value": i.jsValue,
	}
}

type AppCacheHub struct {
	db *leveldb.DB
}
type AppCache struct {
	hub         *AppCacheHub
	projectName string
}

func NewAppCacheHub() *AppCacheHub {
	db, err := leveldb.OpenFile(filepath.Join(AppPath, "pcache"), nil)
	if err != nil {
		panic(err)
	}

	return &AppCacheHub{db}
}
func (c *AppCacheHub) AppCache(projectName string) *AppCache {
	//clear prev cache
	batch := new(leveldb.Batch)

	iter := c.db.NewIterator(util.BytesPrefix([]byte(projectName+"|")), nil)
	defer iter.Release()
	icount := int64(0)
	for iter.Next() {
		batch.Delete(iter.Key())
		icount++
	}
	if err := c.db.Write(batch, nil); err != nil {
		panic(err)
	}
	if icount > 0 {
		log.TRACE.Printf("clear appcache:%s,count:%v", projectName, icount)
	}
	return &AppCache{c, projectName}
}
func (c *AppCache) jsGet(call otto.FunctionCall) otto.Value {
	buf, err := c.hub.db.Get([]byte(c.projectName+"|"+oftenfun.AssertString(call.Argument(0))), nil)
	if err != nil {
		panic(err)
	}
	dec := gob.NewDecoder(bytes.NewBuffer(buf))
	var rev interface{}
	if err = dec.Decode(&rev); err != nil {
		panic(err)
	}
	return oftenfun.JSToValue(call.Otto, rev)
}
func (c *AppCache) jsPut(call otto.FunctionCall) otto.Value {
	key := oftenfun.AssertString(call.Argument(0))
	value := oftenfun.AssertObject(call.Argument(1))
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf) // Will write to network.
	if err := enc.Encode(&value); err != nil {
		panic(err)
	}
	if err := c.hub.db.Put([]byte(c.projectName+"|"+key), buf.Bytes(), nil); err != nil {
		panic(err)
	}
	return call.Argument(1)
}
func (c *AppCache) jsPrexIterator(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, Iterator{
		c.hub.db.NewIterator(
			util.BytesPrefix(
				[]byte(c.projectName+"|"+oftenfun.AssertString(call.Argument(0)))),
			nil),
	})
}
func (c *AppCache) Object() map[string]interface{} {
	return map[string]interface{}{
		"Get":          c.jsGet,
		"Put":          c.jsPut,
		"PrexIterator": c.jsPrexIterator,
	}
}
