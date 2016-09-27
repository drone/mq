package server

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/drone/mq/logger"
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
	logger.Verbosef("stomp: successfully opened socket connection.")

	session := requestSession()
	session.peer = stomp.Conn(conn)

	defer func() {
		if r := recover(); r != nil {
			logger.Warningf("stomp: unexpected panic: %s", r)
		}
		logger.Verbosef("stomp: successfully closed socket connection.")

		s.router.disconnect(session)
		session.peer.Close()
		session.release()
	}()

	if err := s.router.serve(session); err != nil {
		logger.Warningf("stomp: server error. %s", err)
	}
}

// ServeHTTP accepts incoming http.Request, upgrades to a websocket and
// begins sending and receiving STOMP messages.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Verbosef("stomp: handle websocket request.")
	websocket.Handler(func(conn *websocket.Conn) {
		s.Serve(conn)
	}).ServeHTTP(w, r)
}

// HandleSessions writes a JSON-encoded list of sessions to the http.Request.
func (s *Server) HandleSessions(w http.ResponseWriter, r *http.Request) {
	type sessionResp struct {
		Addr    string            `json:"address"`
		User    string            `json:"username"`
		Headers map[string]string `json:"headers"`
	}

	var sessions []sessionResp
	s.router.RLock()
	for sess := range s.router.sessions {
		headers := map[string]string{}
		for i := 0; i < sess.msg.Header.Len(); i++ {
			k, v := sess.msg.Header.Index(i)
			headers[string(k)] = string(v)
		}
		sessions = append(sessions, sessionResp{
			Addr:    sess.peer.Addr(),
			User:    string(sess.msg.User),
			Headers: headers,
		})
	}
	s.router.RUnlock()

	json.NewEncoder(w).Encode(sessions)
}

// HandleDests writes a JSON-encoded list of destinations to the http.Request.
func (s *Server) HandleDests(w http.ResponseWriter, r *http.Request) {
	type destionatResp struct {
		Dest string `json:"destination"`
	}

	var dests []destionatResp
	s.router.RLock()
	for dest := range s.router.destinations {
		d := destionatResp{
			Dest: dest,
		}
		dests = append(dests, d)
	}
	s.router.RUnlock()

	json.NewEncoder(w).Encode(dests)
}

// Client returns a stomp.Client that has a direct peer connection
// to the server.
func (s *Server) Client() *stomp.Client {
	a, b := stomp.Pipe()

	go func() {
		session := requestSession()
		session.peer = b
		if err := s.router.serve(session); err != nil {
			logger.Warningf("stomp: server error. %s", err)
		}
	}()
	return stomp.New(a)
}
