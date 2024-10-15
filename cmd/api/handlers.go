package main

import "net/http"

func (app *application) userPageHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Title = "User Page"
	app.render(w, r, http.StatusOK, "auth_pages.tmpl", data)
}

func (app *application) homeHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Title = "Dashboard"
	app.render(w, r, http.StatusOK, "dashboard.tmpl", data)
}

func (app *application) videosPageHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Title = "Videos"
	app.render(w, r, http.StatusOK, "videos.tmpl", data)
}

func (app *application) videoPageHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Title = "Video"
	app.render(w, r, http.StatusOK, "videos_item.tmpl", data)
}

func (app *application) categoriesPageHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Title = "Categories"
	app.render(w, r, http.StatusOK, "categories.tmpl", data)
}

func (app *application) categoryPageHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Title = "Category"
	app.render(w, r, http.StatusOK, "categories_item.tmpl", data)
}

func (app *application) actorsPageHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Title = "Actors"
	app.render(w, r, http.StatusOK, "actors.tmpl", data)
}

func (app *application) actorPageHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Title = "Actor"
	app.render(w, r, http.StatusOK, "actors_item.tmpl", data)
}

func (app *application) documentsPageHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Title = "Documents"
	app.render(w, r, http.StatusOK, "documents.tmpl", data)
}

func (app *application) documentPageHandler(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Title = "Document"
	app.render(w, r, http.StatusOK, "documents_item.tmpl", data)
}
