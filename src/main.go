package main

import (
	"flag"
	"fmt"
)

func main() {
	var port int
	var maxFileSize int64
	var host, rootDir string
	var showIndex bool

	flag.StringVar(&host, "h", "", "Host to bind a server")
	flag.IntVar(&port, "p", 80, "Port to bind a server")
	flag.Int64Var(&maxFileSize, "m", 5<<10, "Maximum file size in Kb")
	flag.StringVar(&rootDir, "d", "/data", "Root file storage directory")
	flag.BoolVar(&showIndex, "i", true, "Allow to show directory index")

	f := "Starting server at %s:%v\n" +
		 "File storage directory: %s\n" +
		 "Maximum file size is set to: %vK\n"
	fmt.Printf(f, host, port, rootDir, maxFileSize)

	server := &Server{
		rootDir: rootDir,
		maxFileSize: maxFileSize << 13, // Kb Converted to bits
		showIndex: showIndex,
	}

	// Convert host and port to address string to bind server
	// If used with Docker, do not specify any host or set it to 0.0.0.0
	addr := fmt.Sprintf("%s:%v", host, port)

	server.Serve(addr)
}