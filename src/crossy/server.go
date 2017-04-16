package main

// version 1:
//  no auth
//  fetching /puzzle/id gets puz-format fill
//  fetching /solve/id gets current solution
//  anyone can post /solve/
//
//  bobcopeland.com/crosswords/xyzzy
//   --> embeds xwordjs with id=xyzzy
//   --> xwordjs fetches /solve/xyzzy
//   --> if first solver, start the clock
//    --> respond with current fill

import (
	"github.com/gocraft/web"
	"fmt"
	"net/http"
	"encoding/json"
)

type User struct {
	name string
}

type Context struct {
	user User
}

type Puzzle struct {

}

type PuzzleUploadResponse struct {
    Id string
}

type SolveGetResponse struct {
    Id string
    Version int
    Fill string
}

func (c *Context) Auth(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	c.user.name = "bobc"
	next(rw, req)
}

func (c *Context) PuzzleUpload(rw web.ResponseWriter, req *web.Request) {

    req.ParseMultipartForm(32 << 20);
    _, header, err := req.FormFile("file");
    fmt.Printf("puz: %s\n", header.Filename);
    if err != nil {
        panic("no file found")
    }

    // todo generate an id
    // todo save file somewhere, keyed by id

    resp := PuzzleUploadResponse{Id: "xyzzy"}
    b, err := json.Marshal(resp)
    if err != nil {
        panic("cannot generate json response")
    }
    fmt.Fprint(rw, string(b))
}

func (c *Context) PuzzleUploadGet(rw web.ResponseWriter, req *web.Request) {
    body := "<html><body><form method=\"post\" enctype=\"multipart/form-data\" action=\"/puzzle/\"><input type=\"file\" name=\"file\"/><input type=\"submit\"/></form></body></html>"
    fmt.Fprint(rw, string(body))
}

func main() {
	router := web.New(Context{}).
		Middleware(web.LoggerMiddleware).
		Middleware(web.ShowErrorsMiddleware).
		Middleware((*Context).Auth).
		//Get("/puzzle/:id:\\d.*", (*Context).PuzzleGet).
		Get("/puzzle/upload", (*Context).PuzzleUploadGet).
		Post("/puzzle/", (*Context).PuzzleUpload)
	http.ListenAndServe("localhost:4000", router)
}

