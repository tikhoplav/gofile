package main

import (
	"github.com/valyala/fasthttp"
	"strings"
	"bytes"
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
	showIndex bool
	fileHandler fasthttp.RequestHandler
}

func (s *Server) Serve(addr string) {
	h := &fasthttp.Server{
		Handler: s.handleRequest,
		// This is set to prevent preparse request denial.
		// fasthttp.Server will reject all request exceeding this limit
		// Without any proper error description.
		MaxRequestBodySize: int(s.maxFileSize * 2),
	}

	fs := &fasthttp.FS{
		Root: s.rootDir,
		GenerateIndexPages: s.showIndex,
		Compress: true,
	}

	s.fileHandler = fs.NewRequestHandler()

	h.ListenAndServe(addr)
}

func (s *Server) handleRequest(ctx *fasthttp.RequestCtx) {	
	var method = ctx.Method()
	
	if bytes.Equal(method, GET) {
		s.fileHandler(ctx)
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

// Returns full path, directory and file name from request path.
// Uses server's rootDir as the base for path string.
// File name is determined by last slash character.
func (s *Server) parsePath(ctx *fasthttp.RequestCtx) (string, string, string) {
	var b strings.Builder

	b.WriteString(s.rootDir)
	b.Write(ctx.Path())

	path := b.String()
	fileName := path[strings.LastIndex(path, "/"):]
	dir := path[:strings.LastIndex(path, "/")]
	
	return path, fileName, dir
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

	path, _, dir := s.parsePath(ctx)

	err = os.MkdirAll(dir, os.ModePerm)
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

// Deletes file determined by request path.
// Does not allow to delete folders,
// But clears all empty folders in the file path.
func (s *Server) deleteFile(ctx *fasthttp.RequestCtx) {
	path, _, dir := s.parsePath(ctx)

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

	// Clean all empty directories recursively at background.
	go cleanDirs(dir)

	ctx.SetBody([]byte("ok"))
}

// Recursively removes all empty folders.
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