package main

import (
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/robertkrimen/otto"
)

type WSAgent struct {
	conn    *WSConn
	event   string
	message string
}

func (g *WSAgent) Object() map[string]interface{} {
	return map[string]interface{}{
		"Event":   g.event,
		"Message": g.message,
		"Send": func(call otto.FunctionCall) otto.Value {
			str := oftenfun.AssertString(call.Argument(0))
			SocketHub.send(g.conn, str)
			return otto.UndefinedValue()
		},
		"Broadcast": func(call otto.FunctionCall) otto.Value {
			mes := SocketMessage{
				g.conn.RequestUrl.String(),
				oftenfun.AssertString(call.Argument(1)),
			}
			SocketHub.broadcast <- mes
			return otto.UndefinedValue()
		},
	}
}
