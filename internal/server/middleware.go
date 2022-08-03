package server

import (
	"compress/gzip"
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

func DecompressGZIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(`Content-Encoding`) == `gzip` {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			r.Body = gz
			defer gz.Close()
		}
		next.ServeHTTP(w, r)
	})
}

func CheckJWT(s Server) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get(authorizationHeader)
			if len(authHeader) == 0 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			authParts := strings.Split(authHeader, "Bearer ")
			if len(authParts) != 2 {
				http.Error(w, "Malformed JWT", http.StatusUnauthorized)
				return
			}
			jwtToken := authParts[1]

			token, err := jwt.ParseWithClaims(jwtToken, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(s.config.JWTKey), nil
			})
			if err != nil || token == nil {
				http.Error(w, "Malformed JWT", http.StatusUnauthorized)
				return
			}

			if claims, ok := token.Claims.(*AuthClaims); ok && token.Valid {
				ctx := context.WithValue(r.Context(), contextJWTLogin, claims.Login)

				// Access login in handlers like this
				// login, _ := r.Context().Value("login").(string)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				http.Error(w, "Malformed JWT", http.StatusUnauthorized)
			}
		})
	}
}
