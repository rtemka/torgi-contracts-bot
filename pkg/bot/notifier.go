package bot

import (
	"log"
	botDB "tbot/pkg/db"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// how long before the occurrence of the event
// we must send notification
const howLongBefore = time.Minute * 10

const utcOffset = time.Hour * 3

const idlingDuration = time.Hour * 24

// tgNotifier holds the notification logic
type tgNotifier struct {
	logger *log.Logger
	qm     dbQueryManager
	api    *tgbotapi.BotAPI
	recs   []botDB.PurchaseRecord
	chat   int64
	upd    <-chan struct{}
	done   <-chan struct{}
}

func newTgNotifier(logger *log.Logger, qm dbQueryManager,
	api *tgbotapi.BotAPI, chat int64, upd <-chan struct{},
	done <-chan struct{}) *tgNotifier {
	return &tgNotifier{logger: logger,
		qm:   qm,
		api:  api,
		recs: nil,
		chat: chat,
		upd:  upd,
		done: done,
	}
}

// notify will send notification to specified telegram chat
// close to event time
func (n *tgNotifier) notify() {
	// set today's records
	if err := n.todays(); err != nil {
		n.logger.Printf("error due fetching records: [%v]\n", err)
		return
	}

	// get remaining time to the closest
	// event and record index of that event
	i, d := n.nearestEventTime()
	n.logNearestEventTime(i, d)

	for {
		select {
		case <-n.done:
			return

		// if databse was updated we need to
		// update notifier records and
		// remaining time to next event
		case <-n.upd:
			// update records
			if err := n.todays(); err != nil {
				n.logger.Printf("error due fetching records: [%v]\n", err)
				return
			}
			n.logger.Println("got update")
			// update remaining time to next event
			// and record index of that event
			i, d = n.nearestEventTime()
			n.logNearestEventTime(i, d)

		// if remaining time is expired, we notify
		case <-time.After(d):
			// in case we don't have any active records
			if i < 0 {
				continue
			}

			if err := send(n.api, n.chat, buildMessages(n.recs[i])...); err != nil {
				n.logger.Println(err)
			}

			// dequeue the record we notified about
			if len(n.recs) > 1 {
				n.recs = n.recs[i+1:]
			} else {
				n.recs = nil // if it's last
			}
			// update remaining time to next event
			// and record index of that event
			i, d = n.nearestEventTime()
			n.logNearestEventTime(i, d)
		}
	}

}

// nearestEventTime returns nearest remaining time
// to next event and also an inner slice index of nearest event record.
// If there are no records then -1 index will be returned
func (n *tgNotifier) nearestEventTime() (int, time.Duration) {

	now := time.Now().Add(utcOffset)

	for i := range n.recs {
		t := n.recs[i].BiddingDateTimeSql.Time

		if t.Add(-howLongBefore).After(now) {
			return i, t.Add(-howLongBefore).Sub(now)
		}

		if t.After(now) {
			return i, 0
		}
	}
	return -1, idlingDuration
}

// todays gets the today's records from the DB.
// The db returns records in asc order
func (n *tgNotifier) todays() error {
	var err error
	n.recs, err = n.qm.Query(0, botDB.TodayAuction)
	if err != nil {
		return err
	}

	return nil
}

func (n *tgNotifier) logNearestEventTime(idx int, nt time.Duration) {
	if idx < 0 {
		n.logger.Printf("no nearest events; next check in [%s]\n", nt)
		return
	}

	n.logger.Printf("nearest event is [%s, %s, %s, %s]; remaining time to notification [%s]\n",
		n.recs[idx].RegistryNumber, n.recs[idx].PurchaseType,
		n.recs[idx].EtpSql.String, n.recs[idx].BiddingDateTimeSql.Time, nt)
}
