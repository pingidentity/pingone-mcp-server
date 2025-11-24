// Copyright © 2025 Ping Identity Corporation

package testutils

import (
	"time"

	"golang.org/x/oauth2"
)

type StaticTokenSource struct {
	token *oauth2.Token
}

func NewDefaultStaticTokenSource() *StaticTokenSource {
	return &StaticTokenSource{
		token: &oauth2.Token{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			TokenType:    "Bearer",
			Expiry:       time.Now().Add(time.Hour),
		},
	}
}

func NewStaticTokenSource(token *oauth2.Token) *StaticTokenSource {
	return &StaticTokenSource{
		token: token,
	}
}

func (s *StaticTokenSource) Token() (*oauth2.Token, error) {
	if s.token == nil {
		return &oauth2.Token{}, nil
	}
	return s.token, nil
}
