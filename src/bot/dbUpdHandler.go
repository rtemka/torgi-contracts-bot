package bot

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// Handler messages
const (
	dbUpdateSuccess = "database was successfully updated"
	dbUpdateFailure = "unable to update database"
)

// dbHandler holds database manager logic.
// It takes incoming requests with updates, validates
// that it's json and pass requests body to database
// manager
type dbHandler struct {
	dbManager dbManager
}

// dbManager is responsible for the execution
// of CRUD operations over the database
type dbManager interface {
	Upsert(rc io.ReadCloser) error
	Delete() error
}

func newDbHandler(m dbManager) *dbHandler {
	return &dbHandler{dbManager: m}
}

func (d dbHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	headerContentType := r.Header.Get("Content-Type")
	if headerContentType != "application/json" {
		writeResponse(w, "Content Type is not application/json", http.StatusUnsupportedMediaType)
		return
	}

	err := d.dbManager.Upsert(r.Body)
	if err != nil {
		log.Println(err)
		writeResponse(w, dbUpdateFailure, http.StatusInternalServerError)
		return
	}

	writeResponse(w, dbUpdateSuccess, http.StatusOK)
}

func writeResponse(w http.ResponseWriter, message string, httpStatusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	resp := make(map[string]string)
	resp["response"] = message
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)
}
