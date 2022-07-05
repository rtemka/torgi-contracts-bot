package bot

import (
	"fmt"
	"strings"
	botDB "tbot/pkg/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// telegram message formatting mode
const parseMode = "MarkdownV2"

// bot message
const notFoundMsg = "Похоже, что ничего нет\\.\\.\\. 🙃"

// send is helper function that is responsible
// for sending responses to the telegram chat
func send(api *tgbotapi.BotAPI, chatID int64, msgs ...string) error {
	m := tgbotapi.NewMessage(chatID, "")
	m.ParseMode = parseMode
	for i := range msgs {
		m.Text = msgs[i]
		if _, err := api.Send(m); err != nil {
			return fmt.Errorf("[Telegram] -> [due sending response: chat=%d; msg=%v; err=%v]",
				chatID, msgs[i], err)
		}
	}
	return nil
}

// buildMessages is the helper function that interacts with
// database record and builds messages for the response
func buildMessages(recs ...botDB.PurchaseRecord) []string {
	if len(recs) == 0 {
		return []string{notFoundMsg}
	}

	var b strings.Builder
	var msgs []string
	var q botDB.QueryOpt
	// we need replacer to sanitize messages
	// for telegram markdown syntax
	r := strings.NewReplacer(
		"[", "\\[", "]", "\\]", "(", "\\(", ")",
		"\\)", "~", "\\~", "`", "\\`", ">", "\\>", "#",
		"\\#", "+", "\\+", "-", "\\-", "=", "\\=", "|",
		"\\|", "{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!")

	for i := range recs {

		// gets info string from the record
		// and also a query option
		s, qr := recs[i].Info()
		// query option helps us to create
		// messages separated by type

		// if we encounter new query option
		// then the current message is complete
		if q != qr && i != 0 {
			msgs = append(msgs, r.Replace(b.String()))

			// reseting builder and writing
			// new message header
			b.Reset()
			b.WriteString(qr.String())
		} else if q != qr {
			// if its first record
			// we only write header
			b.WriteString(qr.String())
		}

		b.WriteString(s)
		q = qr
	}

	// appending the last message
	msgs = append(msgs, r.Replace(b.String()))

	return msgs
}
