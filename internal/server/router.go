package server

import (
	"github.com/NYTimes/gziphandler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func router(s Server) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(DecompressGZIP)
	r.Use(gziphandler.GzipHandler)

	r.MethodNotAllowed(badRequestHandler)
	r.NotFound(badRequestHandler)

	return r
}
