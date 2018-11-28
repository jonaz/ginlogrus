package ginlogrus

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// New returns a gin compatable middleware using logrus to log
// skipPaths only skips the INFO loglevel
func New(logger *logrus.Logger, skipPaths ...string) gin.HandlerFunc {
	var skip map[string]struct{}

	if length := len(skipPaths); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range skipPaths {
			skip[path] = struct{}{}
		}
	}
	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path
		logger := logger.WithFields(logrus.Fields{
			"method":     c.Request.Method,
			"path":       path,
			"ip":         c.ClientIP(),
			"user-agent": c.Request.UserAgent(),
			"host":       c.Request.Host,
			"length":     c.Request.ContentLength,
		})

		c.Set("ginlogrus", logger)

		c.Next()

		latency := time.Now().Sub(start)
		statusCode := c.Writer.Status()

		entry := logger.WithFields(logrus.Fields{
			"status":         statusCode,
			"latency":        latency,
			"latency_string": latency.String(),
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.String())
			return
		}

		if statusCode > 499 {
			entry.Error()
		} else if statusCode > 399 {
			entry.Warn()
		} else {
			if _, ok := skip[path]; ok {
				return
			}
			entry.Info()
		}
	}
}

// GetLogger takes a gin context and returns a logrus Entry logger if it exists
// on the gin context. If it does not exist it returns nil
func GetLogger(c *gin.Context) *logrus.Entry {
	logger, exists := c.Get("ginlogrus")
	if !exists {
		return nil
	}
	if l, ok := logger.(*logrus.Entry); ok {
		return l
	}
	return nil
}
