package gocash

import (
	"fmt"
	"net"
)

// NewServer creates server on `ip` and `port`, returning handling *Server object
func NewServer(ip string, port int) *Server {
	svr := Server{Sessions: make(chan *net.Conn, 1)}
	go svr.Listen(ip, port)
	return &svr
}

// Server handles listening, accepting new connections and passing off cmds of connected clients
//
type Server struct {
	// Input channel should be used to read new incoming lines, or cmds that need to be processed.
	Sessions chan *net.Conn
	ln       *net.Listener
}

// Listen opens the given `ip` and `port` handling new connections
func (s *Server) Listen(ip string, port int) {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("[ERROR]", err)
		}
		s.Sessions <- &conn
	}
}
