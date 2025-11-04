package ginserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lysoul/gocommon/ginserver"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCacheControlMiddleware(t *testing.T) {
	const (
		headerName   = "Cache-Control"
		defaultValue = "no-store"
	)

	router := gin.Default()
	router.Use(ginserver.CacheControlMiddleware(defaultValue))

	t.Run("Falls back to default when not set by handler", func(t *testing.T) {
		router.GET("/without-set", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{
				"message": "OK",
			})
		})

		req, err := http.NewRequest(http.MethodGet, "/without-set", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		result := recorder.Result()

		expectedHeaderValue := defaultValue
		require.Equal(t, expectedHeaderValue, result.Header.Get(headerName))
	})

	t.Run("Respects handler's set value", func(t *testing.T) {
		expectedHeaderValue := "public, max-age=3600"

		router.GET("/with-set", func(ctx *gin.Context) {
			ctx.Header(headerName, expectedHeaderValue)
			ctx.JSON(http.StatusOK, gin.H{
				"message": "OK",
			})
		})

		req, err := http.NewRequest(http.MethodGet, "/with-set", nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		result := recorder.Result()

		require.Equal(t, expectedHeaderValue, result.Header.Get(headerName))
	})
}
