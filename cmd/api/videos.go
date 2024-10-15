package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jempe/mpc/internal/data"
	"github.com/jempe/mpc/internal/validator"
)

func (app *application) createVideoHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name                 string    `json:"name"`
		ThumbURL             string    `json:"thumb_url"`
		ImageURL             string    `json:"image_url"`
		VideoURL             string    `json:"video_url"`
		SubtitlesURL         string    `json:"subtitles_url"`
		Description          string    `json:"description"`
		ReleaseDate          time.Time `json:"release_date"`
		Width                int       `json:"width"`
		Height               int       `json:"height"`
		Duration             int       `json:"duration"`
		Sequence             int       `json:"sequence"`
		File                 string    `json:"file"`
		OriginalFile         string    `json:"original_file"`
		Path                 string    `json:"path"`
		Md5sum               string    `json:"md5sum"`
		EnableSemanticSearch bool      `json:"enable_semantic_search"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	video := &data.Video{
		Name:                 input.Name,
		ThumbURL:             input.ThumbURL,
		ImageURL:             input.ImageURL,
		VideoURL:             input.VideoURL,
		SubtitlesURL:         input.SubtitlesURL,
		Description:          input.Description,
		ReleaseDate:          input.ReleaseDate,
		Width:                input.Width,
		Height:               input.Height,
		Duration:             input.Duration,
		Sequence:             input.Sequence,
		File:                 input.File,
		OriginalFile:         input.OriginalFile,
		Path:                 input.Path,
		Md5sum:               input.Md5sum,
		EnableSemanticSearch: input.EnableSemanticSearch,
	}

	v := validator.New()

	if data.ValidateVideo(v, video, validator.ActionCreate); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Videos.Insert(video)
	if err != nil {
		app.handleCustomVideoErrors(err, w, r, v)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/videos/%d", video.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"video": video}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showVideoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	video, err := app.models.Videos.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"video": video}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateVideoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	video, err := app.models.Videos.Get(id)
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
		if strconv.FormatInt(int64(video.Version), 32) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r)
			return
		}
	}

	var input struct {
		Name                 *string    `json:"name"`
		ThumbURL             *string    `json:"thumb_url"`
		ImageURL             *string    `json:"image_url"`
		VideoURL             *string    `json:"video_url"`
		SubtitlesURL         *string    `json:"subtitles_url"`
		Description          *string    `json:"description"`
		ReleaseDate          *time.Time `json:"release_date"`
		Width                *int       `json:"width"`
		Height               *int       `json:"height"`
		Duration             *int       `json:"duration"`
		Sequence             *int       `json:"sequence"`
		File                 *string    `json:"file"`
		OriginalFile         *string    `json:"original_file"`
		Path                 *string    `json:"path"`
		Md5sum               *string    `json:"md5sum"`
		EnableSemanticSearch *bool      `json:"enable_semantic_search"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Name != nil {
		video.Name = *input.Name
	}

	if input.ThumbURL != nil {
		video.ThumbURL = *input.ThumbURL
	}

	if input.ImageURL != nil {
		video.ImageURL = *input.ImageURL
	}

	if input.VideoURL != nil {
		video.VideoURL = *input.VideoURL
	}

	if input.SubtitlesURL != nil {
		video.SubtitlesURL = *input.SubtitlesURL
	}

	if input.Description != nil {
		video.Description = *input.Description
	}

	if input.ReleaseDate != nil {
		video.ReleaseDate = *input.ReleaseDate
	}

	if input.Width != nil {
		video.Width = *input.Width
	}

	if input.Height != nil {
		video.Height = *input.Height
	}

	if input.Duration != nil {
		video.Duration = *input.Duration
	}

	if input.Sequence != nil {
		video.Sequence = *input.Sequence
	}

	if input.File != nil {
		video.File = *input.File
	}

	if input.OriginalFile != nil {
		video.OriginalFile = *input.OriginalFile
	}

	if input.Path != nil {
		video.Path = *input.Path
	}

	if input.Md5sum != nil {
		video.Md5sum = *input.Md5sum
	}

	if input.EnableSemanticSearch != nil {
		video.EnableSemanticSearch = *input.EnableSemanticSearch
	}

	v := validator.New()

	if data.ValidateVideo(v, video, validator.ActionUpdate); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Videos.Update(video)
	if err != nil {

		if errors.Is(err, data.ErrEditConflict) {
			app.editConflictResponse(w, r)
			app.handleCustomVideoErrors(err, w, r, v)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"video": video}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteVideoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Videos.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "video successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listVideoHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name                 string
		File                 string
		OriginalFile         string
		Md5sum               string
		EnableSemanticSearch bool
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Name = app.readString(qs, "name", "")

	input.File = app.readString(qs, "file", "")

	input.OriginalFile = app.readString(qs, "original_file", "")

	input.Md5sum = app.readString(qs, "md5sum", "")
	input.EnableSemanticSearch = app.readBool(qs, "enable_semantic_search", false)

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{
		"id",
		"name",
		"release_date",
		"sequence",
		"-id",
		"-name",
		"-release_date",
		"-sequence",
	}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	videos, metadata, err := app.models.Videos.GetAll(
		input.Name,
		input.File,
		input.OriginalFile,
		input.Md5sum,
		input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"videos": videos, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

/*handle_custom_errors_start*/

func (app application) handleCustomVideoErrors(err error, w http.ResponseWriter, r *http.Request, v *validator.Validator) {
	switch {
	//	case errors.Is(err, data.ErrDuplicateVideoTitleEn):
	//		v.AddError("title_en", "a title with this name already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateVideoTitleEs):
	//		v.AddError("title_es", "a title with this name already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateVideoTitleFr):
	//		v.AddError("title_fr", "a title with this name already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateVideoURLEn):
	//		v.AddError("url_en", "a video with this URL already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateVideoURLEs):
	//		v.AddError("url_es", "a video with this URL already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateVideoURLFr):
	//		v.AddError("url_fr", "a video with this URL already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	//	case errors.Is(err, data.ErrDuplicateVideoFolder):
	//		v.AddError("folder", "a video with this folder already exists")
	//		app.failedValidationResponse(w, r, v.Errors)
	default:
		app.serverErrorResponse(w, r, err)
	}
}

/*handle_custom_errors_end*/
