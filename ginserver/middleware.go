package ginserver

import (
	"mime"
	"net/http"
	"strings"

	"slices"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func CORSMiddleware(config Config) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     config.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "DELETE", "PUT", "PATCH"},
		AllowHeaders:     config.AllowedHeaders,
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           config.CORSMaxAge,
	})
}

// CacheControlMiddleware sets the
// Deprecated: does nothing.
func CacheControlMiddleware(value string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Cache-Control", value)
		ctx.Next()
	}
}

func WithContentType(mimetypes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !HasContentType(c.Request, mimetypes) {
			c.AbortWithStatus(http.StatusUnsupportedMediaType)
			return
		}
	}
}

func HasContentType(r *http.Request, mimetypes []string) bool {
	contentType := r.Header.Get("Content-type")
	if contentType == "" {
		return slices.Contains(mimetypes, "application/octet-stream")
	}

	for _, v := range strings.Split(contentType, ",") {
		t, _, err := mime.ParseMediaType(v)
		if err != nil {
			break
		}
		if slices.Contains(mimetypes, t) {
			return true
		}
	}
	return false
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get(xRequestIDKey) == "" {
			xRequestID := xid.New().String()
			c.Request.Header.Set(xRequestIDKey, xRequestID)
			c.Set(xRequestIDKey, xRequestID)
		}
		s := trace.SpanFromContext(c.Request.Context())
		kv := attribute.String("request_id", c.Request.Header.Get(xRequestIDKey))
		s.SetAttributes(kv)

		c.Writer.Header().Set(xRequestIDKey, c.Request.Header.Get(xRequestIDKey))
		c.Next()
	}
}
