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
	// 把路由注册进s.handler.handlers
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
