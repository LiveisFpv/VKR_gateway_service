package middlewares

import (
	"VKR_gateway_service/internal/app"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
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
		userID, ok := extractUserIDFromHeader(resp.Header)
		if !ok {
			userID, ok = extractUserIDFromToken(tokenString)
		}
		if ok && userID > 0 {
			c.Set("user_id", userID)
		}
		c.Next()
	}
}

func extractUserID(body []byte) (int64, bool) {
	if len(body) == 0 {
		return 0, false
	}
	var payload interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, false
	}
	return findUserIDRecursive(payload)
}

func findUserIDRecursive(payload interface{}) (int64, bool) {
	switch v := payload.(type) {
	case map[string]interface{}:
		keys := []string{"user_id", "userId", "userID", "User_id", "UserId", "id", "uid", "sub"}
		for _, key := range keys {
			if raw, ok := v[key]; ok {
				if id, ok := normalizeID(raw); ok {
					return id, true
				}
			}
		}
		for _, raw := range v {
			if id, ok := findUserIDRecursive(raw); ok {
				return id, true
			}
		}
	case []interface{}:
		for _, raw := range v {
			if id, ok := findUserIDRecursive(raw); ok {
				return id, true
			}
		}
	}
	return 0, false
}

func extractUserIDFromHeader(header http.Header) (int64, bool) {
	keys := []string{"X-User-Id", "X-UserId", "X-UserID"}
	for _, key := range keys {
		value := strings.TrimSpace(header.Get(key))
		if value == "" {
			continue
		}
		id, err := strconv.ParseInt(value, 10, 64)
		if err == nil && id > 0 {
			return id, true
		}
	}
	return 0, false
}

func extractUserIDFromToken(tokenString string) (int64, bool) {
	tokenString = strings.TrimSpace(tokenString)
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = strings.TrimSpace(strings.TrimPrefix(tokenString, "Bearer "))
	}
	parts := strings.Split(tokenString, ".")
	if len(parts) < 2 {
		return 0, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, false
	}
	return extractUserID(payload)
}

func normalizeID(v interface{}) (int64, bool) {
	switch t := v.(type) {
	case float64:
		return int64(t), true
	case string:
		id, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return 0, false
		}
		return id, true
	case json.Number:
		id, err := t.Int64()
		if err != nil {
			return 0, false
		}
		return id, true
	default:
		return 0, false
	}
}
