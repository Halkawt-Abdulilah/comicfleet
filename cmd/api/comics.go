package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"comicfleet.halkawtabdulilah.net/internal/data"
	"comicfleet.halkawtabdulilah.net/internal/validator"
)

func (app *application) listComicsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")

	input.Filters.SortSafelist = []string{"id", "title", "year", "volumes", "-id", "-title", "-year", "-volumes"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	comics, metadata, err := app.models.Comics.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"metadata": metadata, "comics": comics}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createComicHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Volumes data.Volumes `json:"volumes"`
		Genres  []string     `json:"genres"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	comic := &data.Comic{
		Title:   input.Title,
		Year:    input.Year,
		Volumes: input.Volumes,
		Genres:  input.Genres,
	}

	v := validator.New()

	if data.ValidateComic(v, comic); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Comics.Insert(comic)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/comics/%d", comic.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"comic": comic}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showComicHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	comic, err := app.models.Comics.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, data.RecordNotFoundError):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"comic": comic}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) updateComicHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)

	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	comic, err := app.models.Comics.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.RecordNotFoundError):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	//if the client passes version number they expect in If-Not-Match / X-Expected-Version header:
	if r.Header.Get("X-Expected-Version") != "" {
		if strconv.FormatInt(int64(comic.Version), 32) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r)
			return
		}
	}
	//

	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Volumes *data.Volumes `json:"volumes"`
		Genres  []string      `json:"genres"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		comic.Title = *input.Title
	}
	if input.Year != nil {
		comic.Year = *input.Year
	}
	if input.Volumes != nil {
		comic.Volumes = *input.Volumes
	}

	if input.Genres != nil {
		comic.Genres = input.Genres
	}

	v := validator.New()

	if data.ValidateComic(v, comic); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Comics.Update(comic)
	if err != nil {
		switch {
		case errors.Is(err, data.EditConflictError):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"comic": comic}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteComicHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Comics.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.RecordNotFoundError):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	//alt: 200 OK and message
	err = app.writeJSON(w, http.StatusNoContent, envelope{}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
