package server

import (
	"log"
	"net"
	"net/http"

	"github.com/drone/mq/stomp"

	"golang.org/x/net/websocket"
)

// Server ...
type Server struct {
	router *router
}

// NewServer returns a new STOMP server.
func NewServer(options ...Option) *Server {
	server := &Server{
		router: newRouter(),
	}
	for _, option := range options {
		option(server)
	}
	return server
}

// Serve accepts incoming net.Conn requests.
func (s *Server) Serve(conn net.Conn) {
	log.Printf("stomp: successfully opened socket connection.")

	session := requestSession()
	session.peer = stomp.Conn(conn)

	defer func() {
		if r := recover(); r != nil {
			log.Printf("stomp: unexpected panic: %s", r)
		}
		log.Printf("stomp: successfully closed socket connection.")

		s.router.disconnect(session)
		session.peer.Close()
		session.release()
	}()

	if err := s.router.serve(session); err != nil {
		log.Printf("stomp: server error. %s", err)
	}
}

// ServeHTTP accepts incoming http.Request, upgrades to a websocket and
// begins sending and receiving STOMP messages.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("stomp: handle websocket request.")
	websocket.Handler(func(conn *websocket.Conn) {
		s.Serve(conn)
	}).ServeHTTP(w, r)
}

// Client returns a stomp.Client that has a direct peer connection
// to the server.
func (s *Server) Client() *stomp.Client {
	a, b := stomp.Pipe()

	go func() {
		session := requestSession()
		session.peer = b
		if err := s.router.serve(session); err != nil {
			log.Printf("stomp: server error. %s", err)
		}
	}()
	return stomp.New(a)
}
