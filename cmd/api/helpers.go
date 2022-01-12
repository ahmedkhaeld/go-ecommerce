package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// readJSON a helper that allows to read generic json, provide clean way to read any kind of json from a request, assuming
// that request body has only a single json value
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	// maxBytes define the max size of request allowed prevents clients from accidentally or maliciously sending a large
	//request and wasting server resources.
	maxBytes := 1048756
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	// assume to decode a json file that has one entry
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only have a single JSON value")
	}

	return nil
}

// writeJSON  writes arbitrary data out as json
//takes w, so it has something to write to, status which set the status for the w, data whatever type we want
// to turn into json, headers to be sent
func (app *application) writeJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	// check if headers has been set, then add those headers
	if len(headers) > 0 {
		for k, v := range headers[0] {
			w.Header()[k] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(out)

	return nil

}

func (app *application) badRequest(w http.ResponseWriter, r *http.Request, err error) error {
	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	payload.Error = true
	payload.Message = err.Error()

	out, err := json.MarshalIndent(payload, "", "\t")
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
	return nil
}
