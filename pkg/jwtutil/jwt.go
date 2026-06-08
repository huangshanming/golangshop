package jwtutil

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	TokenTypeApp   TokenType = "app"
	TokenTypeAdmin TokenType = "admin"
)

type Claims struct {
	UserID   uint64    `json:"user_id"`
	TenantID uint64    `json:"tenant_id"`
	Username string    `json:"username,omitempty"`
	Type     TokenType `json:"type"`
	jwt.RegisteredClaims
}

type Manager struct {
	appSecret   []byte
	adminSecret []byte
	appExpire   time.Duration
	adminExpire time.Duration
}

func NewManager(appSecret, adminSecret string, appHours, adminHours int) *Manager {
	return &Manager{
		appSecret:   []byte(appSecret),
		adminSecret: []byte(adminSecret),
		appExpire:   time.Duration(appHours) * time.Hour,
		adminExpire: time.Duration(adminHours) * time.Hour,
	}
}

func (m *Manager) GenerateAppToken(userID, tenantID uint64) (string, error) {
	return m.sign(TokenTypeApp, userID, tenantID, "", m.appSecret, m.appExpire)
}

func (m *Manager) GenerateAdminToken(userID, tenantID uint64, username string) (string, error) {
	return m.sign(TokenTypeAdmin, userID, tenantID, username, m.adminSecret, m.adminExpire)
}

func (m *Manager) sign(tokenType TokenType, userID, tenantID uint64, username string, secret []byte, expire time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		TenantID: tenantID,
		Username: username,
		Type:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expire)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "goshop",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (m *Manager) ParseAppToken(tokenStr string) (*Claims, error) {
	return m.parse(tokenStr, m.appSecret, TokenTypeApp)
}

func (m *Manager) ParseAdminToken(tokenStr string) (*Claims, error) {
	return m.parse(tokenStr, m.adminSecret, TokenTypeAdmin)
}

func (m *Manager) parse(tokenStr string, secret []byte, expected TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.Type != expected {
		return nil, errors.New("token type mismatch")
	}
	return claims, nil
}
