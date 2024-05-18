package main 

import (
    "net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
    response := envelope{
        "status": "available",
        "environment": app.config.env,
        "version": version,
    }

    err := app.WriteJSON(w, response, nil, http.StatusOK)
    if err != nil {
        app.serverErrorResponse(w, r, err)
    }
}
