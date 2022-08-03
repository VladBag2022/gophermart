package server

type contextKey string

const (
	contextJWTLogin     contextKey = "login"
	contentTypeJSON     string     = "application/json"
	authorizationHeader string     = "Authorization"
)
