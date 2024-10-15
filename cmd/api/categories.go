package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jempe/mpc/internal/data"
	"github.com/jempe/mpc/internal/validator"
)

func (app *application) createCategoryHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name                 string `json:"name"`
		EnableSemanticSearch bool   `json:"enable_semantic_search"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	category := &data.Category{
		Name:                 input.Name,
		EnableSemanticSearch: input.EnableSemanticSearch,
	}

	v := validator.New()

	if data.ValidateCategory(v, category, validator.ActionCreate); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Categories.Insert(category)
	if err != nil {
		app.handleCustomCategoryErrors(err, w, r, v)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/categories/%d", category.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"category": category}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showCategoryHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	category, err := app.models.Categories.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"category": category}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	category, err := app.models.Categories.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if r.Header.Get("X-Expected-Version") != "" {
		if strconv.FormatInt(int64(category.Version), 32) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r)
			return
		}
	}

	var input struct {
		Name                 *string `json:"name"`
		EnableSemanticSearch *bool   `json:"enable_semantic_search"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Name != nil {
		category.Name = *input.Name
	}

	if input.EnableSemanticSearch != nil {
		category.EnableSemanticSearch = *input.EnableSemanticSearch
	}

	v := validator.New()

	if data.ValidateCategory(v, category, validator.ActionUpdate); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Categories.Update(category)
	if err != nil {

		if errors.Is(err, data.ErrEditConflict) {
			app.editConflictResponse(w, r)
			app.handleCustomCategoryErrors(err, w, r, v)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"category": category}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Categories.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "category successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listCategoryHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name                 string
		EnableSemanticSearch bool
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Name = app.readString(qs, "name", "")
	input.EnableSemanticSearch = app.readBool(qs, "enable_semantic_search", false)

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{
		"id",
		"name",
		"-id",
		"-name",
	}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	categories, metadata, err := app.models.Categories.GetAll(
		input.Name,
		input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"categories": categories, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

/*handle_custom_errors_start*/

func (app application) handleCustomCategoryErrors(err error, w http.ResponseWriter, r *http.Request, v *validator.Validator) {
	switch {
	//	case errors.Is(err, data.ErrDuplicateCategoryTitleEn):
	//		v.AddError("title_en", "a title with this name already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateCategoryTitleEs):
	//		v.AddError("title_es", "a title with this name already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateCategoryTitleFr):
	//		v.AddError("title_fr", "a title with this name already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateCategoryURLEn):
	//		v.AddError("url_en", "a video with this URL already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateCategoryURLEs):
	//		v.AddError("url_es", "a video with this URL already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateCategoryURLFr):
	//		v.AddError("url_fr", "a video with this URL already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateCategoryFolder):
	//		v.AddError("folder", "a video with this folder already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	default:
		app.serverErrorResponse(w, r, err)
	}
}

/*handle_custom_errors_end*/
