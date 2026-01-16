package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger logs request information
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		requestID, _ := c.Get("request_id")

		c.Next()

		latency := time.Since(startTime)
		statusCode := c.Writer.Status()

		log.Printf("[%s] %s %s %d %v",
			requestID,
			method,
			path,
			statusCode,
			latency,
		)
	}
}
