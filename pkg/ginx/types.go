package ginx

import "github.com/gin-gonic/gin"

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

type Server struct {
	*gin.Engine
	Addr string
}

func (s *Server) Start() error {
	return s.Engine.Run(s.Addr)
}
