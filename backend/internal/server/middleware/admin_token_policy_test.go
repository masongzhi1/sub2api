//go:build unit

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRequireAdminJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("rejects_admin_api_key_auth", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("auth_method", adminAuthMethodAPIKey)
			c.Next()
		})
		router.POST("/admin/token-management", RequireAdminJWT(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/admin/token-management", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusForbidden, w.Code)
		require.Contains(t, w.Body.String(), "ADMIN_JWT_REQUIRED")
	})

	t.Run("allows_jwt_auth", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("auth_method", adminAuthMethodJWT)
			c.Next()
		})
		router.POST("/admin/token-management", RequireAdminJWT(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/admin/token-management", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)
	})
}
