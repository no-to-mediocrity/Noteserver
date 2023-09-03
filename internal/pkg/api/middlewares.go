package api

import (
	"net/http"
	l "noteserver/internal/pkg/logger"

	"github.com/dgrijalva/jwt-go"
)

func AuthenticateMiddleware(next http.HandlerFunc, jwtSecret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			l.Logger.Error("Error:", err)
			http.Error(w, "Token is not valid", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}
