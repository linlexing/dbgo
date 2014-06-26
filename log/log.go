// log project log.go
package log

import (
	"io/ioutil"
	"log"
	"os"
)

var (
	//TRACE = log.New(ioutil.Discard, "TRACE ", log.Ldate|log.Ltime|log.Lshortfile)
	TRACE    = log.New(os.Stdout, "TRACE ", log.Ldate|log.Ltime|log.Lshortfile)
	SQLTRACE = log.New(ioutil.Discard, "SQL TRACE ", log.Ldate|log.Ltime|log.Lshortfile)
	INFO     = log.New(os.Stdout, "INFO  ", log.Ldate|log.Ltime|log.Lshortfile)
	WARN     = log.New(os.Stdout, "WARN  ", log.Ldate|log.Ltime|log.Lshortfile)
	ERROR    = log.New(os.Stderr, "ERROR ", log.Ldate|log.Ltime|log.Lshortfile)
)
