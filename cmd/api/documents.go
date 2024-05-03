package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jempe/mpc/internal/data"
	"github.com/jempe/mpc/internal/validator"
)

func (app *application) createDocumentHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Content      string `json:"content"`
		Tokens       int    `json:"tokens"`
		Sequence     int    `json:"sequence"`
		ContentField string `json:"content_field"`
		VideoID      int64  `json:"video_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	Document := &data.Document{
		Content:      input.Content,
		Tokens:       input.Tokens,
		Sequence:     input.Sequence,
		ContentField: input.ContentField,
		VideoID:      input.VideoID,
	}

	if input.Tokens == 0 {
		countedTokens, err := app.countTokens(input.Content)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		Document.Tokens = countedTokens
	}

	v := validator.New()

	if data.ValidateDocument(v, Document, validator.ActionCreate); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Documents.Insert(Document)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/documents/%d", Document.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"document": Document}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showDocumentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	Document, err := app.models.Documents.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"document": Document}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateDocumentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	Document, err := app.models.Documents.Get(id)
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
		if strconv.FormatInt(int64(Document.Version), 32) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r)
			return
		}
	}

	var input struct {
		Content      *string `json:"content"`
		Tokens       *int    `json:"tokens"`
		Sequence     *int    `json:"sequence"`
		ContentField *string `json:"content_field"`
		VideoID      *int64  `json:"video_id"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Content != nil {
		Document.Content = *input.Content
	}

	if input.Tokens != nil {
		Document.Tokens = *input.Tokens
	} else if input.Content != nil {
		countedTokens, err := app.countTokens(Document.Content)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		Document.Tokens = countedTokens
	}

	if input.Sequence != nil {
		Document.Sequence = *input.Sequence
	}

	if input.ContentField != nil {
		Document.ContentField = *input.ContentField
	}

	if input.VideoID != nil {
		Document.VideoID = *input.VideoID
	}

	v := validator.New()

	if data.ValidateDocument(v, Document, validator.ActionUpdate); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Documents.Update(Document)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"document": Document}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteDocumentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Documents.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "document successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listDocumentHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ContentField string
		VideoID      int64
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.ContentField = app.readString(qs, "content_field", "")

	input.VideoID = app.readInt64(qs, "video_id", 0, v)

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{
		"id",
		"content_field",
		"video_id",
		"-id",
		"-content_field",
		"-video_id",
	}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	Documents, metadata, err := app.models.Documents.GetAll(
		input.ContentField,
		input.VideoID,
		input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"documents": Documents, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
