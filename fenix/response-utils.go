package fenix

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
)

func (f *Fenix) ReadJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := 1048576 // one mb
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only have a single json value")
	}

	return nil
}

// Write JSON file
func (f *Fenix) WriteJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for k, v := range headers[0] {
			w.Header()[k] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}
	return nil
}

// Write XML file
func (f *Fenix) WriteXML(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := xml.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for k, v := range headers[0] {
			w.Header()[k] = v
		}
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}
	return nil
}

// Download file
func (f *Fenix) DownloadFile(w http.ResponseWriter, r *http.Request, pathToFile, fileName string) error {
	fp := path.Join(pathToFile, fileName)
	fileToServe := filepath.Clean(fp)
	// w.Header().Set("Content-Type", fmt.Sprintf("attachment; file=\"%s\"", fileName))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; file=\"%s\"", fileName))
	http.ServeFile(w, r, fileToServe)
	return nil
}

// Status 400 - Bad Request: The server cannot process the request due to a client error,
// such as invalid syntax or missing parameters.
func (f *Fenix) ErrorBadRequest(w http.ResponseWriter, r *http.Request) {
	f.ErrorStatus(w, http.StatusBadRequest)
}

// Status 404 - Not Found: The server could not find the requested resource.
func (f *Fenix) ErrorNotFound(w http.ResponseWriter, r *http.Request) {
	f.ErrorStatus(w, http.StatusNotFound)
}

// Status 500 -  Internal Server Error: The server encountered an unexpected condition
// that prevented it from fulfilling the request.
func (f *Fenix) ErrorIntServerErr(w http.ResponseWriter, r *http.Request) {
	f.ErrorStatus(w, http.StatusInternalServerError)
}

// Status 401 - Unauthorized: The request requires authentication,
// and the client does not provide valid credentials.
func (f *Fenix) ErrorUnauthorized(w http.ResponseWriter, r *http.Request) {
	f.ErrorStatus(w, http.StatusUnauthorized)
}

// Status 403 - Forbidden: The server understands the request,
// but the client is not allowed to access the requested resource.
func (f *Fenix) ErrorForbidden(w http.ResponseWriter, r *http.Request) {
	f.ErrorStatus(w, http.StatusForbidden)
}

// Status 405 - Method Not Allowed: The request method is not supported for the requested resource.
func (f *Fenix) ErrorMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	f.ErrorStatus(w, http.StatusMethodNotAllowed)
}

// Status 503 - Service Unavailable: The server is currently unavailable, often due to maintenance or overload.
func (f *Fenix) ErrorServiceUnavailable(w http.ResponseWriter, r *http.Request) {
	f.ErrorStatus(w, http.StatusServiceUnavailable)
}

// Error status helper function
func (f *Fenix) ErrorStatus(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
