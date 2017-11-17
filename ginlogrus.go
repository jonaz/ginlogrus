package ginlogrus

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
)

// New returns a gin compatable middleware using logrus to log
func New(logger *logrus.Logger, timeFormat string, utc bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path
		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		if utc {
			end = end.UTC()
		}

		statusCode := c.Writer.Status()

		entry := logger.WithFields(logrus.Fields{
			"status":     statusCode,
			"method":     c.Request.Method,
			"path":       path,
			"ip":         c.ClientIP(),
			"latency":    latency,
			"user-agent": c.Request.UserAgent(),
			"time":       end.Format(timeFormat),
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
			return
		}

		if statusCode > 499 {
			entry.Error()
		} else if statusCode > 399 {
			entry.Warn()
		} else {
			entry.Info()
		}

	}
}
