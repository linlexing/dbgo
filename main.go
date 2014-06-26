// dbgo project main.go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/linlexing/dbgo/httpgzip"
	"github.com/linlexing/dbgo/log"
	"github.com/robfig/cron"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

var (
	// Loggers
	MetaCache      *Cache
	DefaultProject string
	Filters        []Filter
	AppPath        string
	SDB            *SessionManager
	JSP            *JSPool
	Jobs           *cron.Cron
	NetFS          *WeedFS
)

func loadConfig() *Config {
	file, e := ioutil.ReadFile("./cnfg.json")
	if e != nil {
		log.ERROR.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	c := Config{}
	json.Unmarshal(file, &c)
	return &c
}
func main() {
	var err error
	if AppPath, err = filepath.Abs("."); err != nil {
		log.INFO.Fatal(err)
	}

	c := loadConfig()
	NetFS = NewWeedFS(c.WeedMaster)
	DefaultProject = c.DefaultProject
	MetaCache = NewCache(c.MetaDBUrl)
	SDB = NewSessionManager(path.Join(AppPath, "sdb"), c.SessionTimeout)
	defer SDB.Close()
	JSP = NewJSPool(10)
	Jobs := cron.New()
	Jobs.AddFunc("@every 3m", func() {
		log.TRACE.Print("start background job[ClearTimeoutSession]\n")
		if err := SDB.ClearTimeoutSession(); err != nil {
			panic(err)
		}
	})
	Jobs.Start()
	Filters = []Filter{PanicFilter, RouteFilter, SessionFilter, AuthFilter, ActionFilter}
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})
	http.Handle("/public/", httpgzip.NewHandler(http.StripPrefix("/public", http.FileServer(http.Dir(filepath.Join(AppPath, "public"))))))
	http.HandleFunc("/", handle)
	log.INFO.Printf("the http server listen on %v\nstatic file server on path:%v\n", c.Port, filepath.Join(AppPath, "public"))
	log.INFO.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%v", c.Port), "cert.pem", "key.pem", nil))

}
