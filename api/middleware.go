package api

import (
	"net/http"

	"github.com/JinlongWukong/DevLab/auth"
	"github.com/gin-gonic/gin"
)

func AuthorizeToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		const BEARER_SCHEMA = "Bearer "
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) <= len(BEARER_SCHEMA) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is wrong"})
			return
		}
		tokenString := authHeader[len(BEARER_SCHEMA):]
		err, name := auth.ValidateToken(tokenString)
		if err == nil {
			c.Request.Header.Add("account", name)
			c.Next()
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}
