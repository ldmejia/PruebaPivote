package middlewares

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const issuer = "pivote-api"

type contextKey string

const ctxKeySubject contextKey = "auth.subject"

func IssueToken(secret []byte, userID string, ttl time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

func NewRequireAuth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || tokenString == "" {
				http.Error(w, `{"code":"unauthorized","message":"falta el token"}`, http.StatusUnauthorized)
				return
			}
			subject, err := parseToken(secret, tokenString)
			if err != nil {
				http.Error(w, `{"code":"unauthorized","message":"token invalido o vencido"}`, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeySubject, subject)))
		})
	}
}

func UserID(ctx context.Context) string {
	id, _ := ctx.Value(ctxKeySubject).(string)
	return id
}

func parseToken(secret []byte, tokenString string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		return secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(issuer), jwt.WithExpirationRequired())
	if err != nil {
		return "", err
	}
	if claims.Subject == "" {
		return "", errors.New("token sin subject")
	}
	return claims.Subject, nil
}
