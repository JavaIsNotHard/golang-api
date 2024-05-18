package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

    "api/internal/validator"

	"github.com/julienschmidt/httprouter"
)

type envelope map[string]any

func (app *application) getIDparam(r *http.Request) (int64, error) {
    params := httprouter.ParamsFromContext(r.Context())

    id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
    if err != nil || id < 1 {
        return 0, errors.New("Invalid id parameter")
    }

    return id, nil
}

func (app *application) WriteJSON(w http.ResponseWriter, data envelope, headers http.Header, status int) error {
    response, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        return err
    }

    response = append(response, '\n')

    for key, value := range headers {
        w.Header()[key] = value
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    w.Write(response)

    return nil
}

func (app *application) ReadJSON(w http.ResponseWriter, r *http.Request, dst any) error {
    // limit the request header size to be 1.5 MB
    maxBytes := 1_048_576
    r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

    dec := json.NewDecoder(r.Body)
    // disallow fields that do not have a correspondence to the struct tags 
    dec.DisallowUnknownFields()
    err := dec.Decode(dst)

    if err != nil {
        // syntax error in the body of the json request
        var syntaxError *json.SyntaxError
        // error if field types or struct tags do not match
        var unmarshallTypeError *json.UnmarshalTypeError
        // error if we didn't pass a pointer to the destination or some Decoding error 
        var invalidUnmarshalError * json.InvalidUnmarshalError

        switch {
        case errors.As(err, &syntaxError):
            return fmt.Errorf("body contains badly formed JSON (at character %d)", syntaxError.Offset)

        case errors.Is(err, io.ErrUnexpectedEOF):
            return fmt.Errorf("body contains badly formed JSON")

        case errors.As(err, &unmarshallTypeError):
            if unmarshallTypeError.Field != "" {
                return fmt.Errorf("body contains incorrect JSON type of the field %q", unmarshallTypeError.Field)
            }
            return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshallTypeError.Offset)

        case errors.Is(err, io.EOF):
            return errors.New("body must not be empty")

        case errors.As(err, &invalidUnmarshalError):
            panic(err)

        default:
           return err
        }
    }

    err = dec.Decode(&struct{}{})
    if !errors.Is(err, io.EOF) {
        return errors.New("body must only contain a single JSON value")
    }

    return nil    
}

func (app *application) readString(value url.Values, defaultValue string, key string) string {
    result := value.Get(key)

    if result == "" {
        return defaultValue
    }

    return result
}

func (app *application) readCommanSeparateValue(value url.Values, defaultValue []string, key string) []string {
    commaSeparateValue := value.Get(key)

    if commaSeparateValue == "" {
        return defaultValue
    }

    return strings.Split(commaSeparateValue, ",")
}

func (app *application) readInt(value url.Values, defaultValue int, key string, v *validator.Validator) int {
    result := value.Get(key)

    if result == "" {
        return defaultValue
    }

    i, err := strconv.Atoi(result)
    if err != nil {
        v.AddError(key, "must be an integer")
        return defaultValue
    }

    return i
}
