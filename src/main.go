package main

import (
	"github.com/valyala/fasthttp"
	"fmt"
)

func main() {
	fasthttp.ListenAndServe(":80", handleHttpRequest)
}

func handleHttpRequest(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Hello, World!")
}