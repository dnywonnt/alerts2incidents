package v1 // dnywonnt.me/alerts2incidents/internal/api/v1

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

// ConnectToLDAPServer establishes a connection to the LDAP server with the given URL and credentials.
// Returns the LDAP connection object or an error if the connection fails.
func ConnectToLDAPServer(ldapURL, bindDN, bindPassword string, tlsCfg *tls.Config) (*ldap.Conn, error) {
	l, err := ldap.DialURL(ldapURL, ldap.DialWithTLSConfig(tlsCfg))
	if err != nil {
		return nil, fmt.Errorf("error dialing LDAP server: %w", err)
	}

	defer func() {
		if err != nil {
			l.Close()
		}
	}()

	if err = l.Bind(bindDN, bindPassword); err != nil {
		return nil, fmt.Errorf("error binding: %w", err)
	}

	return l, nil
}

// SearchLDAPUser searches for a user in the LDAP directory using the provided login.
// Returns the LDAP entry for the user or an error if the search fails.
func SearchLDAPUser(conn *ldap.Conn, baseDN, login string) (*ldap.Entry, error) {
	searchRequest := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(sAMAccountName=%s)", login),
		[]string{"*"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("error searching LDAP entries: %w", err)
	}

	if len(sr.Entries) == 0 {
		return nil, fmt.Errorf("user not found: %s", login)
	}

	return sr.Entries[0], nil
}

// IsLDAPUserInAllowedGroup checks if the LDAP user belongs to any of the allowed groups.
// Returns true if the user is in one of the allowed groups, false otherwise.
func IsLDAPUserInAllowedGroup(ldapUser *ldap.Entry, allowedGroups []string) bool {
	memberOf := ldapUser.GetAttributeValues("memberOf")

	for _, cn := range memberOf {
		if strings.HasPrefix(cn, "CN=") {
			endIndex := strings.Index(cn, ",")
			if endIndex != -1 {
				userGroup := strings.ToLower(cn[3:endIndex])
				for _, allowedGroup := range allowedGroups {
					if userGroup == strings.ToLower(allowedGroup) {
						return true
					}
				}
			}
		}
	}

	return false
}

// GetLDAPUserName retrieves the common name (cn) attribute of the LDAP user.
// Returns the common name or an error if the attribute is not found.
func GetLDAPUserName(ldapUser *ldap.Entry) (string, error) {
	cn := ldapUser.GetAttributeValue("cn")
	if cn == "" {
		return "", fmt.Errorf("cn attribute not found for user")
	}

	return cn, nil
}

// ExtractLDAPDomain extracts the domain component (dc) from the base DN string.
// Returns the domain component or an error if the dc= prefix is not found.
func ExtractLDAPDomain(baseDn string) (string, error) {
	prefix := "dc="
	startIndex := strings.Index(baseDn, prefix)
	if startIndex == -1 {
		return "", fmt.Errorf("dc= not found in string")
	}

	endIndex := strings.Index(baseDn[startIndex:], ",")
	if endIndex == -1 {
		endIndex = len(baseDn)
	} else {
		endIndex += startIndex
	}

	domainComponent := baseDn[startIndex+len(prefix) : endIndex]

	return domainComponent, nil
}
