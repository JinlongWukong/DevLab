package api

import (
	"net/http"

	"github.com/JinlongWukong/DevLab/account"
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

//Only account with admin role allowed
func AdminRoleOnlyAllowed() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("account")
		if ac, exists := account.AccountDB.Get(header); exists {
			if ac.Role == account.RoleAdmin {
				c.Next()
			} else {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "only account with admin role allowed"})
			}
		} else {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "account not found"})
		}
	}
}
