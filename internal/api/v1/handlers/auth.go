package handlers // dnywonnt.me/alerts2incidents/internal/api/v1/handlers

import (
	"crypto/tls"
	"fmt"
	"net/http"

	v1 "dnywonnt.me/alerts2incidents/internal/api/v1"
	"dnywonnt.me/alerts2incidents/internal/config"
	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

// RegisterAuthRoutes registers the authentication routes with the router.
func RegisterAuthRoutes(router *gin.Engine, apiCfg *config.ApiConfig) {
	routerGroup := router.Group("/api/v1/auth")

	// Register the LDAP authentication route
	routerGroup.POST("/ldap", authenticateLDAPUser(apiCfg))
}

// authenticateLDAPUser handles LDAP user authentication.
func authenticateLDAPUser(apiCfg *config.ApiConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var authRequest struct {
			Login    string `json:"login" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		// Bind JSON request to authRequest struct
		if err := c.ShouldBindJSON(&authRequest); err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Failed to unmarshal request JSON data")
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		// Construct LDAP URL
		ldapUrl := fmt.Sprintf("ldap://%s:%d", apiCfg.LDAP.Host, apiCfg.LDAP.Port)
		domain, err := v1.ExtractLDAPDomain(apiCfg.LDAP.BaseDN)
		if err != nil {
			log.WithFields(log.Fields{
				"baseDN": apiCfg.LDAP.BaseDN,
				"error":  err.Error(),
			}).Error("Failed to extract domain from baseDN string")
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		// Construct bindDN
		bindDN := fmt.Sprintf("%s\\%s", domain, authRequest.Login)

		// Connect to LDAP server
		ldapConn, err := v1.ConnectToLDAPServer(ldapUrl, bindDN, authRequest.Password, &tls.Config{InsecureSkipVerify: true})
		if err != nil {
			log.WithFields(log.Fields{
				"ldapUrl": ldapUrl,
				"bindDN":  bindDN,
				"error":   err.Error(),
			}).Error("Failed to connect to LDAP server")
			c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			return
		}
		defer ldapConn.Close()

		// Search for the LDAP user
		ldapUser, err := v1.SearchLDAPUser(ldapConn, apiCfg.LDAP.BaseDN, authRequest.Login)
		if err != nil {
			log.WithFields(log.Fields{
				"baseDN": apiCfg.LDAP.BaseDN,
				"login":  authRequest.Login,
				"error":  err.Error(),
			}).Error("Failed to search LDAP user")
			c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			return
		}

		// Check if the LDAP user is in allowed groups
		if !v1.IsLDAPUserInAllowedGroup(ldapUser, apiCfg.LDAP.AllowedGroups) {
			log.WithFields(log.Fields{
				"login":         authRequest.Login,
				"allowedGroups": apiCfg.LDAP.AllowedGroups,
			}).Error("LDAP user is not in allowed group")
			c.JSON(http.StatusForbidden, gin.H{"message": "permission denied"})
			return
		}

		// Generate JWT token
		token, err := v1.GenerateJWTToken([]byte(apiCfg.JwtSecretKey), apiCfg.JwtTokenExpirationInterval)
		if err != nil {
			log.WithFields(log.Fields{
				"login": authRequest.Login,
				"error": err.Error(),
			}).Error("Failed to generate JWT token")
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		// Get the LDAP user name
		userName, err := v1.GetLDAPUserName(ldapUser)
		if err != nil {
			log.WithFields(log.Fields{
				"login": authRequest.Login,
				"error": err.Error(),
			}).Error("Failed to retrieve LDAP user name")
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		// Respond with the user name and token
		c.JSON(http.StatusOK, gin.H{
			"user_name":       userName,
			"token":           token,
			"token_life_time": apiCfg.JwtTokenExpirationInterval.Nanoseconds(),
		})
	}
}
