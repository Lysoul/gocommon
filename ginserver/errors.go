package ginserver

import (
	"errors"

	"github.com/Lysoul/gocommon/shared"
	"github.com/gin-gonic/gin"
)

// a simple handler for common errors, use as a fallback error handler.
func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		err := c.Errors.Last()
		if err == nil {
			return
		}

		status, data, headers := errToHTTP(err.Err)
		for k, v := range headers {
			c.Header(k, v)
		}

		c.JSON(status, data)
	}
}

func errToHTTP(err error) (int, any, map[string]string) {
	switch {
	case errors.Is(err, shared.ErrUnauthorized):
		return 401, nil, nil
	case errors.Is(err, shared.ErrPermissionDenied):
		return 403, nil, nil
	case errors.Is(err, shared.ErrNotFound):
		return 404, nil, nil
	case errors.Is(err, shared.ErrValidationFailed):
		return 400, err, nil
	default:
		return 500, nil, nil
	}
}
