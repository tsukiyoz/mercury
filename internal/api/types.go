package api

import "github.com/gin-gonic/gin"

type handler interface {
	RegisterRoutes(server *gin.Engine)
}
