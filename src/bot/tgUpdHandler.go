package bot

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"trbot/src/botDB"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// bot message
const (
	unknownMsg     = `Извини, не знаю такой команды\. Попробуй ➡️ */help*`
	errorMsg       = "Извини 😥, не получилось выполнить команду"
	invalidArgsMsg = "Извини, для команды введены неправильные аргументы 🤷\n" +
		`➡️ */help* \-\[*_имя команды_*\]`
	hiMsg          = "Привет 👋 ➡️ */help* для справки"
	startMsg       = "Готов к работе ⚒️"
	statusMsg      = "Все ок\\!"
	errorOptionMsg = "Неправильная опция команды\n" + `➡️ */help* \-\[*_имя команды_*\]` +
		"\nдля справки по команде"
	notFoundIdMsg = "Не нашел ничего по заданному id"
)

// command help message
const (
	generalHelpMsg = `*Доступные команды:*` + "\n\n" +
		`*/` + todayCmd + `* \- аукционы⚔️ / заявки🔜 сегодня` + "\n\n" +
		`*/` + futureCmd + `* \- аукционы/заявки/обеспечения в будущем 🔮` + "\n\n" +
		`*/` + pastCmd + `* \- результаты закупок ⚰️` + "\n\n" +
		`*/` + infoCmd + `* \- информация по закупке 📝` + "\n\n" +
		"Подробнее о каждой команде:" + "\n" + `*/` + helpCmd + `* \-\[*_имя команды_*\]`
	todayHelpMsg = `*Имя команды:       /` + todayCmd + "\n" + `Использование:   /` + todayCmd + `*    \[*_опции_*\]\.\.\.` +
		"\n" + `*Описание:*` + "\n" + `*/` + todayCmd + `* значит '*_today_*' т\.е '*_сегодня_*'` +
		"\nПоказывает все ожидаемые сегодня торги и заявки, которые нужно подать\n" +
		`*Опции:*` + "\n" + `*_\-` + auctionKey + `, \-` + auctionKeyLong + `_*    ` + auctionKeyUsg + "\n" +
		`*_\-` + goKey + `, \-` + goKeyLong + `_*           ` + goKeyUsg
	futureHelpMsg = `*Имя команды:       /` + futureCmd + "\n" + `Использование:   /` + futureCmd + `*    \[*_опции_*\]\.\.\. *_\=NUM_*` +
		"\n" + `*Описание:*` + "\n" + `*/` + futureCmd + `* значит '*_future_*' т\.е '*_будущее_*'` +
		"\nПоказывает все будущие аукционы и заявки, а также суммы обеспечения заявок\n" +
		`*Опции:*` + "\n" + `*_\-` + auctionKey + `, \-` + auctionKeyLong + `_*    ` + auctionKeyUsg + "\n" +
		`*_\-` + goKey + `, \-` + goKeyLong + `_*           ` + goKeyUsg + "\n" +
		`*_\-` + moneyKey + `, \-` + moneyKeyLong + `_*       ` + moneyKeyUsg + "\n" +
		`*_\-` + daysKey + `, \-` + daysKeyLong + `\=NUM_* ` + daysKeyUsg + " вперед"
	pastHelpMsg = `*Имя команды:       /` + pastCmd + "\n" + `Использование:   /` + pastCmd + `*    \[*_опции_*\]\.\.\. *_\=NUM_*` +
		"\n" + `*Описание:*` + "\n" +
		`*/` + pastCmd + `* значит '*_past_*' т\.е '*_прошлое_*'` + "\nПоказывает результаты прошедших закупок\n" +
		`*Опции:*` + "\n" + `*_\-` + daysKey + `, \-` + daysKeyLong + `\=NUM_* ` + daysKeyUsg + " назад"
	infoHelpMsg = `*Имя команды:      /` + infoCmd + "\n" + `Использование:   /` + infoCmd + `    \=ID*` +
		"\n" + `*Описание:*` + "\n" + `*/` + infoCmd + `* Показывает информацию по конкретной закупке` + "\n" +
		`В выводе других команд есть значение в форме \[*_ID_*\]\.` + "\n" +
		`Это значение нужно ввести как аргумент для этой команды т\.е '*/` + infoCmd + `  _ID_'*`
	cmdHelp = "помощь по команде /"
)

// bot command
const (
	todayCmd  = "t"
	futureCmd = "f"
	pastCmd   = "p"
	infoCmd   = "i"
	helpCmd   = "help"
	statusCmd = "status"
	startCmd  = "start"
	hiCmd     = "hi"
	chatCmd   = "chat"
)

// bot command key
const (
	auctionKey     = "a"
	auctionKeyLong = "auction"
	goKey          = "g"
	goKeyLong      = "go"
	moneyKey       = "m"
	moneyKeyLong   = "money"
	daysKey        = "d"
	daysKeyLong    = "days"
)

// key usage
const (
	auctionKeyUsg = "показывает аукционы"
	goKeyUsg      = "показывает заявки"
	moneyKeyUsg   = "показывает суммы обеспечения"
	daysKeyUsg    = "ограничивает выборку на NUM дней"
)

// dbQueryManager is responsible
// for the retrieving info from database
type dbQueryManager interface {
	Query(int, ...botDB.QueryOpt) ([]botDB.PurchaseRecord, error)
	QueryRow(int64) (botDB.PurchaseRecord, error)
}

// tgUpdHandler processes incoming telegram updates
type tgUpdHandler struct {
	logger *log.Logger
	api    *tgbotapi.BotAPI
	qm     dbQueryManager
}

func newTgUpdHandler(logger *log.Logger, qm dbQueryManager, api *tgbotapi.BotAPI) *tgUpdHandler {
	return &tgUpdHandler{logger: logger, qm: qm, api: api}
}

// handleUpdate redirects incoming update to appropriate handler
func (t *tgUpdHandler) handleUpdate(u *tgbotapi.Update) {
	if !u.Message.IsCommand() {
		return
	}

	// we parse flags from this message as if it was
	// command line arguments
	flags, err := parseMsgArgs(u.Message.CommandArguments())
	if err != nil {
		t.logger.Printf("error due parsing message arguments [%v]\n", err)
		send(t.api, u.Message.Chat.ID, errorOptionMsg)
		return
	}

	// get responses from command handlers
	msgs := t.responses(u, flags)

	// sending responses
	if err = send(t.api, u.Message.Chat.ID, msgs...); err != nil {
		t.logger.Println(err)
	}
}

func (t *tgUpdHandler) responses(u *tgbotapi.Update, flags *flags) []string {
	// choosing appropriate handler
	switch u.Message.Command() {
	case todayCmd:
		return t.todayCmdResponse(flags)
	case futureCmd:
		return t.futureCmdResponse(flags)
	case pastCmd:
		return t.pastCmdResponse(flags)
	case helpCmd:
		return t.helpCmdResponse(flags)
	case infoCmd:
		return t.infoCmdResponse(flags)
	case startCmd:
		return []string{startMsg}
	case statusCmd:
		return []string{statusMsg}
	case hiCmd:
		return t.hiCmdResponse(u.Message)
	case chatCmd:
		return []string{fmt.Sprint(u.Message.Chat.ID)}
	default:
		return []string{unknownMsg}
	}
}

// parseMsgArgs inspects provided arguments
// and returns parsed flags or error
func parseMsgArgs(args string) (*flags, error) {
	var s []string
	if args != "" {
		// we split incoming message command arguments
		s = strings.Split(args, " ")
	}
	// then we parse flags from this message as if it was
	// command line arguments
	// if args is empty we pass a nil slice
	flags, err := parseFlags(s)
	if err != nil {
		return nil, err
	}
	return flags, nil
}

// flags holds flag set and all expected flags
type flags struct {
	set                         *flag.FlagSet
	tf, ff, pf, af, gf, mf, inf bool
	df                          int
}

// parseFlags parses expected flags to the flags struct
func parseFlags(args []string) (*flags, error) {
	f := flags{}
	f.set = flag.NewFlagSet("bot flag set", flag.ContinueOnError)
	if len(args) == 0 {
		return &f, nil // if no arguments provided we don't parsing
	}

	f.set.BoolVar(&f.tf, todayCmd, false, cmdHelp+todayCmd)
	f.set.BoolVar(&f.ff, futureCmd, false, cmdHelp+futureCmd)
	f.set.BoolVar(&f.pf, pastCmd, false, cmdHelp+pastCmd)
	f.set.BoolVar(&f.af, auctionKey, false, auctionKeyUsg)
	f.set.BoolVar(&f.gf, goKey, false, goKeyUsg)
	f.set.BoolVar(&f.af, auctionKeyLong, false, auctionKeyUsg)
	f.set.BoolVar(&f.gf, goKeyLong, false, goKeyUsg)
	f.set.BoolVar(&f.mf, moneyKey, false, moneyKeyUsg)
	f.set.BoolVar(&f.mf, moneyKeyLong, false, moneyKeyUsg)
	f.set.BoolVar(&f.inf, infoCmd, false, cmdHelp+infoCmd)
	f.set.IntVar(&f.df, daysKey, 0, daysKeyUsg)
	f.set.IntVar(&f.df, daysKeyLong, 0, daysKeyUsg)
	err := f.set.Parse(args)
	if err != nil {
		return &f, err
	}

	return &f, nil
}

// hiCmdResponse is the '/hi' command handler
func (t *tgUpdHandler) hiCmdResponse(m *tgbotapi.Message) []string {
	msg := hiMsg
	if m.From.FirstName != "" {
		msg = fmt.Sprintf("Привет, %s 👋\n➡️ */help* для справки", m.From.FirstName)
	} else if m.From.UserName != "" {
		msg = fmt.Sprintf("Привет, %s 👋\n➡️ */help* для справки", m.From.UserName)
	}
	return []string{msg}
}

// unknownArgsErr returns error message when
// input arguments contains some garbage leftovers
func unknownArgsErr(f *flags) []string {
	return []string{fmt.Sprintf("Переданы непонятные для меня аргументы ➡️ %v", f.set.Args())}
}

// helpCmdResponse is the '/help' command handler
func (t *tgUpdHandler) helpCmdResponse(f *flags) []string {

	if f.set.NFlag() == 0 {
		return []string{generalHelpMsg}
	}

	msg := make([]string, 0, f.set.NFlag())

	if f.tf {
		msg = append(msg, todayHelpMsg)
	}
	if f.ff {
		msg = append(msg, futureHelpMsg)
	}
	if f.pf {
		msg = append(msg, pastHelpMsg)
	}
	if f.inf {
		msg = append(msg, infoHelpMsg)
	}

	return msg
}

// todayCmdResponse is the '/t' command handler
func (t *tgUpdHandler) todayCmdResponse(f *flags) []string {

	// check for the garbage in arguments
	if f.set.NArg() > 0 {
		return unknownArgsErr(f)
	}

	if f.set.NFlag() == 0 {
		return t.query(0, botDB.Today)
	}

	opts := make([]botDB.QueryOpt, 0, f.set.NFlag())

	if f.af {
		opts = append(opts, botDB.TodayAuction)
	}
	if f.gf {
		opts = append(opts, botDB.TodayGo)
	}

	return t.query(0, opts...)
}

// futureCmdResponse is the '/f' command handler
func (t *tgUpdHandler) futureCmdResponse(f *flags) []string {

	// check for the garbage in arguments
	if f.set.NArg() > 0 {
		return unknownArgsErr(f)
	}

	opts := make([]botDB.QueryOpt, 0, f.set.NFlag())

	if f.af {
		opts = append(opts, botDB.FutureAuction)
	}
	if f.gf {
		opts = append(opts, botDB.FutureGo)
	}
	if f.mf {
		opts = append(opts, botDB.FutureMoney)
	}

	if len(opts) == 0 {
		opts = append(opts, botDB.Future)
	}

	return t.query(f.df, opts...)
}

// infoCmdResponse is the '/i' command handler
func (t *tgUpdHandler) infoCmdResponse(f *flags) []string {

	// we expecting only one argument which is id
	if f.set.NArg() != 1 {
		return []string{invalidArgsMsg}
	}

	id, err := strconv.ParseInt(f.set.Arg(0), 10, 0)
	if err != nil {
		t.logger.Printf("error due converting id [%v]\n", err)
		return []string{errorMsg}
	}

	p, err := t.qm.QueryRow(id)
	if err != nil {
		if err == botDB.ErrNoRows {
			return []string{notFoundIdMsg}
		}
		t.logger.Printf("error due fetching record [%v]\n", err)
		return []string{errorMsg}
	}

	return buildMessages(p)
}

// pastCmdResponse is the '/p' command handler
func (t *tgUpdHandler) pastCmdResponse(f *flags) []string {
	// check for the garbage in arguments
	if f.set.NArg() > 0 {
		return unknownArgsErr(f)
	}

	return t.query(f.df, botDB.Past)
}

// query is the helper method that transmits
// options to database handler and then
// passes results to the message builder
func (t *tgUpdHandler) query(daysLimit int, opts ...botDB.QueryOpt) []string {

	recs, err := t.qm.Query(daysLimit, opts...) // gets results
	if err != nil {
		t.logger.Printf("error due fetching records [%v]\n", err)
		return []string{errorMsg}
	}

	return buildMessages(recs...) // passes results
}
