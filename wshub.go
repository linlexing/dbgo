package main

import (
	"fmt"
)

type SocketMessage struct {
	Url     string
	Message string
}
type WSHub struct {
	// Registered connections.
	connections map[string]map[*WSConn]bool

	// Inbound messages from the connections.
	broadcast chan SocketMessage

	// Register requests from the connections.
	register chan *WSConn

	// Unregister requests from connections.
	unregister chan *WSConn
}

func NewWSHub() *WSHub {
	return &WSHub{
		connections: make(map[string]map[*WSConn]bool),
		broadcast:   make(chan SocketMessage),
		register:    make(chan *WSConn),
		unregister:  make(chan *WSConn),
	}
}
func (h *WSHub) removeConnection(c *WSConn) {
	if cmaps, ok := h.connections[c.RequestUrl.String()]; ok {
		delete(cmaps, c)
		close(c.send)
		if len(cmaps) == 0 {
			delete(h.connections, c.RequestUrl.String())
		}
	}
}
func (h *WSHub) send(c *WSConn, mes string) {
	fmt.Printf("send %s byte:%d\n", c.RequestUrl, len(mes))
	select {
	case c.send <- mes:
	default:
		h.removeConnection(c)
	}

}
func (h *WSHub) run() {
	for {
		select {
		case c := <-h.register:
			if cmaps, ok := h.connections[c.RequestUrl.String()]; !ok {
				h.connections[c.RequestUrl.String()] = map[*WSConn]bool{
					c: true,
				}
			} else {
				cmaps[c] = true
			}
		case c := <-h.unregister:
			h.removeConnection(c)
		case m := <-h.broadcast:
			for c := range h.connections[m.Url] {
				h.send(c, m.Message)
			}
		}
	}
}
