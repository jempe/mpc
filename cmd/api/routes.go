package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/videos", app.listVideoHandler)
	router.HandlerFunc(http.MethodPost, "/v1/videos", app.createVideoHandler)
	router.HandlerFunc(http.MethodGet, "/v1/videos/:id", app.showVideoHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/videos/:id", app.updateVideoHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/videos/:id", app.deleteVideoHandler)

	router.HandlerFunc(http.MethodGet, "/v1/categories", app.listCategoryHandler)
	router.HandlerFunc(http.MethodPost, "/v1/categories", app.createCategoryHandler)
	router.HandlerFunc(http.MethodGet, "/v1/categories/:id", app.showCategoryHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/categories/:id", app.updateCategoryHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/categories/:id", app.deleteCategoryHandler)

	router.HandlerFunc(http.MethodGet, "/v1/actors", app.listActorHandler)
	router.HandlerFunc(http.MethodPost, "/v1/actors", app.createActorHandler)
	router.HandlerFunc(http.MethodGet, "/v1/actors/:id", app.showActorHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/actors/:id", app.updateActorHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/actors/:id", app.deleteActorHandler)

	router.HandlerFunc(http.MethodGet, "/v1/documents", app.listDocumentHandler)
	router.HandlerFunc(http.MethodPost, "/v1/documents", app.createDocumentHandler)
	router.HandlerFunc(http.MethodGet, "/v1/documents/:id", app.showDocumentHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/documents/:id", app.updateDocumentHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/documents/:id", app.deleteDocumentHandler)

	router.HandlerFunc(http.MethodGet, "/v1/documents_search", app.listDocumentSemanticHandler)

	//custom_routes

	return app.recoverPanic(app.basicAuth(app.cors(app.rateLimit(router))))
}
