package ws

import (
	"../service"
	"golang.org/x/net/websocket"
	"log"
)

type Server struct {
	service    *service.SolutionService
	clients    map[*Client]bool
	recv       chan *service.SolutionMutation
	register   chan *Client
	unregister chan *Client
}

func NewServer(solService *service.SolutionService) *Server {
	return &Server{
		solService,
		make(map[*Client]bool),
		make(chan *service.SolutionMutation),
		make(chan *Client),
		make(chan *Client),
	}
}

func (s *Server) OnConnected(conn *websocket.Conn) {
	log.Println("New client connected")
	HandleClient(s, conn)
}

func (s *Server) Listen() {

	log.Println("Starting WS server thread")
	for {
		select {
		case c := <-s.register:
			s.clients[c] = true
			log.Println("Added client: ", c, " count: ", len(s.clients))

		case c := <-s.unregister:
			log.Println("Unregistered client: ", c, " count: ", len(s.clients))
			if _, ok := s.clients[c]; ok {
				delete(s.clients, c)
				close(c.send)
			}

		case m := <-s.recv:
			log.Println("< client")
			out, err := s.service.Update(m)
			if err == nil {
				for client, _ := range s.clients {
					client.send <- out
				}
			}
		}
	}
}
