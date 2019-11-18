package ginlogrus

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	Key = "ginlogrus"
)

// New returns a gin compatable middleware using logrus to defaultValue
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
			"http_request_method":         c.Request.Method,
			"http_request_path":           path,
			"http_ip":                     c.ClientIP(),
			"http_request_user-agent":     c.Request.UserAgent(),
			"http_request_host":           c.Request.Host,
			"http_request_content-length": c.Request.ContentLength,
		})

		SetLogger(c, logger)

		c.Next()

		entry := GetLogger(c)
		if entry != nil {
			logger = entry
		}

		latency := time.Now().Sub(start)
		statusCode := c.Writer.Status()

		entry = logger.WithFields(logrus.Fields{
			"http_request_status":         statusCode,
			"http_request_latency":        latency,
			"http_request_latency_string": latency.String(),
		})

		if statusCode > 499 {
			entry.Error(defaultValue(c.Errors.String(), statusCode))
		} else if statusCode > 399 {
			entry.Warn(defaultValue(c.Errors.String(), statusCode))
		} else {
			if _, ok := skip[path]; ok {
				return
			}
			entry.Info(defaultValue(c.Errors.String(), statusCode))
		}
	}
}

// defaultValue checks if a string is empty and returns the corresponding http status text
func defaultValue(s string, code int) string {
	if s == "" {
		return http.StatusText(code)
	}
	return s
}

// GetLogger takes a gin context and returns a logrus Entry logger if it exists
// on the gin context. If it does not exist it returns nil
func GetLogger(c *gin.Context) *logrus.Entry {
	logger, exists := c.Get(Key)
	if !exists {
		return nil
	}
	if l, ok := logger.(*logrus.Entry); ok {
		return l
	}
	return nil
}

// SetLogger sets a logger on the current gin context
func SetLogger(c *gin.Context, logger *logrus.Entry) {
	c.Set(Key, logger)
}
