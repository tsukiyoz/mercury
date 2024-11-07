package web

import (
	"github.com/gin-gonic/gin"
)

type Page struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type handler interface {
	RegisterRoutes(server *gin.Engine)
}
