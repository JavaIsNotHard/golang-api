package main

import (
	"api/internal/data"
    "errors"
    "fmt"
	"api/internal/validator"
	"net/http"
)


func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
    var input struct {
        Title string `json:"title"`
        Year  int32  `json:"year"`
        Runtime data.Runtime `json:"runtime"`
        Genres []string `json:"genres"`
    }

    err := app.ReadJSON(w, r, &input)
    if err != nil {
        app.badRequestResponse(w, r, err)
        return
    }

    movie := &data.Movie {
        Title: input.Title, 
        Year: input.Year,
        Runtime: input.Runtime, 
        Genres: input.Genres,
    }

    v := validator.New()
    
    if data.ValidateMovie(v, movie); !v.Valid() {
        app.failedValidationResponse(w, r, v.Errors)
        return
    }

    err = app.models.Movies.Insert(movie)
    if err != nil {
        app.serverErrorResponse(w, r, err)
        return
    }

    headers := make(http.Header)
    headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

    err = app.WriteJSON(w, envelope{"movie": movie}, headers, http.StatusCreated)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
    id, err := app.getIDparam(r)
    
    if err != nil {
        app.notFoundErrorResponse(w, r)
        return
    }

    movie, err := app.models.Movies.Get(id)
    if err != nil {
        switch {
        case errors.Is(err, data.ErrRecordNotFound):
            app.notFoundErrorResponse(w, r)
        default:
            app.serverErrorResponse(w, r, err)
        }
        return
    }

    err = app.WriteJSON(w, envelope{"movie": movie}, nil, http.StatusOK)
    if err != nil {
        app.logger.Error(err.Error())
        app.serverErrorResponse(w, r, err)
    }
}

func (app *application) updateMoviehandler(w http.ResponseWriter, r *http.Request) {
    id, err := app.getIDparam(r)
    if err != nil {
        app.notFoundErrorResponse(w, r)
        return
    }

    movie, err := app.models.Movies.Get(id)
    if err != nil {
        switch {
        case errors.Is(err, data.ErrRecordNotFound):
            app.notFoundErrorResponse(w, r)
        default:
            app.serverErrorResponse(w, r, err)
        }
        return
    }

    var input struct {
        Title       *string      `json:"title"`
        Year        *int32       `json:"year, omitempty"`
        Runtime     *data.Runtime     `json:"runtime, omitempty"`
        Genres      []string    `json:"genres, omitempty"`
    }

    err = app.ReadJSON(w, r, &input)
    if err != nil {
        app.badRequestResponse(w, r, err)
        return
    }

    // dereference the value because we are pointing to the location using a pointer
    // we are using pointer so that nil comparison is possible
    if input.Title != nil {
        movie.Title = *input.Title
    }

    if input.Year != nil {
        movie.Year = *input.Year
    }

    if input.Runtime != nil {
        movie.Runtime = *input.Runtime
    }

    if input.Genres != nil {
        movie.Genres = input.Genres
    }

    v := validator.New()

    if data.ValidateMovie(v, movie); !v.Valid() {
        app.failedValidationResponse(w, r, v.Errors)
        return
    }

    err = app.models.Movies.Update(movie)
    if err != nil {
        app.serverErrorResponse(w, r, err)
        return
    }

    err = app.WriteJSON(w, envelope{"movie": movie}, nil, http.StatusOK)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
    id, err := app.getIDparam(r)
    if err != nil {
        app.notFoundErrorResponse(w, r)
        return
    }

    err = app.models.Movies.Delete(id)
    if err != nil {
        switch {
        case errors.Is(err, data.ErrRecordNotFound):
            app.notFoundErrorResponse(w, r)
        default:
            app.serverErrorResponse(w, r, err)
        }
        return
    }

    err = app.WriteJSON(w, envelope{"message": "item deleted successfully"}, nil, http.StatusOK)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}

func (app *application) listMovieHandler(w http.ResponseWriter, r *http.Request) {
    var input struct {
        Title string 
        Genres []string 
        data.Filters
    }

    v := validator.New()

    query := r.URL.Query()

    input.Title = app.readString(query, "", "title")
    input.Genres = app.readCommanSeparateValue(query, []string{}, "genres")
    input.Filters.Page = app.readInt(query, 1, "page", v)
    input.Filters.PageSize = app.readInt(query, 20, "page_size", v)
    input.Filters.Sort = app.readString(query, "id", "sort")
    input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

    if data.ValidateFilters(v, input.Filters); !v.Valid() {
        app.failedValidationResponse(w, r, v.Errors)
        return
    }

    movies, err := app.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
    if err != nil {
        app.serverErrorResponse(w, r, err)
        return
    }

    err = app.WriteJSON(w, envelope{"movies": movies}, nil, http.StatusOK)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}
