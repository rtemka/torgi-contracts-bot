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
	unknownMsg     = "Извини, не знаю такой команды. Попробуй -- /help"
	errorMsg       = "Извини 😥, не получилось выполнить команду"
	invalidArgsMsg = "Извини, для команды введены неправильные аргументы 🤷"
	hiMsg          = "Привет 👋 -- /help для справки"
	startMsg       = "Готов к работе ⚒️"
	statusMsg      = "👌 -- /help для справки"
)

// command help message
const (
	generalHelpMsg = "Доступные команды:\n\t" + todayKey + " - аукционы⚔️/заявки🔜 сегодня" +
		"\n\t" + futureKey + " - аукционы/заявки/обеспечения в будущем ⏳" +
		"\n\t" + pastKey + " - результаты прошедших закупок ⚰️" +
		"\n\t" + infoKey + " - информация по конкретной закупке 📝" +
		"\nПодробнее о каждой команде: /" + helpCmd + " -[имя команды]"
	todayHelpMsg = "Имя:\n\t" + todayKey + "\n\nИспользование:\n\t" + todayKey + "\t[опции]...\nОписание:" +
		"\n\t" + todayKey + " значит 'today' т.е 'сегодня'.\n\tПоказывает все ожидаемые сегодня торги и заявки, которые нужно подать" +
		"\nОпции:\n\t-" + auctionKey + ", " + auctionKeyLong + "\t " + auctionKeyUsg + "\n\t-g, --go\t" + goKeyUsg
	futureHelpMsg = "Имя:\n\t" + futureKey + "\nИспользование:\n\t" + futureKey + "\t[опции]...\nОписание:" +
		"\n\t" + futureKey + " значит 'future' т.е 'будущее'.\n\tПоказывает все будущие аукционы и заявки," +
		"\nа также агрегированные по региону и типу закупки суммы обеспечения\n" +
		"\nОпции:\n\t-" + auctionKey + ", " + auctionKeyLong + "\t " + auctionKeyUsg + "\n\t-g, --go\t" + goKeyUsg +
		"\n\t-" + moneyKey + ", --" + moneyKeyLong + "\t" + moneyKeyUsg +
		"\n\t-" + daysKey + ", --" + daysKeyLong + "=NUM\t" + daysKeyUsg + " вперед"
	pastHelpMsg = "Имя:\n\t" + pastKey + "\nИспользование:\n\t" + pastKey + "\t[опции]...\nОписание:" +
		"\n\t" + pastKey + " значит 'past' т.е 'прошлое'.\n\tПоказывает результаты прошедших закупок\n" +
		"\n\t-" + daysKey + ", --" + daysKeyLong + "=NUM\t" + daysKeyUsg + " назад"
	infoHelpMsg = "Имя:\n\t" + infoKey + "\nИспользование:\n\t" + infoKey + "\t=ID\nОписание:" +
		"\n\t" + infoKey + "Показывает информацию по конкретной закупке\n" +
		"\n\tВ выводе других команд есть значение в форме [id]." +
		"\n\tЭто значение нужно ввести как аргумент для этой команды т.е '" + infoKey + "id'"
	cmdHelp = "помощь по команде "
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
	todayKey       = "/" + todayCmd
	futureKey      = "/" + futureCmd
	pastKey        = "/" + pastCmd
	infoKey        = "/" + infoCmd
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
	qm dbQueryManager
}

func newTgUpdHandler(qm dbQueryManager) *tgUpdHandler {
	return &tgUpdHandler{qm: qm}
}

// handleUpdate redirects incoming update to appropriate handler
func (t *tgUpdHandler) handleUpdate(api *tgbotapi.BotAPI, u *tgbotapi.Update) {
	if !u.Message.IsCommand() {
		return
	}

	// we split incoming message command arguments
	args := strings.Split(u.Message.CommandArguments(), " ")
	// then we parse flags from this message as if it was
	// command line arguments
	flags, err := parseFlags(args)
	if err != nil {
		log.Println(err)
		api.Send(tgbotapi.NewMessage(u.Message.Chat.ID, errorMsg))
		return
	}

	var msgs []string

	// choosing appropriate handler
	switch u.Message.Command() {
	case todayCmd:
		msgs = t.todayCmdResponse(flags)
	case futureCmd:
		msgs = t.futureCmdResponse(flags)
	case pastCmd:
		msgs = t.pastCmdResponse(flags)
	case helpCmd:
		msgs = t.helpCmdResponse(flags)
	case infoCmd:
		msgs = t.infoCmdResponse(flags)
	case startCmd:
		msgs = []string{startMsg}
	case statusCmd:
		msgs = []string{statusMsg}
	case hiCmd:
		msgs = t.hiCmdResponse(u.Message)
	case chatCmd:
		msgs = []string{fmt.Sprint(u.Message.Chat.ID)}
	default:
		msgs = []string{unknownMsg}
	}

	// sending responses
	for i := range msgs {
		msg := tgbotapi.NewMessage(u.Message.Chat.ID, msgs[i])
		api.Send(msg)
	}
}

// flags holds flag set and all expected flags
type flags struct {
	set                    *flag.FlagSet
	tf, ff, pf, af, gf, mf bool
	df                     int
}

// parseFlags parses expected flags to the flags struct
func parseFlags(args []string) (*flags, error) {
	f := flags{}
	f.set = flag.NewFlagSet("bot flag set", flag.ContinueOnError)
	if len(args) == 0 {
		return &f, nil // if no arguments provided we don't parsing
	}

	f.set.BoolVar(&f.tf, todayKey, false, cmdHelp+todayKey)
	f.set.BoolVar(&f.ff, futureKey, false, cmdHelp+futureKey)
	f.set.BoolVar(&f.pf, pastKey, false, cmdHelp+pastKey)
	f.set.BoolVar(&f.af, auctionKey, false, auctionKeyUsg)
	f.set.BoolVar(&f.gf, goKey, false, goKeyUsg)
	f.set.BoolVar(&f.af, auctionKeyLong, false, auctionKeyUsg)
	f.set.BoolVar(&f.gf, goKeyLong, false, goKeyUsg)
	f.set.BoolVar(&f.mf, moneyKey, false, moneyKeyUsg)
	f.set.BoolVar(&f.mf, moneyKeyLong, false, moneyKeyUsg)
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
		msg = fmt.Sprintf("Привет, %s 👋\n--> /help для справки", m.From.FirstName)
	} else if m.From.UserName != "" {
		msg = fmt.Sprintf("Привет, %s 👋\n--> /help для справки", m.From.UserName)
	}
	return []string{msg}
}

// helpCmdResponse is the '/help' command handler
func (t *tgUpdHandler) helpCmdResponse(f *flags) []string {

	// at least one message will be returned
	msg := make([]string, 0, f.set.NFlag()+1)

	switch {
	case f.set.NArg() > 0:
		errMsg := fmt.Sprintf("переданы непонятные для меня аргументы --> %v", f.set.Args())
		msg = append(msg, errMsg)
	case f.tf:
		msg = append(msg, todayHelpMsg)
		fallthrough
	case f.ff:
		msg = append(msg, futureHelpMsg)
		fallthrough
	case f.pf:
		msg = append(msg, pastHelpMsg)
	default:
		msg = append(msg, generalHelpMsg)
	}

	return msg
}

// todayCmdResponse is the '/t' command handler
func (t *tgUpdHandler) todayCmdResponse(f *flags) []string {

	// at least one query option will be build
	opts := make([]botDB.QueryOpt, 0, f.set.NFlag()+1)

	switch {
	case f.set.NArg() > 0:
		errMsg := fmt.Sprintf("переданы непонятные для меня аргументы --> %v", f.set.Args())
		return []string{errMsg}
	case f.af:
		opts = append(opts, botDB.TodayAuction)
		fallthrough
	case f.gf:
		opts = append(opts, botDB.TodayGo)
	default:
		opts = append(opts, botDB.Today)
	}

	return t.query(0, opts...)
}

// futureCmdResponse is the '/f' command handler
func (t *tgUpdHandler) futureCmdResponse(f *flags) []string {

	// at least one query option will be build
	opts := make([]botDB.QueryOpt, 0, f.set.NFlag()+1)

	switch {
	case f.set.NArg() > 0:
		errMsg := fmt.Sprintf("переданы непонятные для меня аргументы --> %v", f.set.Args())
		return []string{errMsg}
	case f.af:
		opts = append(opts, botDB.FutureAuction)
		fallthrough
	case f.gf:
		opts = append(opts, botDB.FutureGo)
		fallthrough
	case f.mf:
		opts = append(opts, botDB.FutureMoney)
	default:
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
		log.Println(err)
		return []string{errorMsg}
	}

	p, err := t.qm.QueryRow(id)
	if err != nil {
		if err == botDB.ErrNoRows {
			return []string{"не нашел ничего по заданному id"}
		}
		log.Println(err)
		return []string{errorMsg}
	}

	return buildMessages(p)
}

// pastCmdResponse is the '/p' command handler
func (t *tgUpdHandler) pastCmdResponse(f *flags) []string {

	if f.set.NArg() > 0 {
		errMsg := fmt.Sprintf("переданы непонятные для меня аргументы --> %v", f.set.Args())
		return []string{errMsg}
	}

	return t.query(f.df, botDB.Past)
}

// query is the helper method that transmits
// options to database handler and then
// passes results to the message builder
func (t *tgUpdHandler) query(daysLimit int, opts ...botDB.QueryOpt) []string {

	recs, err := t.qm.Query(daysLimit, opts...) // gets results
	if err != nil {
		log.Println(err)
		return []string{errorMsg}
	}

	return buildMessages(recs...) // passes results
}

// buildMessages is the helper method that interacts with
// database record and builds messages for the response
func buildMessages(recs ...botDB.PurchaseRecord) []string {
	if len(recs) == 0 {
		return nil
	}

	var b strings.Builder
	var msgs []string
	var q botDB.QueryOpt

	for i := range recs {

		// gets info string from the record
		// and also a query option
		s, qr := recs[i].Info()
		// query option helps us to create
		// messages separated by type

		// if we encounter new query option
		// than the current message is complete
		if q != qr && i != 0 {
			msgs = append(msgs, b.String())

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
	msgs = append(msgs, b.String())

	return msgs
}
