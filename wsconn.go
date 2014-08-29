package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 5 * 1024 * 1024
)

// connection is an middleman between the websocket connection and the hub.
type WSConn struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send       chan string
	RequestUrl *url.URL
	SessionID  string
}

// readPump pumps messages from the websocket connection to the hub.
func (c *WSConn) readPump() {
	defer func() {
		SocketHub.unregister <- c
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		if err = c.callAction(string(message)); err != nil {
			log.Println(err)
			break
		}
	}
}

// write writes a message with the given message type and payload.
func (c *WSConn) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *WSConn) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, []byte(message)); err != nil {
				return
			}
		case <-ticker.C:
			log.Print("time out ,send ping\n")
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
func (c *WSConn) callAction(mes string) (errResult error) {
	defer func() {
		if err := recover(); err != nil {
			switch tv := err.(type) {
			case error:
				errResult = tv
			default:
				errResult = fmt.Errorf("%v", tv)
			}
		}
	}()
	cagent := NewAgentWS(
		&WSAgent{c, "message", mes},
	)
	cagent.jsRuntime = JSP.Get()
	defer JSP.Release(cagent.jsRuntime)
	Filters[0](cagent, Filters[1:])
	switch tv := cagent.Result.(type) {
	case *ErrorResult:
		panic(tv.Error)
	}
	return
}
