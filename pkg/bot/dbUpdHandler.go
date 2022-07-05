package bot

import (
	"net/http"
	"time"
)

// database update handler message
const (
	dbUpdateSuccess = "database was successfully updated"
	dbUpdateFailure = "unable to update database"
)

func (bot *Bot) dbUpdateHandler(updateTimeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// pass body to database handler
		err := bot.db.Upsert(r.Body)
		if err != nil {
			bot.logger.Printf(" | [DB Update Handler] -> [due updating records: err=%v]", err)
			writeResponse(w, dbUpdateFailure, http.StatusInternalServerError)
			return
		}

		go func() {
			select {
			// inform to update channel
			case bot.dbUpd <- struct{}{}:
				// or wait for a timeout and go off
			case <-time.After(updateTimeout):
				return
			}
		}()

		bot.logger.Printf("[DB Update Handler] -> [%s]", dbUpdateSuccess)
		writeResponse(w, dbUpdateSuccess, http.StatusOK)
	})
}
