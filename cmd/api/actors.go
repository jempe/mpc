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

func (app *application) createActorHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name       string    `json:"name"`
		Gender     string    `json:"gender"`
		BirthDate  time.Time `json:"birth_date"`
		BirthPlace string    `json:"birth_place"`
		Biography  string    `json:"biography"`
		Height     int       `json:"height"`
		ImageURL   string    `json:"image_url"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	actor := &data.Actor{
		Name:       input.Name,
		Gender:     input.Gender,
		BirthDate:  input.BirthDate,
		BirthPlace: input.BirthPlace,
		Biography:  input.Biography,
		Height:     input.Height,
		ImageURL:   input.ImageURL,
	}

	v := validator.New()

	if data.ValidateActor(v, actor, validator.ActionCreate); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Actors.Insert(actor)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/actors/%d", actor.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"actor": actor}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showActorHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	actor, err := app.models.Actors.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"actor": actor}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateActorHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	actor, err := app.models.Actors.Get(id)
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
		if strconv.FormatInt(int64(actor.Version), 32) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r)
			return
		}
	}

	var input struct {
		Name       *string    `json:"name"`
		Gender     *string    `json:"gender"`
		BirthDate  *time.Time `json:"birth_date"`
		BirthPlace *string    `json:"birth_place"`
		Biography  *string    `json:"biography"`
		Height     *int       `json:"height"`
		ImageURL   *string    `json:"image_url"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Name != nil {
		actor.Name = *input.Name
	}

	if input.Gender != nil {
		actor.Gender = *input.Gender
	}

	if input.BirthDate != nil {
		actor.BirthDate = *input.BirthDate
	}

	if input.BirthPlace != nil {
		actor.BirthPlace = *input.BirthPlace
	}

	if input.Biography != nil {
		actor.Biography = *input.Biography
	}

	if input.Height != nil {
		actor.Height = *input.Height
	}

	if input.ImageURL != nil {
		actor.ImageURL = *input.ImageURL
	}

	v := validator.New()

	if data.ValidateActor(v, actor, validator.ActionUpdate); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Actors.Update(actor)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"actor": actor}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteActorHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Actors.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "actor successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listActorHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name   string
		Gender string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Name = app.readString(qs, "name", "")

	input.Gender = app.readString(qs, "gender", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{
		"id",
		"name",
		"gender",
		"birth_date",
		"height",
		"-id",
		"-name",
		"-gender",
		"-birth_date",
		"-height",
	}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	actors, metadata, err := app.models.Actors.GetAll(
		input.Name,
		input.Gender,
		input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"actors": actors, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
