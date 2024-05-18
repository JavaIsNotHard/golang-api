package main 

import (
    "fmt"
    "net/http"
)

// How a middleware works in go?
// basically we have a chain of handlers that is getting called one after another
// we wrap the router around another handler that calls the router's serverhttp method 

// panic recovery middleware which runs before the router servehttp such that any panic in the routes will exit the function and call the defer function which sends a error response to the client telling to close the connection
func (app *application) recoverPanic(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                w.Header().Set("Connection", "close")
                app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
            }
        }()
        next.ServeHTTP(w, r) // call the next handler in the chain 
    })
}
