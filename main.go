// dbgo project main.go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/daaku/go.httpgzip"
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
	SocketHub      *WSHub
	CacheHub       *AppCacheHub
	Meta           MetaProject
	DefaultProject string
	Filters        []Filter
	AppPath        string
	SDB            *SessionManager
	JSP            *JSPool
	Jobs           *cron.Cron
)

func loadConfig() *Config {
	file, e := ioutil.ReadFile("./cnfg.json")
	if e != nil {
		log.ERROR.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	c := Config{}
	if err := json.Unmarshal(file, &c); err != nil {
		panic(err)
	}
	return &c
}

func main() {
	var err error
	SocketHub = NewWSHub()
	CacheHub = NewAppCacheHub()
	if AppPath, err = filepath.Abs("."); err != nil {
		log.INFO.Fatal(err)
	}
	c := loadConfig()
	DefaultProject = c.DefaultProject
	Meta = NewMetaProject(c.MetaDBUrl)
	SDB = NewSessionManager(path.Join(AppPath, "sdb"), c.SessionTimeout)
	defer SDB.Close()
	JSP = NewJSPool(10)
	Jobs := cron.New()
	Jobs.AddFunc("@every 3m", func() {
		//log.TRACE.Print("start background job[ClearTimeoutSession]\n")
		if err := SDB.ClearTimeoutSession(); err != nil {
			panic(err)
		}
	})
	Jobs.Start()
	go SocketHub.run()
	Filters = []Filter{PanicFilter, RouteFilter, ParseJsonFilter, SessionFilter, UserFilter, BuildObjectFilter, InterceptFilter, LoadControlFilter, UrlAuthFilter, ActionFilter}
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})
	http.Handle("/public/", httpgzip.NewHandler(http.StripPrefix("/public", http.FileServer(http.Dir(filepath.Join(AppPath, "public"))))))
	http.HandleFunc("/", handle)
	log.INFO.Printf("the http server listen on %v\nstatic file server on path:%v\n", c.Port, filepath.Join(AppPath, "public"))
	log.INFO.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%v", c.Port), "cert.pem", "key.pem", nil))

}
