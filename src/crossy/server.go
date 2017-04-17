package main

// version 1:
//  no auth
//  fetching /puzzle/id gets puz-format fill - DONE
//  server parses puz format into something more reasonable
//  fetching /solve/id gets current solution
//    fetch anything works DONE
//    solution saved in backend store
//  anyone can post /solve/
//  database as backend store
//
//  bobcopeland.com/crosswords/xyzzy
//   --> embeds xwordjs with id=xyzzy (send in uri for now) DONE
//   --> xwordjs periodically fetches /solve/xyzzy DONE
//   --> xwordjs periodically posts /solve/xyzzy for updates
//   --> if first solver, start the clock
//    --> respond with current fill

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/gocraft/web"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
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
	Id      string
	Version int
	Fill    []string
}

func (c *Context) Auth(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	c.user.name = "bobc"
	next(rw, req)
}

func (c *Context) Headers(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	rw.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	next(rw, req)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randString(n int) string {
	rbytes := make([]byte, n)
	_, err := rand.Read(rbytes)
	if err != nil {
		panic("rand")
	}

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[int(rbytes[i])%len(letters)]
	}
	return string(b)
}

func SaveFile(file multipart.File) (id string) {

	id = randString(8)

	defer file.Close()
	f, err := os.OpenFile("./store/"+id+".puz", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic("could not save file")
	}

	defer f.Close()
	io.Copy(f, file)
	return id
}

func LoadFile(id string) (file *os.File) {
	// TODO check id for sanity
	f, err := os.OpenFile("./store/"+id+".puz", os.O_RDONLY, 0)
	if err != nil {
		panic("could not load file")
	}
	return f
}

func (c *Context) SolveGet(rw web.ResponseWriter, req *web.Request) {
	id := req.PathParams["id"]
	r := SolveGetResponse{}
	r.Id = id
	r.Version = 2
	r.Fill = make([]string, 21)
	for i := range r.Fill {
		r.Fill[i] = strings.Repeat(" ", 21)
	}
	r.Fill[0] = "abcde#xxxx#yyyy#zzzzz"
	b, err := json.Marshal(r)
	if err != nil {
		panic("cannot generate json response")
	}
	fmt.Fprint(rw, string(b))
}

func (c *Context) PuzzleUpload(rw web.ResponseWriter, req *web.Request) {

	req.ParseMultipartForm(32 << 20)
	file, header, err := req.FormFile("file")
	fmt.Printf("puz: %s\n", header.Filename)
	if err != nil {
		panic("no file found")
	}

	id := SaveFile(file)
	resp := PuzzleUploadResponse{Id: id}
	b, err := json.Marshal(resp)
	if err != nil {
		panic("cannot generate json response")
	}
	fmt.Fprint(rw, string(b))
}

func (c *Context) PuzzleGet(rw web.ResponseWriter, req *web.Request) {
	id := req.PathParams["id"]
	file := LoadFile(id)
	defer file.Close()
	_, err := io.Copy(rw, file)
	if err != nil {
		panic("cannot load file")
	}
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
		Middleware((*Context).Headers).
		Get("/puzzle/:id", (*Context).PuzzleGet).
		Get("/puzzle/upload", (*Context).PuzzleUploadGet).
		Get("/solve/:id", (*Context).SolveGet).
		Post("/puzzle/", (*Context).PuzzleUpload)
	http.ListenAndServe("localhost:4000", router)
}
