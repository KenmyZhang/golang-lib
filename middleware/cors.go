package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/mattermost/mattermost-server/utils"
	"net/http"
	"strings"
)

func Cors(allowCorsFrom string, allowedMethods []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if allowCorsFrom != "" {
			if utils.CheckOrigin(c.Request, allowCorsFrom) {
				c.Writer.Header().Set("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))

				if c.Request.Method == "OPTIONS" {
					c.Writer.Header().Set(
						"Access-Control-Allow-Methods",
						strings.Join(allowedMethods, ", "))

					c.Writer.Header().Set(
						"Access-Control-Allow-Headers",
						c.Request.Header.Get("Access-Control-Request-Headers"))
				}
			}
		}

		if c.Request.Method == "OPTIONS" {
			c.JSON(http.StatusOK, "Options Request!")
		}
		c.Next()
	}
}

func CheckOrigin(r *http.Request, allowedOrigins string) bool {
	origin := r.Header.Get("Origin")
	if allowedOrigins == "*" {
		return true
	}
	for _, allowed := range strings.Split(allowedOrigins, " ") {
		if allowed == origin {
			return true
		}
	}
	return false
}
