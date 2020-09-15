package main

import (
	"github.com/valyala/fasthttp"
	"strings"
	"bytes"
	// "fmt"
	"os"
	"io/ioutil"
)

var (
	GET = []byte("GET")
	POST = []byte("POST")
	PUT = []byte("PUT")
	DELETE = []byte("DELETE")
)

type Server struct {
	rootDir string
	maxFileSize int64
}

func (s *Server) Serve(addr string) {
	h := &fasthttp.Server{
		Handler: s.handleRequest,
		// This is set to prevent preparse request denial
		// fasthttp.Server will reject all request exceeding this limit
		// Without any proper error description
		MaxRequestBodySize: int(s.maxFileSize * 2),
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

	if bytes.Equal(method, DELETE) {
		s.deleteFile(ctx)
		return
	}
	
	ctx.Error("Method Not Allowed", 405)
}

func (s *Server) serveFile(ctx *fasthttp.RequestCtx) {
	ctx.Write([]byte("This is get"))
}

// Receives and saves file from input.
// Name, extention and path resolved form request url.
// If path is provided along with file name (ex `/user/foo.txt`),
// it creates all necessary nested folder under the root dir.
func (s *Server) receiveFile(ctx *fasthttp.RequestCtx) {	
	header, err := ctx.FormFile("file")
	if err != nil {
		ctx.Error(err.Error(), 400)
		return
	}

	// Check if file size is less then specified limit.
	// This check will be done only if request body size limit
	// Is lover then defined in fasthttp.Server.MaxRequestBodySize
	if header.Size > s.maxFileSize {
		ctx.Error("File size exceeds the given limit", 400)
		return
	}

	// Prepare all directories if required
	// Name of the file separated by last `/` sign
	var b strings.Builder
	b.WriteString(s.rootDir)
	b.Write(ctx.Path())

	path := b.String()
	dirs := path[:strings.LastIndex(path, "/")]

	err = os.MkdirAll(dirs, os.ModePerm)
	if err != nil {
		ctx.Error(err.Error(), 500)
	}

	// Write the file
	err = fasthttp.SaveMultipartFile(header, path)
	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}

	ctx.SetBody([]byte("ok"))
}

// Deletes file determined by request path
// Does not allow to delete folders manually
func (s *Server) deleteFile(ctx *fasthttp.RequestCtx) {
	var b strings.Builder
	b.WriteString(s.rootDir)
	b.Write(ctx.Path())

	path := b.String()

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		ctx.Error("Not exists", 404)
		return
	}

	if info.IsDir() {
		ctx.Error("Target is a directory", 400)
		return
	}

	err = os.Remove(path)
	if err != nil {
		ctx.Error(err.Error(), 500)
		return
	}

	// Clean all empty directories recursive at background
	dirs := path[:strings.LastIndex(path, "/")]
	go cleanDirs(dirs)

	ctx.SetBody([]byte("ok"))
}

func cleanDirs(path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return
	}

	if len(files) > 0 {
		return
	}

	err = os.Remove(path)
	if err != nil {
		return
	}

	path = path[:strings.LastIndex(path, "/")]
	cleanDirs(path)
}