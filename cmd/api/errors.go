package main 

import (
    "fmt"
    "net/http"
)

func (app *application) logError(r *http.Request, err error) {
    var (
        method = r.Method
        uri = r.URL.RequestURI()
    )

    app.logger.Error(err.Error(), "method", method, "uri", uri)
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
    env := envelope{"error": message}

    err := app.WriteJSON(w, env, nil, status) 
    if err != nil {
        app.logError(r, err)
        w.WriteHeader(500)
    }
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
    app.logError(r, err)

    message := "the server encountered a problem and could not process your request"
    app.errorResponse(w, r, http.StatusInternalServerError, message)
}

func (app *application) notFoundErrorResponse(w http.ResponseWriter, r *http.Request) {
    message := "the request method could not be found"
    app.errorResponse(w, r, http.StatusNotFound, message)
}

func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
    message := fmt.Sprintf("The %s method is not allowed for this resource", r.Method)
    app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
    app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
    app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}
