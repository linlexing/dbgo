package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"github.com/linlexing/dbgo/log"
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/robertkrimen/otto"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"time"
)

const (
	SessionLWTPrex = "lwt."
)

type SessionManager struct {
	timeout int
	db      *leveldb.DB
}

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register(json.Number(""))
	gob.Register(time.Now())
}
func NewSessionManager(dbpath string, timeout int) *SessionManager {
	db, err := leveldb.OpenFile(dbpath, nil)
	if err != nil {
		panic(err)
	}
	return &SessionManager{timeout: timeout, db: db}
}
func (s *SessionManager) Close() error {
	return s.db.Close()
}
func (s *SessionManager) Put(pname, sid, key string, value interface{}) error {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf) // Will write to network.
	if err := enc.Encode(&value); err != nil {
		return err
	}

	if err := s.db.Put([]byte(SessionLWTPrex+pname+sid), time2Bytes(time.Now()), nil); err != nil {
		return err
	}

	return s.db.Put([]byte(pname+sid+key), buf.Bytes(), nil)
}
func (s *SessionManager) Get(pname, sid, key string) interface{} {
	if err := s.db.Put([]byte(SessionLWTPrex+pname+sid), time2Bytes(time.Now()), nil); err != nil {
		panic(err)
	}
	buf, err := s.db.Get([]byte(pname+sid+key), nil)
	if err != nil && err != leveldb.ErrNotFound {
		panic(err)
	}
	if len(buf) > 0 {
		dec := gob.NewDecoder(bytes.NewBuffer(buf))
		var rev interface{}
		if err := dec.Decode(&rev); err != nil {
			panic(err)
		}
		return rev
	} else {
		return nil
	}
}
func (s *SessionManager) All() (map[string]string, error) {
	rev := map[string]string{}
	iter := s.db.NewIterator(nil, nil)
	for iter.Next() {
		// Remember that the contents of the returned slice should not be modified, and
		// only valid until the next call to Next.
		rev[string(iter.Key())] = string(iter.Value())
	}
	iter.Release()
	return rev, iter.Error()
}
func bytes2Time(b []byte) time.Time {
	result := time.Time{}
	if err := result.UnmarshalBinary(b); err != nil {
		result = time.Unix(0, 0)
	}
	return result
}
func time2Bytes(t time.Time) []byte {
	if bys, err := t.MarshalBinary(); err != nil {
		panic(err)
	} else {
		return bys
	}
}

func (s *SessionManager) clearSession(pname, sid string) error {
	if err := s.db.Delete([]byte(SessionLWTPrex+pname+sid), nil); err != nil && err != leveldb.ErrNotFound {
		return err
	}
	iter := s.db.NewIterator(&util.Range{Start: []byte(pname + sid)}, nil)
	defer iter.Release()
	keyPrex := []byte(pname + sid)
	icount := int64(0)
	for iter.Next() {
		keyBys := iter.Key()
		if bytes.Compare(keyBys[:len(keyPrex)], keyPrex) == 0 {
			if err := s.db.Delete(keyBys, nil); err != nil {
				return err
			}
			icount++
		} else {
			break
		}
	}
	log.TRACE.Printf("clear session:%s.%s,count:%v", pname, sid, icount)

	return iter.Error()

}
func (s *SessionManager) ClearTimeoutSession() error {
	iter := s.db.NewIterator(&util.Range{Start: []byte(SessionLWTPrex)}, nil)
	defer iter.Release()
	for iter.Next() {
		keyBys := iter.Key()
		if bytes.Compare(keyBys[:len(SessionLWTPrex)], []byte(SessionLWTPrex)) == 0 {
			key := string(keyBys)
			pname := key[len(SessionLWTPrex) : len(key)-24]
			sid := key[len(key)-24:]
			t := bytes2Time(iter.Value())
			if time.Now().Sub(t).Minutes() > float64(s.timeout) {
				if err := s.clearSession(pname, sid); err != nil {
					return err
				}
			}
		} else {
			break
		}
	}
	return iter.Error()
}

type Session struct {
	SessionID   string
	ProjectName string
}

func NewSession(id, pname string) *Session {
	return &Session{id, pname}
}
func (s *Session) Set(key string, value interface{}) (err error) {
	return SDB.Put(s.ProjectName, s.SessionID, key, value)
}
func (s *Session) Get(key string) interface{} {
	return SDB.Get(s.ProjectName, s.SessionID, key)
}
func (s *Session) All() (map[string]string, error) {
	return SDB.All()
}
func (s *Session) Abandon() error {
	return SDB.clearSession(s.ProjectName, s.SessionID)
}
func (s *Session) jsGet(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, s.Get(call.Argument(0).String()))
}
func (s *Session) jsSet(call otto.FunctionCall) otto.Value {
	key := oftenfun.AssertString(call.Argument(0))
	value := oftenfun.AssertValue(call.Argument(1))[0]
	return oftenfun.JSToValue(call.Otto, s.Set(key, value))
}
func (s *Session) jsAbandon(call otto.FunctionCall) otto.Value {
	return oftenfun.JSToValue(call.Otto, s.Abandon())
}
func (s *Session) Object() map[string]interface{} {
	return map[string]interface{}{
		"Set":     s.jsSet,
		"Get":     s.jsGet,
		"Abandon": s.jsAbandon,
	}
}
