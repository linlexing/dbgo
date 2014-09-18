package main

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/linlexing/dbgo/log"
	"net/http"
	"strings"
)

var (
	Error_NoResult = errors.New("the result is nil")
)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func createWSConn(w http.ResponseWriter, r *http.Request) *WSConn {
	paths := strings.Split(r.URL.Path, "/")
	if len(paths) < 2 {
		panic(fmt.Errorf("the path must have project/channel"))
	}
	ck, err := r.Cookie("sid")
	if err != nil {
		panic(err)
	}
	if ck.Value == "" {
		panic("session id can't is empty")
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	return &WSConn{
		send:       make(chan string, 256),
		ws:         ws,
		RequestUrl: r.URL,
		SessionID:  ck.Value,
	}

}
func handle(w http.ResponseWriter, r *http.Request) {
	var c *ControllerAgent
	defer func() {
		if err := recover(); err != nil {
			if c.ws != nil {
				log.ERROR.Println("[websocket]recover :", err)
				c.ws.conn.send <- fmt.Sprint(err)
			} else {
				log.ERROR.Println("recover :", err)
				http.Error(w, fmt.Sprint(err), 200)
			}
		}
	}()
	upgrade := r.Header.Get("Upgrade")
	if upgrade == "websocket" || upgrade == "Websocket" {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}
		conn := createWSConn(w, r)
		SocketHub.register <- conn
		go conn.writePump()
		go conn.readPump()
		c = NewAgentWS(
			&WSAgent{conn, "open", ""},
		)
	} else {
		c = NewAgent(w, r)
	}
	c.jsRuntime = JSP.Get()
	defer JSP.Release(c.jsRuntime)
	Filters[0](c, Filters[1:])
	if c.ws == nil {
		if c.Result == nil {
			c.RenderError(Error_NoResult)
		}
		c.Result.Apply(r, w)
	} else {
		switch tv := c.Result.(type) {
		case *ErrorResult:
			c.ws.conn.send <- tv.Error.Error()
		}
	}
}
