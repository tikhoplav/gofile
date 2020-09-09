package main

import (
	"github.com/valyala/fasthttp"
	"bytes"
	"fmt"
)

var (
	GET = []byte("GET")
	POST = []byte("POST")
	PUT = []byte("PUT")
	DELETE = []byte("DELETE")
)

type Server struct {
	rootDir string
	maxFileSize int
}

func (s *Server) Serve(addr string) {
	h := &fasthttp.Server{
		Handler: s.handleRequest,
		MaxRequestBodySize: 40 << 20,
	}
	h.ListenAndServe(addr)
}

func (s *Server) handleRequest(ctx *fasthttp.RequestCtx) {	
	var method = ctx.Method()

	if bytes.Equal(method, GET) {
		s.serveFile(ctx)
		return
	}

	if bytes.Equal(method, POST) || bytes.Equal(method, PUT) {
		s.receiveFile(ctx)
		return
	}
	
	ctx.Error("Method Not Allowed", 405)
}

func (s *Server) serveFile(ctx *fasthttp.RequestCtx) {
	ctx.Write([]byte("This is get"))
}

func (s *Server) receiveFile(ctx *fasthttp.RequestCtx) {
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.Error(err.Error(), 0)
		return
	}
	for _, v := range form.File {
		for _, header := range v {
			err = fasthttp.SaveMultipartFile(header, fmt.Sprintf("/tmp/%s", header.Filename))
			if err != nil {
				ctx.Error(err.Error(), 0)
				return
			}
			ctx.WriteString("Success")
		}
	}
}