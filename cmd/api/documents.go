package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jempe/mpc/internal/data"
	"github.com/jempe/mpc/internal/validator"
	"github.com/pgvector/pgvector-go"
)

func (app *application) createDocumentHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title        string `json:"title"`
		Content      string `json:"content"`
		Tokens       int    `json:"tokens"`
		Sequence     int    `json:"sequence"`
		ContentField string `json:"content_field"`
		VideoID      int64  `json:"video_id"`
		CategoryID   int64  `json:"category_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	document := &data.Document{
		Title:        input.Title,
		Content:      input.Content,
		Tokens:       input.Tokens,
		Sequence:     input.Sequence,
		ContentField: input.ContentField,
		VideoID:      input.VideoID,
		CategoryID:   input.CategoryID,
	}

	if input.Tokens == 0 {
		countedTokens, err := app.countTokens(input.Content)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		document.Tokens = countedTokens
	}

	v := validator.New()

	if data.ValidateDocument(v, document, validator.ActionCreate); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Documents.Insert(document)
	if err != nil {
		app.handleCustomDocumentErrors(err, w, r, v)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/documents/%d", document.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"document": document}, headers)
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

	document, err := app.models.Documents.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"document": document}, nil)

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

	document, err := app.models.Documents.Get(id)
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
		if strconv.FormatInt(int64(document.Version), 32) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r)
			return
		}
	}

	var input struct {
		Title        *string `json:"title"`
		Content      *string `json:"content"`
		Tokens       *int    `json:"tokens"`
		Sequence     *int    `json:"sequence"`
		ContentField *string `json:"content_field"`
		VideoID      *int64  `json:"video_id"`
		CategoryID   *int64  `json:"category_id"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		document.Title = *input.Title
	}

	if input.Tokens != nil {
		document.Tokens = *input.Tokens
	}

	if input.Content != nil {
		document.Content = *input.Content

		countedTokens, err := app.countTokens(document.Content)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		document.Tokens = countedTokens
	}

	if input.Sequence != nil {
		document.Sequence = *input.Sequence
	}

	if input.ContentField != nil {
		document.ContentField = *input.ContentField
	}

	if input.VideoID != nil {
		document.VideoID = *input.VideoID
	}

	if input.CategoryID != nil {
		document.CategoryID = *input.CategoryID
	}

	v := validator.New()

	if data.ValidateDocument(v, document, validator.ActionUpdate); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Documents.Update(document)
	if err != nil {

		if errors.Is(err, data.ErrEditConflict) {
			app.editConflictResponse(w, r)
			app.handleCustomDocumentErrors(err, w, r, v)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"document": document}, nil)
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
		VideoID    int64
		CategoryID int64
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.VideoID = app.readInt64(qs, "video_id", 0, v)

	input.CategoryID = app.readInt64(qs, "category_id", 0, v)

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{
		"id",
		"video_id",
		"category_id",
		"-id",
		"-video_id",
		"-category_id",
	}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	documents, metadata, err := app.models.Documents.GetAll(
		input.VideoID,
		input.CategoryID,
		input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"documents": documents, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

/*handle_custom_errors_start*/

func (app application) handleCustomDocumentErrors(err error, w http.ResponseWriter, r *http.Request, v *validator.Validator) {
	switch {
	//	case errors.Is(err, data.ErrDuplicateDocumentTitleEn):
	//		v.AddError("title_en", "a title with this name already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateDocumentTitleEs):
	//		v.AddError("title_es", "a title with this name already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateDocumentTitleFr):
	//		v.AddError("title_fr", "a title with this name already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateDocumentURLEn):
	//		v.AddError("url_en", "a video with this URL already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateDocumentURLEs):
	//		v.AddError("url_es", "a video with this URL already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateDocumentURLFr):
	//		v.AddError("url_fr", "a video with this URL already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateDocumentFolder):
	//		v.AddError("folder", "a video with this folder already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	default:
		app.serverErrorResponse(w, r, err)
	}
}

/*handle_custom_errors_end*/
/*list_document_semantic_start*/
func (app *application) listDocumentSemanticHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Search             string
		Similarity         float64
		EmbeddingsProvider string
		ContentFields      []string
		VideoID            int64
		CategoryID         int64
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	defaultEmbeddingsProvider := app.config.embeddings.defaultProvider

	input.Search = app.readString(qs, "search", "")
	input.Similarity = app.readFloat(qs, "similarity", 0.7, v)
	input.EmbeddingsProvider = app.readString(qs, "embeddings-provider", defaultEmbeddingsProvider)

	input.ContentFields = app.readCSV(qs, "content_fields", []string{})

	//Additional Semantic Search Filters
	input.VideoID = app.readInt64(qs, "video_id", 0, v)
	input.CategoryID = app.readInt64(qs, "category_id", 0, v)

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 5, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{
		"id",
		"-id",
	}

	if input.Search == "" {
		app.serverErrorResponse(w, r, errors.New("missing required search parameter"))
		return
	}

	if !(input.EmbeddingsProvider == "sentence-transformers" || input.EmbeddingsProvider == "openai" || input.EmbeddingsProvider == "google") {
		app.serverErrorResponse(w, r, errors.New("invalid embeddings provider"))
		return
	}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	searchInput := []string{
		input.Search,
	}

	var embeddings [][]float32
	var err error
	if input.EmbeddingsProvider == "sentence-transformers" {
		embeddings, err = app.fetchSentenceTransformersEmbeddings(searchInput)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	if len(embeddings) == 0 {
		app.serverErrorResponse(w, r, errors.New("no embeddings returned"))
		return
	}

	documents, metadata, err := app.models.Documents.GetAllSemantic(
		pgvector.NewVector(embeddings[0]),
		input.Similarity,
		input.EmbeddingsProvider,
		input.ContentFields,
		input.VideoID,
		input.CategoryID,
		input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"documents": documents, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

/*list_document_semantic_end*/
