package middlewares

import (
	"VKR_gateway_service/internal/app"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT via external SSO HTTP endpoint.
// It sends GET SSO_HTTP_URL + "/api/auth/validate" with the same Authorization header.
// On non-200 it aborts request and returns JSON: {"error": "string"}.
func AuthMiddleware(a *app.App) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Avoid recursive validation if someone points SSO to this same service
        if strings.HasSuffix(c.Request.URL.Path, "/api/auth/validate") {
            c.Next()
            return
        }
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
			c.Abort()
			return
		}
		if a == nil || a.Config == nil || a.Config.SSO_HTTP_URL == "" {
			c.JSON(http.StatusBadGateway, gin.H{"error": "SSO url not configured"})
			c.Abort()
			return
		}

		target := strings.TrimRight(a.Config.SSO_HTTP_URL, "/") + "/api/auth/validate"
		a.Logger.Debug("Send request to ", target)
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, target, nil)
		if err != nil {
			a.Logger.Debug("Error", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		req.Header.Set("Authorization", tokenString)
		req.Header.Set("Accept", "application/json")
		timeout := 5 * time.Second
		if a.Config.GRPCTimeout > 0 {
			timeout = a.Config.GRPCTimeout
		}
		httpClient := &http.Client{Timeout: timeout}
		resp, err := httpClient.Do(req)
		if err != nil {
			a.Logger.Debug("Error", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
			msg := strings.TrimSpace(string(body))
			if msg == "" {
				msg = "invalid token"
			}
			c.JSON(resp.StatusCode, gin.H{"error": msg})
			c.Abort()
			return
		}
		c.Next()
	}
}
