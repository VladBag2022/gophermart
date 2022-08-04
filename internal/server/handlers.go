package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	log "github.com/sirupsen/logrus"
	"github.com/theplant/luhn"
)

type UserAuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
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

		if r.Header.Get("Content-Type") != contentTypeJSON {
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

		w.Header().Set(authorizationHeader, header)
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

		if r.Header.Get("Content-Type") != contentTypeJSON {
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

		w.Header().Set(authorizationHeader, header)
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

		order, err := strconv.ParseInt(string(body), 10, 64)
		if err != nil {
			http.Error(w, "Bad order number", http.StatusUnprocessableEntity)
			return
		}

		if !luhn.Valid(int(order)) {
			http.Error(w, "Bad order number", http.StatusUnprocessableEntity)
			return
		}

		owner, err := s.repository.OrderOwner(r.Context(), order)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jwtOwner, _ := r.Context().Value(contextJWTLogin).(string)

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

func listHandler(s Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jwtOwner, _ := r.Context().Value(contextJWTLogin).(string)

		orders, err := s.repository.Orders(r.Context(), jwtOwner)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(orders) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		response, err := json.Marshal(&orders)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", contentTypeJSON)
		w.WriteHeader(http.StatusOK)

		_, err = w.Write(response)
		if err != nil {
			log.Trace("Log in prod")
		}
	}
}

func balanceHandler(s Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jwtOwner, _ := r.Context().Value(contextJWTLogin).(string)

		balance, err := s.repository.Balance(r.Context(), jwtOwner)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response, err := json.Marshal(&balance)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", contentTypeJSON)
		w.WriteHeader(http.StatusOK)

		_, err = w.Write(response)
		if err != nil {
			log.Trace("Log in prod")
		}
	}
}

func withdrawHandler(s Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if r.Header.Get("Content-Type") != contentTypeJSON {
			http.Error(w, "Bad content type", http.StatusBadRequest)
			return
		}

		var request WithdrawRequest
		if err = json.Unmarshal(body, &request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(request.Order) == 0 {
			http.Error(w, "Order must not be null", http.StatusBadRequest)
			return
		}

		order, err := strconv.ParseInt(request.Order, 10, 64)
		if err != nil {
			http.Error(w, "Bad order number", http.StatusUnprocessableEntity)
			return
		}

		if !luhn.Valid(int(order)) {
			http.Error(w, "Bad order number", http.StatusUnprocessableEntity)
			return
		}

		jwtLogin, _ := r.Context().Value(contextJWTLogin).(string)

		balance, err := s.repository.Balance(r.Context(), jwtLogin)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if request.Sum > balance.Current {
			http.Error(w, "No money - no honey", http.StatusPaymentRequired)
			return
		}

		err = s.repository.Withdraw(r.Context(), jwtLogin, order, request.Sum)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func withdrawalsHandler(s Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jwtLogin, _ := r.Context().Value(contextJWTLogin).(string)

		withdrawals, err := s.repository.Withdrawals(r.Context(), jwtLogin)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(withdrawals) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		response, err := json.Marshal(&withdrawals)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", contentTypeJSON)
		w.WriteHeader(http.StatusOK)

		_, err = w.Write(response)
		if err != nil {
			log.Trace("Log in prod")
		}
	}
}
