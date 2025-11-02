package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	PSP_API_KEY_HEADER = "X-PSP-API-Key"
	PSP_API_KEY        = "psp_secret_key_12345" // TODO: Move to secret
)

func VerifyPSPApiKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader(PSP_API_KEY_HEADER)

		if apiKey != PSP_API_KEY {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Unauthorized - Invalid API key",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
