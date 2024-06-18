package v1 // dnywonnt.me/alerts2incidents/internal/api/v1

import (
	"net/http"
	"strings"
	"time"

	"dnywonnt.me/alerts2incidents/internal/config"
	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

// LoggerMiddleware returns a Gin middleware function that logs information about incoming requests.
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now() // Record the start time of the request processing.

		c.Next() // Continue processing subsequent middleware and the request handler.

		endTime := time.Now()             // Record the end time of the request processing.
		latency := endTime.Sub(startTime) // Calculate the duration of the request processing.

		// Log information about the received request.
		log.WithFields(log.Fields{
			"method":  c.Request.Method,   // HTTP method of the request.
			"path":    c.Request.URL.Path, // Request path.
			"status":  c.Writer.Status(),  // HTTP status code of the response.
			"ip":      c.ClientIP(),       // Client IP address.
			"latency": latency.String(),   // Duration of the request processing as a string.
		}).Info("Received a new request")
	}
}

// JWTMiddleware checks for the presence and validity of a JWT token in requests
func JWTMiddleware(apiCfg *config.ApiConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Error("Authorization header is missing")
			c.JSON(http.StatusUnauthorized, gin.H{"message": "authorization header is missing"})
			c.Abort()
			return
		}

		// Validate the token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.WithFields(log.Fields{
				"authHeader": authHeader,
			}).Error("Invalid token format")
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate the token
		if err := ValidateJWTToken(tokenString, []byte(apiCfg.JwtSecretKey)); err != nil {
			log.WithFields(log.Fields{
				"token": tokenString,
				"error": err,
			}).Error("Failed to validate token")
			c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			c.Abort()
			return
		}

		c.Next()
	}
}
