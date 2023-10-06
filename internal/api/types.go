package api

import "github.com/gin-gonic/gin"

type handler interface {
	RegisterRoutes(engine *gin.Engine)
}
