package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"prodata/logs"
)

type ApiFunc func(ctx *Context)

type Context struct {
	Request      *http.Request
	Writer       http.ResponseWriter
	Json         func(v any) error
	ReadJson     func(v any) error
	Logger       logs.Logger
	PureRoute    string
	Error        func(error any, code int)
	WriteHeader  func(code int)
	IP           string
	IfErrNotNull func(err error) bool
}

type Routes struct {
	RealRoute    string
	DynamicRoute func() string
}

func (ctx *Context) NewRoutes() *Routes {
	return &Routes{
		RealRoute: ctx.Request.URL.Path,
		DynamicRoute: func() string {
			r := ctx.PureRoute
			if len(r) == len(ctx.Request.URL.Path) {
				return ""
			}

			return ctx.Request.URL.Path[len(r):]
		},
	}
}

func NewContext(w http.ResponseWriter, r *http.Request, route string) *Context {
	return &Context{
		Request: r,
		Writer:  w,
		Json: func(v any) error {
			w.Header().Set("Content-Type", "application/json")
			return json.NewEncoder(w).Encode(v)
		},
		ReadJson: func(v any) error {
			return json.NewDecoder(r.Body).Decode(v)
		},
		Logger:    *logs.NewSistemLogger(),
		PureRoute: route,
		Error: func(error any, code int) {
			http.Error(w, fmt.Sprintf("%s", error), code)
		},
		WriteHeader: func(code int) {
			w.WriteHeader(code)
		},
		IP: r.RemoteAddr,
		IfErrNotNull: func(err error) bool {
			if err != nil {
				logs.NewLogger().LogAndSendSystemMessage(err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return false
			}

			return true
		},
	}
}

func (ctx *Context) Return() {
	return
}

func Post(route string, handler func(ctx *Context)) {
	http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Wrong Method", http.StatusMethodNotAllowed)
			return
		}

		ctx := NewContext(w, r, route)
		handler(ctx)
	})
}

func Get(route string, handler func(ctx *Context)) {
	http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", "GET")

		if r.Method != http.MethodGet {
			http.Error(w, "Wrong Method", http.StatusMethodNotAllowed)
			return
		}

		ctx := NewContext(w, r, route)
		handler(ctx)
	})
}
