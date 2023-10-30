package app

import (
	"legocerthub-backend/pkg/output"
	"net/http"
)

// handlerNotFound is called when there is not a matching route on the router
func (app *Application) handlerNotFound() http.Handler {
	// the base handler function (before middleware)
	handlerFunc := func(w http.ResponseWriter, r *http.Request) error {
		// return 404 not found
		err := app.output.WriteErrorJSON(w, output.ErrNotFound)
		if err != nil {
			return err
		}

		return nil
	}

	// Add Middleware

	// NO CORS
	// no cors info to provide if route is 404

	// Logger / handle custom handler func's error
	httpHandlerFunc := middlewareApplyErrorHandling(handlerFunc, false, app.logger.SugaredLogger, app.output)

	return httpHandlerFunc
}

// handlerGlobalOptions is called to respond to OPTIONS requests. This is
// particularly important for CORS.
func (app *Application) handlerGlobalOptions() http.Handler {
	// the base handler function (before middleware)
	handlerFunc := func(w http.ResponseWriter, r *http.Request) error {
		// OPTIONS should always return a response to prevent preflight errors
		// see: https://stackoverflow.com/questions/52047548/response-for-preflight-does-not-have-http-ok-status-in-angular

		// Note: Disabled as CORS will write the headers automatically
		// w.WriteHeader(http.StatusNoContent)

		return nil
	}

	// Add Middleware

	// CORS
	handlerFunc = middlewareApplyCORS(handlerFunc, app.config.CORSPermittedCrossOrigins)

	// Logger / handle custom handler func's error
	httpHandlerFunc := middlewareApplyErrorHandling(handlerFunc, false, app.logger.SugaredLogger, app.output)

	return httpHandlerFunc
}
