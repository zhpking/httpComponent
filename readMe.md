## 最简单demo

```go
// server.go
package main

import "net/http"

type Server interface {
	Route(pattern string, handleFunc http.HandlerFunc)
	Start(address string) error
}

type sdkHttpServer struct {
	Name string
}

func (s *sdkHttpServer) Route(pattern string, handleFunc http.HandlerFunc) {
	http.HandleFunc(pattern, handleFunc)
}

func (s *sdkHttpServer) Start(address string) error {
	return http.ListenAndServe(address, nil)
}

func NewHttpServer(name string) Server {
	return &sdkHttpServer{Name: name}
}
```

```go
// user.go
func func SignUp(w http.ResponseWriter, r *http.Request) {
}
```

```go
// main.go
package main

func main() {
	server := NewHttpServer("test-server")
	server.Route("/signUp", SignUp)
	server.Start(":8080")
}
```



## 封装context

```go
// server.go
package main

import "net/http"

type Server interface {
	Route(pattern string, handleFunc http.HandlerFunc)
	Start(address string) error
}

type sdkHttpServer struct {
	Name string
}

func (s *sdkHttpServer) Route(pattern string, handleFunc http.HandlerFunc) {
	http.HandleFunc(pattern, handleFunc)
}

func (s *sdkHttpServer) Start(address string) error {
	return http.ListenAndServe(address, nil)
}

func NewHttpServer(name string) Server {
	return &sdkHttpServer{Name: name}
}
```

```go
// context.go
package main

import (
	"encoding/json"
	"io"
	"net/http"
)

type Context struct {
	W http.ResponseWriter
	R *http.Request
}

func (c *Context) ReadJson(req interface{}) error {
	r := c.R
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Context) WriteJson(code int, resp interface{}) error {
	c.W.WriteHeader(code)
	respJson, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	_, err = c.W.Write(respJson)
	return err
}

func (c *Context) OkJson(resp interface{}) error {
	return c.WriteJson(http.StatusOK, resp)
}

func (c *Context) SystemERRORJson(resp interface{}) error {
	return c.WriteJson(http.StatusInternalServerError, resp)
}

func (c *Context) BadRequestJson(resp interface{}) error {
	return c.WriteJson(http.StatusBadRequest, resp)
}
```

```go
// user.go
package main

import (
	"fmt"
	"net/http"
)

type signUpReq struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	ConfirmedPassword string `json:"confirmed_password"`
}

type commonResponse struct {
	Code    int
	Message string
	Data    interface{}
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	req := &signUpReq{}
	ctx := &Context{W: w, R: r}
	resp := &commonResponse{Code: 0, Message: "suc"}
	err := ctx.ReadJson(req)
	if err != nil {
		ctx.BadRequestJson(resp)
		return
	}

	err = ctx.WriteJson(http.StatusOK, resp)
	if err != nil {
		fmt.Printf("写入响应失败:%v", err)
	}
}
```

```go
// main.go
package main

func main() {
	server := NewHttpServer("test-server")
	server.Route("/signUp", SignUp)
	server.Start(":8080")
}
```

## 不让用户自建创建Context，让httpComponent自己创建

```go
// server.go
package main

import "net/http"

type Server interface {
	Route(pattern string, handleFunc func(ctx *Context))
	Start(address string) error
}

type sdkHttpServer struct {
	Name string
}

func (s *sdkHttpServer) Route(pattern string, handleFunc func(ctx *Context)) {
	http.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		ctx := NewContext(writer, request)
		handleFunc(ctx)
	})
}

func (s *sdkHttpServer) Start(address string) error {
	return http.ListenAndServe(address, nil)
}

func NewHttpServer(name string) Server {
	return &sdkHttpServer{Name: name}
}
```

```go
// context.go
package main

import (
	"encoding/json"
	"io"
	"net/http"
)

type Context struct {
	W http.ResponseWriter
	R *http.Request
}

func NewContext(writer http.ResponseWriter, request *http.Request) *Context {
	return &Context{W: writer, R: request}
}

func (c *Context) ReadJson(req interface{}) error {
	r := c.R
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Context) WriteJson(code int, resp interface{}) error {
	c.W.WriteHeader(code)
	respJson, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	_, err = c.W.Write(respJson)
	return err
}

func (c *Context) OkJson(resp interface{}) error {
	return c.WriteJson(http.StatusOK, resp)
}

func (c *Context) SystemERRORJson(resp interface{}) error {
	return c.WriteJson(http.StatusInternalServerError, resp)
}

func (c *Context) BadRequestJson(resp interface{}) error {
	return c.WriteJson(http.StatusBadRequest, resp)
}
```

```go
// user.go
package main

import (
	"fmt"
	"net/http"
)

type signUpReq struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	ConfirmedPassword string `json:"confirmed_password"`
}

type commonResponse struct {
	Code    int
	Message string
	Data    interface{}
}

func SignUp(ctx *Context) {
	req := &signUpReq{}
	resp := &commonResponse{Code: 0, Message: "suc"}
	err := ctx.ReadJson(req)
	if err != nil {
		ctx.BadRequestJson(resp)
		return
	}

	err = ctx.WriteJson(http.StatusOK, resp)
	if err != nil {
		fmt.Printf("写入响应失败:%v", err)
	}
}
```

```go
// main.go
package main

func main() {
	server := NewHttpServer("test-server")
	server.Route("/signUp", SignUp)
	server.Start(":8080")
}
```

## 支持restful api（基于map的形式实现）

```go
// server.go
package main

import "net/http"

type Server interface {
	Route(method string, pattern string, handleFunc func(ctx *Context))
	Start(address string) error
}

type sdkHttpServer struct {
	Name    string
	handler *HandlerBaseOnMap
}

func (s *sdkHttpServer) Route(method string, pattern string, handleFunc func(ctx *Context)) {
	key := s.handler.key(method, pattern)
	s.handler.handlers[key] = handleFunc
	/*
		http.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
			ctx := NewContext(writer, request)
			handleFunc(ctx)
		})
	*/
}

func (s *sdkHttpServer) Start(address string) error {
	http.Handle("/", s.handler)
	return http.ListenAndServe(address, nil)
}

func NewHttpServer(name string) Server {
	return &sdkHttpServer{Name: name}
}
```

```go
// context.go
package main

import (
	"encoding/json"
	"io"
	"net/http"
)

type Context struct {
	W http.ResponseWriter
	R *http.Request
}

func NewContext(writer http.ResponseWriter, request *http.Request) *Context {
	return &Context{W: writer, R: request}
}

func (c *Context) ReadJson(req interface{}) error {
	r := c.R
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Context) WriteJson(code int, resp interface{}) error {
	c.W.WriteHeader(code)
	respJson, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	_, err = c.W.Write(respJson)
	return err
}

func (c *Context) OkJson(resp interface{}) error {
	return c.WriteJson(http.StatusOK, resp)
}

func (c *Context) SystemERRORJson(resp interface{}) error {
	return c.WriteJson(http.StatusInternalServerError, resp)
}

func (c *Context) BadRequestJson(resp interface{}) error {
	return c.WriteJson(http.StatusBadRequest, resp)
}
```

```go
//user.go
package main

import (
	"fmt"
	"net/http"
)

type signUpReq struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	ConfirmedPassword string `json:"confirmed_password"`
}

type commonResponse struct {
	Code    int
	Message string
	Data    interface{}
}

func SignUp(ctx *Context) {
	req := &signUpReq{}
	resp := &commonResponse{Code: 0, Message: "suc"}
	err := ctx.ReadJson(req)
	if err != nil {
		ctx.BadRequestJson(resp)
		return
	}

	err = ctx.WriteJson(http.StatusOK, resp)
	if err != nil {
		fmt.Printf("写入响应失败:%v", err)
	}
}
```

```go
// mapHandler.go
package main

import "net/http"

type HandlerBaseOnMap struct {
	// key为 method + uri
	handlers map[string]func(ctx *Context)
}

func (h *HandlerBaseOnMap) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	key := h.key(request.Method, request.URL.Path)
	if handler, ok := h.handlers[key]; ok {
		handler(NewContext(writer, request))
	} else {
		writer.WriteHeader(http.StatusNotFound)
		writer.Write([]byte("Not Found"))
	}
}

func (h *HandlerBaseOnMap) key(method string, path string) string {
	return method + "#" + path
}
```

