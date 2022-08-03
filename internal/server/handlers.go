package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/theplant/luhn"
)

type UserAuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthClaims struct {
	Login string `json:"login"`
	jwt.StandardClaims
}

func getAuthJWT(s Server, login string) (token string, err error) {
	claims := AuthClaims{
		login,
		jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Unix() + int64(time.Hour/time.Second),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(s.config.JWTKey))
}

func getAuthHeader(s Server, login string) (header string, err error) {
	token, err := getAuthJWT(s, login)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Bearer %s", token), nil
}

func badRequestHandler(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "Bad request", http.StatusBadRequest)
}

func registerHandler(s Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Bad content type", http.StatusBadRequest)
			return
		}

		var request UserAuthRequest
		if err = json.Unmarshal(body, &request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(request.Login) == 0 {
			http.Error(w, "Login is required", http.StatusBadRequest)
			return
		}

		if len(request.Password) == 0 {
			http.Error(w, "Password is required", http.StatusBadRequest)
			return
		}

		available, err := s.repository.IsLoginAvailable(r.Context(), request.Login)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !available {
			http.Error(w, "Provided login is not available", http.StatusConflict)
			return
		}

		err = s.repository.Register(r.Context(), request.Login, request.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		header, err := getAuthHeader(s, request.Login)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Authentication", header)
		w.WriteHeader(http.StatusOK)
	}
}

func loginHandler(s Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Bad content type", http.StatusBadRequest)
			return
		}

		var request UserAuthRequest
		if err = json.Unmarshal(body, &request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(request.Login) == 0 {
			http.Error(w, "Login is required", http.StatusBadRequest)
			return
		}

		if len(request.Password) == 0 {
			http.Error(w, "Password is required", http.StatusBadRequest)
			return
		}

		success, err := s.repository.Login(r.Context(), request.Login, request.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !success {
			http.Error(w, "Wrong login or password", http.StatusUnauthorized)
			return
		}

		header, err := getAuthHeader(s, request.Login)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Authentication", header)
		w.WriteHeader(http.StatusOK)
	}
}

func uploadHandler(s Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if r.Header.Get("Content-Type") != "text/plain" {
			http.Error(w, "Bad content type", http.StatusBadRequest)
			return
		}

		order, err := strconv.Atoi(string(body))
		if err != nil {
			http.Error(w, "Bad order number", http.StatusUnprocessableEntity)
			return
		}

		if !luhn.Valid(order) {
			http.Error(w, "Bad order number", http.StatusUnprocessableEntity)
			return
		}

		owner, err := s.repository.OrderOwner(r.Context(), order)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jwtOwner, _ := r.Context().Value("login").(string)

		if owner == jwtOwner {
			w.WriteHeader(http.StatusOK)
			return
		} else if len(owner) > 0 && owner != jwtOwner {
			http.Error(w, "Order was uploaded by another user", http.StatusConflict)
			return
		}

		err = s.repository.UploadOrder(r.Context(), jwtOwner, order)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
