package main

import (
	"net/http"

	"github.com/jempe/mpc/ui"
	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/admin/", app.homeHandler)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/videos", app.requireActivatedUser(app.listVideoHandler))
	router.HandlerFunc(http.MethodPost, "/v1/videos", app.requireActivatedUser(app.createVideoHandler))
	router.HandlerFunc(http.MethodGet, "/v1/videos/:id", app.requireActivatedUser(app.showVideoHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/videos/:id", app.requireActivatedUser(app.updateVideoHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/videos/:id", app.requireActivatedUser(app.deleteVideoHandler))

	router.HandlerFunc(http.MethodGet, "/v1/categories", app.requireActivatedUser(app.listCategoryHandler))
	router.HandlerFunc(http.MethodPost, "/v1/categories", app.requireActivatedUser(app.createCategoryHandler))
	router.HandlerFunc(http.MethodGet, "/v1/categories/:id", app.requireActivatedUser(app.showCategoryHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/categories/:id", app.requireActivatedUser(app.updateCategoryHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/categories/:id", app.requireActivatedUser(app.deleteCategoryHandler))

	router.HandlerFunc(http.MethodGet, "/v1/actors", app.requireActivatedUser(app.listActorHandler))
	router.HandlerFunc(http.MethodPost, "/v1/actors", app.requireActivatedUser(app.createActorHandler))
	router.HandlerFunc(http.MethodGet, "/v1/actors/:id", app.requireActivatedUser(app.showActorHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/actors/:id", app.requireActivatedUser(app.updateActorHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/actors/:id", app.requireActivatedUser(app.deleteActorHandler))

	router.HandlerFunc(http.MethodGet, "/v1/documents", app.requireActivatedUser(app.listDocumentHandler))
	router.HandlerFunc(http.MethodPost, "/v1/documents", app.requireActivatedUser(app.createDocumentHandler))
	router.HandlerFunc(http.MethodGet, "/v1/documents/:id", app.requireActivatedUser(app.showDocumentHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/documents/:id", app.requireActivatedUser(app.updateDocumentHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/documents/:id", app.requireActivatedUser(app.deleteDocumentHandler))

	router.HandlerFunc(http.MethodGet, "/v1/documents_search", app.requireActivatedUser(app.listDocumentSemanticHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/password", app.updateUserPasswordHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/activation", app.createActivationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/password_reset", app.createPasswordResetTokenHandler)

	router.HandlerFunc(http.MethodGet, "/admin/login.html", app.userPageHandler)
	router.HandlerFunc(http.MethodGet, "/admin/signup.html", app.userPageHandler)
	router.HandlerFunc(http.MethodGet, "/admin/activate.html", app.userPageHandler)
	router.HandlerFunc(http.MethodGet, "/admin/reset_password.html", app.userPageHandler)
	router.HandlerFunc(http.MethodGet, "/admin/forgot_password.html", app.userPageHandler)
	router.HandlerFunc(http.MethodGet, "/admin/request_activation.html", app.userPageHandler)

	router.HandlerFunc(http.MethodGet, "/admin/videos.html", app.videosPageHandler)
	router.HandlerFunc(http.MethodGet, "/admin/video.html", app.videoPageHandler)

	router.HandlerFunc(http.MethodGet, "/admin/categories.html", app.categoriesPageHandler)
	router.HandlerFunc(http.MethodGet, "/admin/category.html", app.categoryPageHandler)

	router.HandlerFunc(http.MethodGet, "/admin/actors.html", app.actorsPageHandler)
	router.HandlerFunc(http.MethodGet, "/admin/actor.html", app.actorPageHandler)

	router.HandlerFunc(http.MethodGet, "/admin/documents.html", app.documentsPageHandler)
	router.HandlerFunc(http.MethodGet, "/admin/document.html", app.documentPageHandler)

	router.Handler(http.MethodGet, "/static/*filepath", http.FileServerFS(ui.Files))

	//custom_routes

	return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router))))
}
