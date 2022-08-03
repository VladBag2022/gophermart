package server

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func rootRouter(s Server) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(DecompressGZIP)
	r.Use(gziphandler.GzipHandler)

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", registerHandler(s))
		r.Post("/login", loginHandler(s))

		r.Mount("/", func(s Server) http.Handler {
			ra := chi.NewRouter()
			ra.Use(CheckJWT(s))

			ra.Post("/orders", uploadHandler(s))

			return ra
		}(s))
	})

	r.MethodNotAllowed(badRequestHandler)
	r.NotFound(badRequestHandler)

	return r
}
