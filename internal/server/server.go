package server

import (
	"fmt"
	"io"
	"net"

	"github.com/Dharmik7601/http-protocol-golang/internal/request"
	"github.com/Dharmik7601/http-protocol-golang/internal/response"
)

type Server struct {
	closed  bool
	handler Handler
}

func newServer(handler Handler) *Server {
	return &Server{
		closed:  false,
		handler: handler,
	}
}

type HandleError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w *response.Writer, req *request.Request)

func runConnection(s *Server, conn io.ReadWriteCloser) {
	defer conn.Close()

	responseWriter := response.NewWriter(conn)
	headers := response.GetDefaultHeader(0)
	r, err := request.RequestFromReader(conn)
	if err != nil {
		responseWriter.WriteStatusLine(response.StatusBadRequest)
		responseWriter.WriteHeaders(headers)
		return
	}

	// writer := bytes.NewBuffer([]byte{})
	s.handler(responseWriter, r)

}

func runServer(s *Server, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		if s.closed {
			return
		}

		go runConnection(s, conn)
	}
}

func Serve(port uint16, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := newServer(handler)
	go runServer(server, listener)

	return server, nil
}

func (s *Server) Close() error {
	s.closed = true
	return nil
}
