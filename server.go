package main

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	Error_NoResult = errors.New("the result is nil")
)

func handle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("request:", r.URL.Path)
	c := NewAgent(w, r)
	c.jsRuntime = JSP.Get()
	defer JSP.Release(c.jsRuntime)
	Filters[0](c, Filters[1:])
	if c.Result == nil {
		c.RenderError(Error_NoResult)
	}

	c.Result.Apply(r, w)
}
