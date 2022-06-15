package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var mockErr = fmt.Errorf("intentional error")

type dbManagerMock struct{ needErr bool }

func (d dbManagerMock) Upsert(_ io.ReadCloser) error {
	if d.needErr {
		return mockErr
	}
	return nil
}

func (d dbManagerMock) Delete() error { return nil }

func TestDbUpdHandlerServeHTTP(t *testing.T) {

	// set up database update handler
	// with mocked database handler
	dbm := dbManagerMock{needErr: false}
	upd := make(chan struct{})
	defer close(upd)
	logger := log.New(io.Discard, "", 0)
	dbh := newDbHandler(logger, dbm, upd)

	// set up fake request
	req := httptest.NewRequest(http.MethodPost, "http://example.com/", nil)
	req.Header["Content-Type"] = []string{"application/json"}

	t.Run("good_request", func(t *testing.T) {

		w := httptest.NewRecorder()
		dbh.ServeHTTP(w, req)

		select {
		case <-upd:
		case <-time.After(time.Second * 10):
			t.Fatal("expected to receive update message, got nothing")
		}

		resp := w.Result()

		assertStatusCode(http.StatusOK, resp.StatusCode, t)
		assertMediaType(resp.Header.Get("Content-Type"), t)

		r := decodeResponse(resp.Body, t)

		if r["response"] != dbUpdateSuccess {
			t.Fatalf("expected to receive [%s], got [%s]", dbUpdateSuccess, r["response"])
		}
	})

	t.Run("bad_request_db", func(t *testing.T) {

		// this time we expect error
		dbh.dbManager = dbManagerMock{needErr: true}

		w := httptest.NewRecorder()
		dbh.ServeHTTP(w, req)

		resp := w.Result()

		assertStatusCode(http.StatusInternalServerError, resp.StatusCode, t)
		assertMediaType(resp.Header.Get("Content-Type"), t)

		r := decodeResponse(resp.Body, t)

		if r["response"] != dbUpdateFailure {
			t.Fatalf("expected to receive [%s], got [%s]", dbUpdateFailure, r["response"])
		}
	})

	t.Run("bad_request_media", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "http://example.com/", nil)
		// set up a wrong media type
		req.Header["Content-Type"] = []string{"bad/media"}
		w := httptest.NewRecorder()

		dbh.ServeHTTP(w, req)

		resp := w.Result()

		assertStatusCode(http.StatusUnsupportedMediaType, resp.StatusCode, t)
	})
}

func assertStatusCode(expected, got int, t *testing.T) {
	if expected != got {
		t.Fatalf("expected to receive status code [%d], got [%d]", expected, got)
	}
}

func assertMediaType(ct string, t *testing.T) {
	if ct != "application/json" {
		t.Fatalf("expected to receive application/json Content-Type, got [%v]", ct)
	}
}

func decodeResponse(body io.ReadCloser, t *testing.T) map[string]string {
	r := make(map[string]string)
	err := json.NewDecoder(body).Decode(&r)
	if err != nil {
		t.Fatalf("due reading response body [%v]", err)
	}
	return r
}
