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
	Projects       *DBGo
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
func writeConfig(c *Config) error {
	file, e := os.Create("./cnfg.json")
	defer file.Close()
	if e != nil {
		return e
	}
	if bys, err := json.Marshal(c); err != nil {
		return err
	} else if _, err := file.Write(bys); err != nil {
		return err
	}
	return nil
}
func initiDBGo(c *Config) error {
	Projects = NewDBGo(c.DBUrl)
	if c.MetaDBUrl == "" {
		metaUrl, err := Projects.CreateProject("meta", "meta123")
		if err != nil {
			return err
		}
		c.MetaDBUrl = metaUrl
		if err := writeConfig(c); err != nil {
			return err
		}
	}
	Projects.ReadyMetaProject(c.MetaDBUrl)
	return nil
}
func main() {
	var err error
	if AppPath, err = filepath.Abs("."); err != nil {
		log.INFO.Fatal(err)
	}
	c := loadConfig()
	DefaultProject = c.DefaultProject
	if err = initiDBGo(c); err != nil {
		log.INFO.Fatal(err)
	}
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
	Filters = []Filter{PanicFilter, RouteFilter, ParseJsonFilter, SessionFilter, BuildObjectFilter, InterceptFilter, LoadControlFilter, UrlAuthFilter, UserFilter, ActionFilter}
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})
	http.Handle("/public/", httpgzip.NewHandler(http.StripPrefix("/public", http.FileServer(http.Dir(filepath.Join(AppPath, "public"))))))
	http.HandleFunc("/", handle)
	log.INFO.Printf("the http server listen on %v\nstatic file server on path:%v\n", c.Port, filepath.Join(AppPath, "public"))
	log.INFO.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%v", c.Port), "cert.pem", "key.pem", nil))

}
