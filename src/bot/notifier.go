package bot

import (
	"log"
	"time"
	"trbot/src/botDB"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// how long before the occurrence of the event
// we must send notification
const howLongBefore = time.Minute * 5

const utcOffset = time.Hour * 3

const idlingDuration = time.Hour * 24

// tgNotifier holds the notification logic
type tgNotifier struct {
	qm   dbQueryManager
	api  *tgbotapi.BotAPI
	recs []botDB.PurchaseRecord
	chat int64
	upd  <-chan struct{}
	done <-chan struct{}
}

func newTgNotifier(qm dbQueryManager, api *tgbotapi.BotAPI, chat int64, upd <-chan struct{}, done <-chan struct{}) *tgNotifier {
	return &tgNotifier{qm: qm, recs: nil, chat: chat, upd: upd, done: done}
}

// notify will send notification to specified telegram chat
// close to event time
func (n *tgNotifier) notify() {
	// set today's records
	if err := n.todays(); err != nil {
		log.Printf("Notifier\t->\terror due fetching records: [%s]\n", err.Error())
		return
	}

	// clean up
	defer func() {
		n.recs = nil
	}()

	// get remaining time to the closest
	// event and record index of that event
	i, d := n.nearestEventTime()

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
				log.Printf("Notifier\t->\terror due fetching records: [%s]\n", err.Error())
			}
			// update remaining time to next event
			// and record index of that event
			i, d = n.nearestEventTime()

		// if remaining time is expired, we notify
		case <-time.After(d):
			// in case we don't have any active records
			if i < 0 {
				continue
			}

			txt, _ := n.recs[i].Info()
			msg := tgbotapi.NewMessage(n.chat, txt)

			if _, err := n.api.Send(msg); err != nil {
				log.Printf("Notifier\t->\terror due sending notification: [%s]\n", err.Error())
			}

			// dequeue the record we notified about
			n.recs = n.recs[i:]
			// update remaining time to next event
			// and record index of that event
			i, d = n.nearestEventTime()
		}
	}

}

// nearestEventTime returns nearest remaining time
// to next event and also an inner slice index of nearest event record.
// If there are no records then -1 index will be returned
func (n *tgNotifier) nearestEventTime() (int, time.Duration) {

	now := time.Now().Add(utcOffset)

	for i := range n.recs {
		t := n.recs[i].BiddingDateTimeSql.Time.Add(utcOffset)
		if t.Add(-howLongBefore).After(now) {
			return i, now.Sub(t.Add(-howLongBefore))
		}
		if t.After(now) {
			return i, 0
		}
	}
	return -1, idlingDuration
}

// todays gets the today's records from the DB
func (n *tgNotifier) todays() error {
	var err error
	n.recs, err = n.qm.Query(0, botDB.TodayAuction)
	if err != nil {
		return err
	}
	return nil
}
