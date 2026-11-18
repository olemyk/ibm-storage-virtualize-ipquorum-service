package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// responseWriter wraps gin.ResponseWriter to capture response size
type responseWriter struct {
	gin.ResponseWriter
	size int
}

func (w *responseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}

func (w *responseWriter) WriteString(s string) (int, error) {
	size, err := w.ResponseWriter.WriteString(s)
	w.size += size
	return size, err
}

// Middleware returns a Gin middleware that records API metrics
func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Get request size
		requestSize := c.Request.ContentLength

		// Wrap response writer to capture size
		rw := &responseWriter{
			ResponseWriter: c.Writer,
			size:           0,
		}
		c.Writer = rw

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start)
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		endpoint := c.FullPath()

		// If endpoint is empty (404), use the request path
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		m.RecordAPIRequest(method, endpoint, status, duration, requestSize, int64(rw.size))

		// Record errors if status >= 400
		if c.Writer.Status() >= 400 {
			errorType := "client_error"
			if c.Writer.Status() >= 500 {
				errorType = "server_error"
			}
			m.RecordAPIError(method, endpoint, errorType)
		}
	}
}

// Made with Bob
