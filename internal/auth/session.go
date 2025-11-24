// Copyright © 2025 Ping Identity Corporation

package auth

import (
	"time"

	"golang.org/x/oauth2"
)

type AuthSession struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	Expiry       time.Time `json:"expiry"`
	SessionId    string    `json:"sessionId"`
}

func NewAuthSession(token oauth2.Token, sessionId string) AuthSession {
	return AuthSession{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
		SessionId:    sessionId,
	}
}
