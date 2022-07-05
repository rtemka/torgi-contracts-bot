package bot

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"tbot/pkg/bot/memdb"
	"testing"
	"time"
)

func TestBot_dbUpdateHandler(t *testing.T) {

	// set up database update handler
	// with mocked database handler
	dbm := memdb.New(false)
	upd := make(chan struct{})
	defer close(upd)
	logger := log.New(io.Discard, "", 0)
	tb := Bot{
		db:     dbm,
		logger: logger,
		dbUpd:  upd,
	}

	timeout := 1000 * time.Millisecond
	h := tb.headersMiddleware(tb.enforceJsonMiddleware(tb.dbUpdateHandler(timeout)))
	req := httptest.NewRequest(http.MethodPost, "http://test.com/", nil)
	req.Header["Content-Type"] = []string{"application/json"}

	t.Run("good_request", func(t *testing.T) {

		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		select {
		case <-upd:
		case <-time.After(timeout):
			t.Fatal("Bot.dbUpdateHandler() expected to receive update msg, got nothing")
		}

		resp := w.Result()

		assert("Bot.dbUpdateHandler()", resp.StatusCode, http.StatusOK, t)
		assert("Bot.dbUpdateHandler()", resp.Header.Get("Content-Type"), "application/json", t)

		r := decodeResponse("Bot.dbUpdateHandler()", resp.Body, t)

		assert("Bot.dbUpdateHandler()", r["response"], dbUpdateSuccess, t)
	})

	t.Run("bad_request_db", func(t *testing.T) {

		// this time we expect error
		tb.db = memdb.New(true)

		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		resp := w.Result()

		assert("Bot.dbUpdateHandler()", resp.StatusCode, http.StatusInternalServerError, t)
		assert("Bot.dbUpdateHandler()", resp.Header.Get("Content-Type"), "application/json", t)

		r := decodeResponse("Bot.dbUpdateHandler()", resp.Body, t)

		assert("Bot.dbUpdateHandler()", r["response"], dbUpdateFailure, t)
	})

	t.Run("bad_request_media", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodPost, "http://example.com/", nil)
		// set up a wrong media type
		req.Header["Content-Type"] = []string{"bad/mediatype"}
		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		resp := w.Result()
		assert("Bot.dbUpdateHandler()", resp.StatusCode, http.StatusUnsupportedMediaType, t)
	})
}

func assert[T comparable](op string, got T, want T, t *testing.T) {
	if got != want {
		t.Fatalf("%s got=%v, want=%v", op, got, want)
	}
}

func decodeResponse(op string, body io.ReadCloser, t *testing.T) map[string]string {
	r := make(map[string]string)
	err := json.NewDecoder(body).Decode(&r)
	if err != nil {
		t.Fatalf("%s error=%v", op, err)
	}
	return r
}
