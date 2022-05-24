package webserver

import (
	"asearch/logger"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}
		if errorMessage == "" {
			logger.Infof("[GIN] %3d | %13v | %15s |%-7s %#v",
				statusCode, latency, clientIP, method, path)
		} else {
			logger.Infof("[GIN] %3d | %13v | %15s |%-7s %#v\n%s",
				statusCode, latency, clientIP, method, path, errorMessage)
		}
	}
}
