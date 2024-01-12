package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	//comics endpoints
	router.HandlerFunc(http.MethodGet, "/v1/comics", app.requirePermission("comics:read", app.listComicsHandler))
	router.HandlerFunc(http.MethodPost, "/v1/comics", app.requirePermission("comics:write", app.createComicHandler))
	router.HandlerFunc(http.MethodGet, "/v1/comics/:id", app.requirePermission("comics:read", app.showComicHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/comics/:id", app.requirePermission("comics:write", app.updateComicHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/comics/:id", app.requirePermission("comics:write", app.deleteComicHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activate", app.activateUserHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authenticate", app.createAuthenticationTokenHandler)

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
