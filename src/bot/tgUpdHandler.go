package bot

import (
	"flag"
	"fmt"
	"log"
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
	generalHelpMsg = "Доступные команды:\n\t" + todayKey + " - аукционы⚔️/заявки🖥️ сегодня" +
		"\n\t" + futureKey + " - аукционы/заявки/обеспечения в будущем ⏳" +
		"\n\t" + pastKey + " - результаты прошедших закупок ⚰️" +
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
	cmdHelp = "помощь по команде "
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
	todayKey       = "/t"
	futureKey      = "/f"
	pastKey        = "/p"
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

	args := strings.Split(u.Message.CommandArguments(), " ")
	flags, err := parseFlags(args)
	if err != nil {
		log.Println(err)
		api.Send(tgbotapi.NewMessage(u.Message.Chat.ID, errorMsg))
		return
	}

	var msgs []string

	switch u.Message.Command() {
	case todayCmd:
		msgs = t.todayCmdResponse(flags)
	case futureCmd:
		msgs = t.futureCmdResponse(flags)
	case pastCmd:
		msgs = t.pastCmdResponse(flags)
	case helpCmd:
		msgs = t.helpCmdResponse(flags)
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

	for i := range msgs {
		msg := tgbotapi.NewMessage(u.Message.Chat.ID, msgs[i])
		api.Send(msg)
	}
}

type flags struct {
	set                    *flag.FlagSet
	tf, ff, pf, af, gf, mf bool
	df                     int
}

func parseFlags(args []string) (*flags, error) {
	f := flags{}
	f.set = flag.NewFlagSet("bot flag set", flag.ContinueOnError)
	if len(args) == 0 {
		return &f, nil
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

func (t *tgUpdHandler) hiCmdResponse(m *tgbotapi.Message) []string {
	msg := hiMsg
	if m.From.FirstName != "" {
		msg = fmt.Sprintf("Привет, %s 👋\n--> /help для справки", m.From.FirstName)
	} else if m.From.UserName != "" {
		msg = fmt.Sprintf("Привет, %s 👋\n--> /help для справки", m.From.UserName)
	}
	return []string{msg}
}

func (t *tgUpdHandler) helpCmdResponse(f *flags) []string {

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

func (t *tgUpdHandler) todayCmdResponse(f *flags) []string {

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

	return t.query(0, opts)
}

func (t *tgUpdHandler) futureCmdResponse(f *flags) []string {

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

	return t.query(f.df, opts)
}

func (t *tgUpdHandler) pastCmdResponse(f *flags) []string {

	if f.set.NArg() > 0 {
		errMsg := fmt.Sprintf("переданы непонятные для меня аргументы --> %v", f.set.Args())
		return []string{errMsg}
	}

	return t.query(f.df, []botDB.QueryOpt{botDB.Past})
}

func (t *tgUpdHandler) query(daysLimit int, opts []botDB.QueryOpt) []string {
	recs, err := t.qm.Query(daysLimit, opts...)
	if err != nil {
		log.Println(err)
		return []string{errorMsg}
	}

	return []string{buildMessage(recs)}
}

func buildMessage(recs []botDB.PurchaseRecord) string {
	var b strings.Builder

	for i := range recs {
		b.WriteString(recs[i].String())
	}

	return b.String()
}
