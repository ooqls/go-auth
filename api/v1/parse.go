package v1

import "github.com/gin-gonic/gin"

func ParseToken(c *gin.Context) string {
	token := c.Request.Header.Get("Authorization")
	if token == "" {
		return ""
	}

	if len(token) < 7 || token[:7] != "Bearer " {
		return ""
	}

	return token[7:]
}
