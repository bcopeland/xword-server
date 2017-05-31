package ws

import (
	"../service"
	"golang.org/x/net/websocket"
	"io"
	"log"
)

type Client struct {
	server *Server
	conn   *websocket.Conn
	send   chan *service.SolutionMutation
	id     string
}

func HandleClient(server *Server, conn *websocket.Conn) {

	log.Println("receiving...")
	msg := service.SolutionMutation{}
	log.Println("hi...")
	err := websocket.JSON.Receive(conn, &msg)
	log.Println("done...")
	if err == io.EOF {
		log.Println("no msg, aborting client...")
		return
	} else if err != nil {
		log.Println("err: ", err)
	}
	log.Println("msg: ", msg)

	client := &Client{
		server,
		conn,
		make(chan *service.SolutionMutation),
		msg.Id}

	go client.writerThread()

	log.Println("registering...")
	client.server.register <- client
	log.Println("recv msg...")
	client.server.recv <- &msg
	log.Println("start reader thread...")
	client.readerThread()
}

func (c *Client) readerThread() {
	defer func() {
		c.server.unregister <- c
		c.conn.Close()
	}()
	for {
		log.Println("getting msg")
		msg := service.SolutionMutation{}
		err := websocket.JSON.Receive(c.conn, &msg)
		log.Println("msg rcvd: ", msg, err)
		if err == io.EOF {
			return
		} else if err == nil {
			c.server.recv <- &msg
		}
	}
}

func (c *Client) writerThread() {
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				break
			}
			log.Println("Got update ", c.id, " ", msg)
			if c.id == msg.Id {
				log.Println("> ", msg)
				websocket.JSON.Send(c.conn, msg)
			}
		}
	}
}
