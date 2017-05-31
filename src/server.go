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
	"./conf"
	"./db"
	"./formats"
	"./puzzle"
	"./service"
	"./ws"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gocraft/web"
	"golang.org/x/net/websocket"
	"io/ioutil"
	"net/http"
)

type User struct {
	name string
}

type Context struct {
	user User
	db   *sql.DB
}

type PuzzleUploadResponse struct {
	Id string
}

func (c *Context) Headers(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	// allow testing with localhost client
	rw.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	next(rw, req)
}

func (c *Context) PuzzleUpload(rw web.ResponseWriter, req *web.Request) {

	req.ParseMultipartForm(32 << 20)
	file, _, err := req.FormFile("file")
	if err != nil {
		panic("no file found")
	}
	b, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	var p puzzle.Puzzle
	formats := [...]formats.Format{formats.NewPuz(), formats.NewXPF()}
	for _, f := range formats {
		p, err = f.Parse(b)
		if err == nil {
			break
		}
	}
	if err != nil {
		panic("could not parse file")
	}

	session := db.NewSession(c.db)
	id, err := session.PuzzleCreate(&p)
	if err != nil {
		panic(err)
	}

	resp := PuzzleUploadResponse{Id: id}
	b, err = json.Marshal(resp)
	if err != nil {
		panic("cannot generate json response")
	}
	fmt.Fprint(rw, string(b))
}

func (c *Context) PuzzleGet(rw web.ResponseWriter, req *web.Request) {
	id := req.PathParams["id"]
	session := db.NewSession(c.db)

	p, err := session.PuzzleGetById(id)
	if err != nil {
		panic(err)
	}
	b, err := formats.NewXPF().Format(p)
	if err != nil {
		panic(err)
	}
	fmt.Fprint(rw, string(b))
}

func (c *Context) SolutionStart(rw web.ResponseWriter, req *web.Request) {
	var request struct {
		PuzzleId string
	}
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&request)
	if err != nil {
		panic(err)
	}

	session := db.NewSession(c.db)
	p, err := session.PuzzleGetById(request.PuzzleId)
	if err != nil {
		panic(err)
	}

	entries := make([]db.Entry, p.Width*p.Height)
	for i := 0; i < p.Width*p.Height; i++ {
		entries[i].Ordinal = i
		entries[i].Value = " "
	}
	s := db.Solution{
		PuzzleId: p.Id,
		Version:  1,
		Entries:  entries}

	id, err := session.SolutionCreate(&s)
	if err != nil {
		panic(err)
	}

	var response struct {
		Id       string
		PuzzleId string
		Version  int
		Grid     string
	}

	response.Id = id
	response.PuzzleId = p.Id
	response.Version = s.Version
	response.Grid = s.GridString()

	b, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	fmt.Fprint(rw, string(b))
}

func (c *Context) SolutionGet(rw web.ResponseWriter, req *web.Request) {
	id := req.PathParams["id"]

	var response struct {
		PuzzleId string
		Version  int
		Grid     string
	}
	session := db.NewSession(c.db)

	s, err := session.SolutionGetById(id)
	if err != nil {
		panic(err)
	}
	response.PuzzleId = s.PuzzleId
	response.Version = s.Version
	response.Grid = s.GridString()

	b, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	fmt.Fprint(rw, string(b))
}

func (c *Context) PuzzleUploadGet(rw web.ResponseWriter, req *web.Request) {
	body := "<html><body><form method=\"post\" enctype=\"multipart/form-data\" action=\"/puzzle/\"><input type=\"file\" name=\"file\"/><input type=\"submit\"/></form></body></html>"
	fmt.Fprint(rw, string(body))
}

func main() {
	config, err := conf.LoadConfig("config")
	if err != nil {
		panic(err)
	}
	dbhandle, err := sql.Open("mysql", config.DBUri)
	defer dbhandle.Close()
	if err != nil {
		panic(err)
	}
	session := db.NewSession(dbhandle)
	wsServer := ws.NewServer(service.SolutionServiceNew(session))
	go wsServer.Listen()

	router := web.New(Context{}).
		Middleware(web.LoggerMiddleware).
		Middleware(func(c *Context, rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
			c.db = dbhandle
			next(rw, req)
		}).
		Middleware((*Context).Auth).
		Middleware((*Context).Headers).
		Get("/puzzle/:id", (*Context).PuzzleGet).
		Get("/puzzle/upload", (*Context).PuzzleUploadGet).
		Get("/solution/:id", (*Context).SolutionGet).
		Get("/ws", func(rw web.ResponseWriter, req *web.Request) {
			websocket.Handler(wsServer.OnConnected).ServeHTTP(rw, req.Request)
		}).
		Post("/solution", (*Context).SolutionStart).
		Post("/puzzle/", (*Context).PuzzleUpload)
	http.ListenAndServe("localhost:4000", router)
}
