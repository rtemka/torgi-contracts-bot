package bot

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
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
	name      string
	dbManager dbManager
	upd       chan<- struct{}
}

// dbManager is responsible for the execution
// of CRUD operations over the database
type dbManager interface {
	Upsert(rc io.ReadCloser) error
	Delete() error
}

func newDbHandler(name string, m dbManager, upd chan<- struct{}) *dbHandler {
	return &dbHandler{name: name, dbManager: m, upd: upd}
}

func (d dbHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// check if request has json body
	headerContentType := r.Header.Get("Content-Type")
	if headerContentType != "application/json" {
		writeResponse(w, "Content Type is not application/json", http.StatusUnsupportedMediaType)
		return
	}

	// pass body to database handler
	err := d.dbManager.Upsert(r.Body)
	if err != nil {
		log.SetOutput(os.Stderr)
		log.Printf("%serror due updating records [%s]\n", d.name, err.Error())
		writeResponse(w, dbUpdateFailure, http.StatusInternalServerError)
		return
	}

	log.Printf("%s%s\n", d.name, dbUpdateSuccess)

	go func() {
		select {
		// inform to update channel
		case d.upd <- struct{}{}:
		// or wait for a timeout and go off
		case <-time.After(time.Second * 10):
			return
		}
	}()

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
