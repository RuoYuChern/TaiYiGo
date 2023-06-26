package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func loadCnSharesHistory(c *gin.Context) {
	c.String(http.StatusOK, "Hello")
}
