package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/linlexing/dbgo/log"
	"github.com/linlexing/dbgo/oftenfun"
	"path/filepath"

	"github.com/robertkrimen/otto"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Iterator struct {
	projectName string
	dbIterator  iterator.Iterator
}

func (i *Iterator) jsNext(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, i.dbIterator.Next())
}

func (i *Iterator) jsRelease(call otto.FunctionCall) otto.Value {
	i.dbIterator.Release()
	return otto.UndefinedValue()
}
func (i *Iterator) jsKey(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, string(i.dbIterator.Key())[len(i.projectName)+1:])
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
		"Next":    i.jsNext,
		"Key":     i.jsKey,
		"Value":   i.jsValue,
		"Release": i.jsRelease,
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
	if err == leveldb.ErrNotFound {
		return otto.NullValue()
	}
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
func (c *AppCache) jsHasPrex(call otto.FunctionCall) otto.Value {
	iter := c.hub.db.NewIterator(util.BytesPrefix([]byte(c.projectName+"|"+oftenfun.AssertString(call.Argument(0)))), nil)
	defer iter.Release()
	rev := iter.Next()
	return oftenfun.JSToValue(call.Otto, rev)
}
func (c *AppCache) jsCount(call otto.FunctionCall) otto.Value {
	iter := c.hub.db.NewIterator(util.BytesPrefix([]byte(c.projectName+"|"+oftenfun.AssertString(call.Argument(0)))), nil)
	defer iter.Release()
	rev := 0
	for ok := iter.Next(); ok; ok = iter.Next() {
		rev++
	}
	return oftenfun.JSToValue(call.Otto, rev)
}
func (c *AppCache) jsDelete(call otto.FunctionCall) otto.Value {
	err := c.hub.db.Delete([]byte(c.projectName+"|"+oftenfun.AssertString(call.Argument(0))), nil)
	if err == nil || err == leveldb.ErrNotFound {
		return otto.NullValue()
	} else {
		panic(err)
	}
}
func (c *AppCache) jsSet(call otto.FunctionCall) otto.Value {
	key := oftenfun.AssertString(call.Argument(0))
	value := oftenfun.AssertValue(call.Argument(1))[0]
	if err := c.hub.db.Put([]byte(c.projectName+"|"+key), encodeVal(value), nil); err != nil {
		panic(err)
	}
	return call.Argument(1)
}
func encodeVal(value interface{}) []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf) // Will write to network.
	if err := enc.Encode(&value); err != nil {
		panic(err)
	}
	return buf.Bytes()
}
func (c *AppCache) jsBatchWrite(call otto.FunctionCall) otto.Value {
	arr := oftenfun.AssertArray(call.Argument(0))
	batch := new(leveldb.Batch)
	for _, v := range arr {
		switch tv := v.(type) {
		case map[string]interface{}:
			switch tv["opt"].(string) {
			case "set":
				batch.Put([]byte(c.projectName+"|"+tv["key"].(string)), encodeVal(tv["value"]))
			case "delete":
				batch.Delete([]byte(c.projectName + "|" + tv["key"].(string)))
			default:
				panic(fmt.Errorf("invalid opt:%v", tv["opt"]))
			}

		default:
			panic(fmt.Errorf("the value %#v(%T) not is object", v))
		}
	}
	if err := c.hub.db.Write(batch, nil); err != nil {
		panic(err)
	}
	return otto.UndefinedValue()
}
func (c *AppCache) jsPrexIterator(call otto.FunctionCall) otto.Value {
	obj := &Iterator{
		c.projectName,
		c.hub.db.NewIterator(
			util.BytesPrefix(
				[]byte(c.projectName+"|"+oftenfun.AssertString(call.Argument(0)))),
			nil),
	}
	return oftenfun.JSToValue(call.Otto, obj.Object())
}
func (c *AppCache) Object() map[string]interface{} {
	return map[string]interface{}{
		"HasPrex":      c.jsHasPrex,
		"BatchWrite":   c.jsBatchWrite,
		"Count":        c.jsCount,
		"Get":          c.jsGet,
		"Set":          c.jsSet,
		"Delete":       c.jsDelete,
		"PrexIterator": c.jsPrexIterator,
	}
}
