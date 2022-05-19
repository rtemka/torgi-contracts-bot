package bot

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
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
	logger    *log.Logger
	dbManager dbManager
	upd       chan<- struct{}
}

// dbManager is responsible for the execution
// of CRUD operations over the database
type dbManager interface {
	Upsert(rc io.ReadCloser) error
	Delete() error
}

func newDbHandler(logger *log.Logger, m dbManager, upd chan<- struct{}) *dbHandler {
	return &dbHandler{logger: logger, dbManager: m, upd: upd}
}

func (d dbHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
	}()

	// check if request has json body
	headerContentType := r.Header.Get("Content-Type")
	if headerContentType != "application/json" {
		if err := writeResponse(w, "Content-Type is not application/json", http.StatusUnsupportedMediaType); err != nil {
			d.logger.Printf("error due writing response [%v]\n", err)
		}
		return
	}

	// pass body to database handler
	err := d.dbManager.Upsert(r.Body)
	if err != nil {
		d.logger.Printf("error due updating records [%v]\n", err)
		if err := writeResponse(w, dbUpdateFailure, http.StatusInternalServerError); err != nil {
			d.logger.Printf("error due writing response [%v]\n", err)
		}
		return
	}

	d.logger.Println(dbUpdateSuccess)

	go func() {
		select {
		// inform to update channel
		case d.upd <- struct{}{}:
		// or wait for a timeout and go off
		case <-time.After(time.Second * 10):
			return
		}
	}()

	if err = writeResponse(w, dbUpdateSuccess, http.StatusOK); err != nil {
		d.logger.Printf("error due writing response [%v]\n", err)
	}
}

func writeResponse(w http.ResponseWriter, message string, httpStatusCode int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	resp := make(map[string]string)
	resp["response"] = message
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	_, err = w.Write(jsonResp)
	return err
}
