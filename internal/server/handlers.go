package server

import (
	"encoding/json"
	"io"
	"net/http"
)

type UserAuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
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

		w.WriteHeader(http.StatusOK)
	}
}
