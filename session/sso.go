package session

import (
	"fmt"
	"strings"
)

// Cookie holds a browser cookie with its domain.
type Cookie struct {
	Name   string
	Value  string
	Domain string
}

type JWToken struct {
	Name  string
	Value string
}

// SatSession contains the access token and cookies obtained after login.
type SatSession struct {
	AccessToken string
	jwts        []JWToken
	cookies     []Cookie
}

// TokenHeader returns the cookies formatted as a Cookie HTTP header value.
func (s *SatSession) Cookies() string {
	parts := make([]string, 0, len(s.cookies))
	for _, c := range s.cookies {
		parts = append(parts, fmt.Sprintf("%s=%s", c.Name, c.Value))
	}
	return strings.Join(parts, ";")
}

func (s *SatSession) GetToken(name string) string {
	for _, c := range s.jwts {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}

type SatSSO interface {
	FetchSession(forceWindow bool) (*SatSession, error)
}
