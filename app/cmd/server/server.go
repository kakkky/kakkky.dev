package main

import (
	"context"

	"github.com/kakkky/kakkky.dev/driver/httpserver"
)

type Server struct {
	httpSrv *httpserver.HTTPServer
}

func NewServer(hs *httpserver.HTTPServer) *Server {
	return &Server{httpSrv: hs}
}

func (s *Server) Run(ctx context.Context) error {
	return s.httpSrv.Run(ctx)
}
