package main

import (
	"bytes"
	"github.com/linlexing/dbgo/log"
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
func (s *SessionManager) Put(pname, sid, key, value string) error {
	if err := s.db.Put([]byte(SessionLWTPrex+pname+sid), time2Bytes(time.Now()), nil); err != nil {
		return err
	}
	return s.db.Put([]byte(pname+sid+key), []byte(value), nil)
}
func (s *SessionManager) Get(pname, sid, key string) string {
	if err := s.db.Put([]byte(SessionLWTPrex+pname+sid), time2Bytes(time.Now()), nil); err != nil {
		panic(err)
	}
	buf, err := s.db.Get([]byte(pname+sid+key), nil)
	if err != nil && err != leveldb.ErrNotFound {
		panic(err)
	} else {
		return string(buf)
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
func (s *Session) Set(key, value string) (err error) {
	return SDB.Put(s.ProjectName, s.SessionID, key, value)
}
func (s *Session) Get(key string) string {
	return SDB.Get(s.ProjectName, s.SessionID, key)
}
func (s *Session) All() (map[string]string, error) {
	return SDB.All()
}
func (s *Session) jsGet(call otto.FunctionCall) otto.Value {
	r, _ := otto.ToValue(s.Get(call.Argument(0).String()))
	return r
}
func (s *Session) jsSet(call otto.FunctionCall) otto.Value {
	err := s.Set(call.Argument(0).String(), call.Argument(1).String())
	if err != nil {
		v, _ := otto.ToValue(err.Error())
		return v
	}
	return otto.NullValue()
}
func (s *Session) Object() map[string]interface{} {
	return map[string]interface{}{
		"Set": s.jsSet,
		"Get": s.jsGet,
	}
}
