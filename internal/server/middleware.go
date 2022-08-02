package server

import (
	"compress/gzip"
	"net/http"
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
