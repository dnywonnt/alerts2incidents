package v1 // dnywonnt.me/alerts2incidents/internal/api/v1

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// GenerateJWTToken generates a JWT token with an expiration interval.
func GenerateJWTToken(secretKey []byte, expInterval time.Duration) (string, error) {
	// Create a new token object, specifying signing method and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(expInterval).Unix(), // Expiration time
	})

	// Sign the token with the secret key
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWTToken validates a JWT token string using a secret key.
// It returns an error if the token is invalid or if the signing method is not HMAC.
func ValidateJWTToken(tokenString string, secretKey []byte) error {
	// Parse the JWT token string.
	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check if the signing method is HMAC.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			// Return an error if the signing method is unexpected.
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the secret key to validate the token.
		return secretKey, nil
	})

	// Return an error if the token is invalid.
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// Return nil if the token is valid.
	return nil
}
