package bot

import (
	"flag"
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
	todayHelpMsg = "Имя:\n\t" + todayKey + "\n\nИспользование:\n\t" + todayKey + "\t[опции]...\n2020-07-16 10:15:06Описание:" +
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

// amount of command keys
const (
	pastCmdKeysNum   = 0
	todayCmdKeysNum  = 2
	futureCmdKeysNum = 3
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

	args := strings.Split(m.CommandArguments(), " ")
	if len(args) == 0 {
		return []string{generalHelpMsg}
	}

	var tf, ff, pf bool

	hCmd := flag.NewFlagSet(helpCmd, flag.ContinueOnError)
	hCmd.BoolVar(&tf, todayKey, false, cmdHelp+todayKey)
	hCmd.BoolVar(&ff, futureKey, false, cmdHelp+futureKey)
	hCmd.BoolVar(&pf, pastKey, false, cmdHelp+pastKey)
	err := hCmd.Parse(args)
	if err != nil {
		log.Println(err)
		return []string{errorMsg}
	}

	var msg []string

	switch {
	case tf:
		msg = append(msg, todayHelpMsg)
		fallthrough
	case ff:
		msg = append(msg, futureHelpMsg)
		fallthrough
	case pf:
		msg = append(msg, pastHelpMsg)
	default:
		msg = []string{generalHelpMsg}
	}

	return msg
}

func (t *tgUpdHandler) todayCmdResponse(m *tgbotapi.Message) []string {

	args := strings.Split(m.CommandArguments(), " ")
	if len(args) == 0 {
		return t.query(0, []botDB.QueryOpt{botDB.Today})
	}

	var af, gf bool

	tCmd := flag.NewFlagSet(todayCmd, flag.ContinueOnError)
	tCmd.BoolVar(&af, auctionKey, false, auctionKeyUsg)
	tCmd.BoolVar(&gf, goKey, false, goKeyUsg)
	tCmd.BoolVar(&af, auctionKeyLong, false, auctionKeyUsg)
	tCmd.BoolVar(&gf, goKeyLong, false, goKeyUsg)
	err := tCmd.Parse(args)
	if err != nil {
		log.Println(err)
		return []string{errorMsg}
	}

	opts := make([]botDB.QueryOpt, 0, todayCmdKeysNum)
	switch {
	case af:
		opts = append(opts, botDB.TodayAuction)
		fallthrough
	case gf:
		opts = append(opts, botDB.TodayGo)
	case len(args) > 0:
		return []string{errorMsg}
	}

	return t.query(0, opts)
}

func (t *tgUpdHandler) futureCmdResponse(m *tgbotapi.Message) []string {

	args := strings.Split(m.CommandArguments(), " ")
	if len(args) == 0 {
		return t.query(0, []botDB.QueryOpt{botDB.Future})
	}

	var af, gf, mf bool
	var df int

	fCmd := flag.NewFlagSet(futureCmd, flag.ContinueOnError)
	fCmd.BoolVar(&af, auctionKey, false, auctionKeyUsg)
	fCmd.BoolVar(&gf, goKey, false, goKeyUsg)
	fCmd.BoolVar(&af, auctionKeyLong, false, auctionKeyUsg)
	fCmd.BoolVar(&gf, goKeyLong, false, goKeyUsg)
	fCmd.BoolVar(&mf, moneyKey, false, moneyKeyUsg)
	fCmd.BoolVar(&mf, moneyKeyLong, false, moneyKeyUsg)
	fCmd.IntVar(&df, daysKey, 0, daysKeyUsg)
	fCmd.IntVar(&df, daysKeyLong, 0, daysKeyUsg)
	err := fCmd.Parse(args)
	if err != nil {
		log.Println(err)
		return []string{errorMsg}
	}

	opts := make([]botDB.QueryOpt, 0, futureCmdKeysNum)
	switch {
	case af:
		opts = append(opts, botDB.FutureAuction)
		fallthrough
	case gf:
		opts = append(opts, botDB.FutureGo)
		fallthrough
	case mf:
		opts = append(opts, botDB.FutureMoney)
	default:
		return []string{errorMsg}
	}

	return t.query(df, opts)
}

func (t *tgUpdHandler) pastCmdResponse(m *tgbotapi.Message) []string {

	args := strings.Split(m.CommandArguments(), " ")
	if len(args) == 0 {
		return t.query(0, []botDB.QueryOpt{botDB.Past})
	}

	var df int

	pCmd := flag.NewFlagSet(pastCmd, flag.ContinueOnError)
	pCmd.IntVar(&df, daysKey, 0, daysKeyUsg)
	pCmd.IntVar(&df, daysKeyLong, 0, daysKeyUsg)
	err := pCmd.Parse(args)
	if err != nil {
		log.Println(err)
		return []string{errorMsg}
	}

	return t.query(df, []botDB.QueryOpt{botDB.Past})
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
		b.WriteString(recs[i].InfoString())
	}

	return b.String()
}
