package app

import (
	"net/http"

	"github.com/rs/cors"
)

// enableCORS applies CORS to an http.Handler and is intended to wrap the router
// if no cross origins are permitted, this function is a no-op and just returns
// next
func (app *Application) enableCORS(next http.Handler) http.Handler {
	// are any cross origins allowed? if not, do not use CORS
	if app.config.CORSPermittedCrossOrigins == nil {
		return next
	}

	// set up CORS
	c := cors.New(cors.Options{
		// permitted cross origins
		// WARNING: nil / empty slice == allow all!
		AllowedOrigins: app.config.CORSPermittedCrossOrigins,

		// credentials must be allowed for access to work properly
		AllowCredentials: true,

		// allowed request headers (client can send to server)
		AllowedHeaders: []string{
			// general
			"content-type",

			// access token
			"authorization",

			// pem download authentication
			"X-API-Key", "apiKey",

			// conditionals for pem downloads
			"if-match", "if-modified-since", "if-none-match", "if-range", "if-unmodified-since",

			// retry tracker for refresh token logic on frontend
			"x-no-retry",
		},

		// allowed methods the client can send to the server
		AllowedMethods: []string{http.MethodDelete, http.MethodGet, http.MethodHead,
			http.MethodPost, http.MethodPut},

		// headers for client to expose to the cross origin requester (in server response)
		ExposedHeaders: []string{
			// general
			"content-length", "content-security-policy", "content-type", "strict-transport-security", "x-content-type-options", "x-frame-options",

			// set name of file when client downloads something (used with pem, zip)
			"content-disposition",

			// conditionals for pem downloads
			"last-modified", "etag",
		},
	})

	return c.Handler(next)
}