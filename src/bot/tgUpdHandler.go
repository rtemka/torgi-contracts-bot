package bot

import (
	"bufio"
	"strconv"
	"strings"
	"trbot/src/botDB"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// bot message
const (
	unknownMsg     = "Извини, не знаю такой команды. Попробуй -- /help"
	errorMsg       = "Извини 😥, не получилось выполнить команду"
	invalidArgsMsg = "Извини, для команды введены неправильные аргументы 🤷"
	hiMsg          = "Привет 👋 -- /help для справки"
	startMsg       = "Готов к работе ⚒️"
	statusMsg      = "👌 -- /help для справки"
)

// command help message
const (
	generalHelpMsg = "Доступные команды:\n\n\t/t - аукционы⚔️/заявки🖥️ сегодня" +
		"\n\t/f - аукционы/заявки/обеспечения в будущем ⏳" +
		"\n\t/p - результаты прошедших закупок ⚰️" +
		"\n\nПодробнее о каждой команде: /help -[имя команды]"
	todayHelpMsg = "Имя:\n\n\t/t\n\nИспользование:\n\n\t/t\t[опции]...\n\nОписание:" +
		"\n\n\t/t значит 'today' т.е 'сегодня'.\n\tПоказывает все ожидаемые сегодня торги и заявки, которые нужно подать" +
		"\n\nОпции:\n\n\t-a, --auction\t показывает только аукционы\n\t-g, --go\t показывает только заявки"
	futureHelpMsg = "Имя:\n\t/f\nИспользование:\n\t/f\t[опции]...\nОписание:" +
		"\n\t/f значит 'future' т.е 'будущее'.\n\tПоказывает все будущие аукционы и заявки," +
		"\nа также агрегированные по региону и типу закупки суммы обеспечения\n" +
		"Опции:\n\t-a, --auction\t показывает только аукционы\n\t-g, --go\t показывает только заявки" +
		"\n\t-m, --money\t показывает только суммы обеспечения" +
		"\n\t-d, --days=[+]NUM\t ограничивает выборку на NUM дней вперед"
	pastHelpMsg = "Имя:\n\t/p\nИспользование:\n\t/p\t[опции]...\nОписание:" +
		"\n\t/p значит 'past' т.е 'прошлое'.\n\tПоказывает результаты прошедших закупок\n" +
		"Опции:\n\t-d, --days=[+]NUM\t ограничивает выборку на NUM дней назад"
)

// bot command
const (
	todayCmd  = "t"
	futureCmd = "f"
	pastCmd   = "p"
	helpCmd   = "help"
	statusCmd = "status"
	startCmd  = "start"
	hiCmd     = "hi"
)

// bot command argument
const (
	auctionArg     = "-a"
	auctionArgLong = "--auction"
	goingArg       = "-g"
	goingArgLong   = "--go"
	moneyArg       = "-m"
	moneyArgLong   = "--money"
	daysArg        = "-d"
	daysArgLong    = "--days"
	todayArg       = "-/t"
	futureArg      = "-/f"
	pastArg        = "-/p"
)

// dbQueryManager is responsible
// for the retrieving info from database
type dbQueryManager interface {
	Query(botDB.QueryOpt, int) ([]botDB.PurchaseRecord, error)
}

// tgUpdHandler processes incoming telegram updates
type tgUpdHandler struct {
	qm dbQueryManager
}

func newTgUpdHandler(qm dbQueryManager) *tgUpdHandler {
	return &tgUpdHandler{qm: qm}
}

func (t *tgUpdHandler) handleUpdate(api *tgbotapi.BotAPI, u *tgbotapi.Update) {
	if !u.Message.IsCommand() {
		return
	}

	var msgs []string

	switch u.Message.Command() {
	case todayCmd:
		msgs = t.todayCmdResponse(u.Message)
	case futureCmd:
		msgs = t.futureCmdResponse(u.Message)
	case pastCmd:
		msgs = t.pastCmdResponse(u.Message)
	case helpCmd:
		msgs = t.helpCmdResponse(u.Message)
	case startCmd:
		msgs = []string{startMsg}
	case statusCmd:
		msgs = []string{statusMsg}
	case hiCmd:
		msgs = []string{hiMsg}
	default:
		msgs = []string{unknownMsg}
	}

	for i := range msgs {
		msg := tgbotapi.NewMessage(u.Message.Chat.ID, msgs[i])
		api.Send(msg)
	}
}

func (t *tgUpdHandler) helpCmdResponse(m *tgbotapi.Message) []string {

	msg := generalHelpMsg

	args := m.CommandArguments()
	if len(args) == 0 {
		return []string{msg}
	}

	switch args {
	case todayArg:
		msg = todayHelpMsg
	case futureArg:
		msg = futureHelpMsg
	case pastHelpMsg:
		msg = pastHelpMsg
	default:
		msg = invalidArgsMsg
	}

	return []string{msg}
}

func (t *tgUpdHandler) todayCmdResponse(m *tgbotapi.Message) []string {

	args := m.CommandArguments()
	opt := botDB.Today

	switch args {
	case auctionArg, auctionArgLong:
		opt = botDB.TodayAuction
	case goingArg, goingArgLong:
		opt = botDB.TodayGo
	default:
		return []string{errorMsg}
	}

	recs, err := t.qm.Query(opt, 0)
	if err != nil {
		return []string{errorMsg}
	}

	return []string{buildMessage(recs, opt)}
}

func (t *tgUpdHandler) futureCmdResponse(m *tgbotapi.Message) []string {
	var (
		d   int
		err error
	)
	args := m.CommandArguments()
	opt := botDB.Future
	days := strings.Contains(args, daysArg) || strings.Contains(args, daysArgLong)

	if days {
		m := map[string]bool{daysArg: true, daysArgLong: true}
		args, d, err = parseAndStripOptions(args, m)
		if err != nil {
			return []string{errorMsg}
		}
	}

	switch args {
	case auctionArg, auctionArgLong:
		opt = botDB.FutureAuction
	case goingArg, goingArgLong:
		opt = botDB.FutureGo
	case moneyArg, moneyArgLong:
		opt = botDB.FutureMoney
	default:
		return []string{errorMsg}
	}

	recs, err := t.qm.Query(opt, d)
	if err != nil {
		return []string{errorMsg}
	}

	return []string{buildMessage(recs, opt)}
}

func (t *tgUpdHandler) pastCmdResponse(m *tgbotapi.Message) []string {
	var (
		d   int
		err error
	)
	args := m.CommandArguments()
	opt := botDB.Past
	days := strings.Contains(args, daysArg) || strings.Contains(args, daysArgLong)

	if days {
		m := map[string]bool{daysArg: true, daysArgLong: true}
		args, d, err = parseAndStripOptions(args, m)
		if err != nil {
			return []string{errorMsg}
		}
	}

	if args != "" {
		return []string{errorMsg}
	}

	recs, err := t.qm.Query(opt, d)
	if err != nil {
		return []string{errorMsg}
	}

	return []string{buildMessage(recs, opt)}
}

func buildMessage(recs []botDB.PurchaseRecord, opt botDB.QueryOpt) string {
	var b strings.Builder

	for i := range recs {
		b.WriteString(recs[i].InfoString(opt))
	}

	return b.String()
}
func parseAndStripOptions(args string, m map[string]bool) (string, int, error) {
	var (
		d   int
		err error
		b   strings.Builder
	)
	scanner := bufio.NewScanner(strings.NewReader(args))
	scanner.Split(bufio.ScanWords)

	for next, t := false, ""; scanner.Scan(); next = m[t] {
		t = scanner.Text()
		if next {
			d, err = strconv.Atoi(t)
			if err != nil {
				return b.String(), d, err
			}
			continue
		}
		if m[t] {
			continue
		}
		b.WriteString(t)
	}

	return b.String(), d, nil
}
