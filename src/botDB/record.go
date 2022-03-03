package botDB

import (
	"database/sql"
	"fmt"
	"time"
)

// purchase status
const (
	statusGo       = "идем"
	statusEstim    = "расчет"
	statusAuction  = "допущены"
	statusAuction2 = "заявлены"
	statusWin      = "выиграли"
	statusLost     = "не выиграли"
)

const noTime = "--.--.-- --.--"

// PurchaseRecord represents incoming data that needs
// to be inserted/updated against DB
type PurchaseRecord struct {
	RegistryNumber          string          `json:"registry_number"`
	PurchaseSubject         string          `json:"purchase_subject"`
	PurchaseId              int64           `json:"purchase_id"`
	PurchaseSubjectAbbr     string          `json:"purchase_abbr,omitempty"`
	PurchaseSubjectAbbrId   int64           `json:"-"`
	PurchaseType            string          `json:"purchase_type"`
	PurchaseTypeId          int64           `json:"-"`
	CollectingDateTime      time.Time       `json:"collecting_datetime"`
	ApprovalDateTime        time.Time       `json:"approval_datetime,omitempty"`
	ApprovalDateTimeSql     sql.NullTime    `json:"-"`
	BiddingDateTime         time.Time       `json:"bidding_datetime,omitempty"`
	BiddingDateTimeSql      sql.NullTime    `json:"-"`
	Region                  string          `json:"region"`
	RegionId                int64           `json:"-"`
	CustomerType            string          `json:"customer_type,omitempty"`
	CustomerTypeId          int64           `json:"-"`
	MaxPrice                float64         `json:"max_price"`
	ApplicationGuarantee    float64         `json:"application_guarantee,omitempty"`
	ApplicationGuaranteeSql sql.NullFloat64 `json:"-"`
	ContractGuarantee       float64         `json:"contract_guarantee,omitempty"`
	ContractGuaranteeSql    sql.NullFloat64 `json:"-"`
	Status                  string          `json:"status"`
	StatusSql               sql.NullString  `json:"-"`
	StatusId                int64           `json:"-"`
	OurParticipants         string          `json:"our_participants,omitempty"`
	OurParticipantsSql      sql.NullString  `json:"-"`
	Estimation              float64         `json:"estimation,omitempty"`
	EstimationSql           sql.NullFloat64 `json:"-"`
	ETP                     string          `json:"etp,omitempty"`
	EtpSql                  sql.NullString  `json:"-"`
	ETPId                   int64           `json:"-"`
	Winner                  string          `json:"winner,omitempty"`
	WinnerSql               sql.NullString  `json:"-"`
	WinnerPrice             float64         `json:"winner_price,omitempty"`
	WinnerPriceSql          sql.NullFloat64 `json:"-"`
	Participants            string          `json:"participants,omitempty"`
	ParticipantsSql         sql.NullString  `json:"-"`
	queryOpt                QueryOpt        `json:"-"`
}

// Info returns string representation of record
func (p *PurchaseRecord) Info() (string, QueryOpt) {

	switch p.queryOpt {

	case TodayAuction:
		return p.auctionString(), p.queryOpt

	case Future, FutureAuction, TodayGo, FutureGo:
		return p.participateString(), p.queryOpt

	case Today:
		if p.StatusSql.String == statusGo || p.StatusSql.String == statusEstim {
			return p.participateString(), TodayGo
		}
		return p.auctionString(), TodayAuction

	case FutureMoney:
		return p.moneyString(), p.queryOpt

	case Past:
		return p.pastString(), p.queryOpt

	default:
		return p.generalString(), p.queryOpt

	}

}

func (p *PurchaseRecord) truncNum() string {
	if len(p.RegistryNumber) < 3 {
		return ""
	}
	return p.RegistryNumber[len(p.RegistryNumber)-3:]
}

func (p *PurchaseRecord) generalString() string {

	tc := p.CollectingDateTime.Format("02.01.2006 15:04")
	tb := noTime
	if p.BiddingDateTimeSql.Valid {
		tb = p.BiddingDateTimeSql.Time.Format("02.01.2006 15:04")
	}

	return fmt.Sprintf("*[%d]* _%s_\n%s *_%s_*\nНМЦК: *%.2f ₽* 🔝\nПодача: *%v* ⏳\nАукцион: *%v* ⏰\nОбеспечение: *%.2f* 💸\nСтатус: *%s*\nПлощадка: *%s*\n\n",
		p.PurchaseId, p.RegistryNumber, p.Region, p.PurchaseSubjectAbbr,
		p.MaxPrice, tc, tb, p.ApplicationGuaranteeSql.Float64, p.StatusSql.String, p.EtpSql.String)
}

func (p *PurchaseRecord) auctionString() string {
	tb := noTime
	if p.BiddingDateTimeSql.Valid {
		tb = p.BiddingDateTimeSql.Time.Format("15:04")
	}
	ptc := "--не установлен--"
	if p.OurParticipantsSql.Valid {
		ptc = p.OurParticipantsSql.String
	}

	return fmt.Sprintf("*[%d]* %s *_%s %s_*\nВремя: *%v* ⏰\nРасчёт: *%.2f* ⬇️\nПлощадка: *%s*\nУчастник: *%s*\n\n",
		p.PurchaseId, p.Region, p.truncNum(), p.PurchaseSubjectAbbr, tb,
		p.EstimationSql.Float64, p.EtpSql.String, ptc)
}

func (p *PurchaseRecord) participateString() string {
	var t string
	if p.StatusSql.String == statusGo || p.StatusSql.String == statusEstim {
		t = fmt.Sprintf("Подача до: *_%v_* ⏳", p.CollectingDateTime.Format("02.01.2006 15:04"))
	} else {
		t = fmt.Sprintf("Аукцион: *_%v_* ⏰", p.BiddingDateTimeSql.Time.Format("02.01.2006 15:04"))
	}

	return fmt.Sprintf("*[%d]* %s *_%s %s_*\n%s\nСтатус: *%s*\n\n",
		p.PurchaseId, p.Region, p.truncNum(), p.PurchaseSubjectAbbr, t, p.StatusSql.String)
}

func (p *PurchaseRecord) pastString() string {
	res := '🏆'
	if p.StatusSql.String == statusLost {
		res = '❌'
	}
	tb := noTime
	if p.BiddingDateTimeSql.Valid {
		tb = p.BiddingDateTimeSql.Time.Format("02.01.2006")
	}

	return fmt.Sprintf("*[%d]* *_%s %s %s_*\nДата проведения *_%v_*\n*Результат ->* %c \n\n",
		p.PurchaseId, p.Region, p.truncNum(), p.PurchaseSubjectAbbr, tb, res)
}

func (p *PurchaseRecord) moneyString() string {
	return fmt.Sprintf("Участник: *%s*\nСо статусом *%s* -> *_%.2f ₽_* 💸\n\n",
		p.OurParticipantsSql.String, p.StatusSql.String, p.ApplicationGuaranteeSql.Float64)
}

// setForeignKeys take reference map and check self id
// fields if they exist in map. If so field is sets to that
// id. Otherwise update map with new data is formed. In the end
// update map will be either empty or containing new data
func (p *PurchaseRecord) setForeignKeys(rm refTablesMap) map[string]string {
	var ok bool
	// this is update map in case of
	// record values is new for database
	um := make(map[string]string)

	// compare values of record against reference map
	// in case it's new we put it in the update map
	if p.PurchaseSubjectAbbrId, ok = rm[purchaseStringCodeTableName][p.PurchaseSubjectAbbr]; !ok {
		um[purchaseStringCodeTableName] = p.PurchaseSubjectAbbr
	}
	if p.PurchaseTypeId, ok = rm[purchTypeTableName][p.PurchaseType]; !ok {
		um[purchTypeTableName] = p.PurchaseType
	}
	if p.CustomerTypeId, ok = rm[custTableName][p.CustomerType]; !ok {
		um[custTableName] = p.CustomerType
	}
	if p.RegionId, ok = rm[regionTableName][p.Region]; !ok {
		um[regionTableName] = p.Region
	}
	if p.ETPId, ok = rm[etpTableName][p.ETP]; !ok {
		um[etpTableName] = p.ETP
	}
	if p.StatusId, ok = rm[statusTableName][p.Status]; !ok {
		um[statusTableName] = p.Status
	}

	return um
}

// args returns PurchaseRecord fields that
// supposed to taking a part in insert/update/query operation
// based on provided table option
func (p *PurchaseRecord) args(to tableOpt) []interface{} {

	switch to {
	case query:
		return []interface{}{
			&p.PurchaseId, &p.RegistryNumber, &p.PurchaseSubject, &p.PurchaseSubjectAbbr,
			&p.PurchaseType, &p.CollectingDateTime, &p.ApprovalDateTimeSql,
			&p.BiddingDateTimeSql, &p.Region, &p.CustomerType, &p.MaxPrice,
			&p.ApplicationGuaranteeSql, &p.ContractGuaranteeSql, &p.StatusSql, &p.OurParticipantsSql,
			&p.EstimationSql, &p.EtpSql, &p.WinnerSql, &p.WinnerPriceSql, &p.ParticipantsSql,
		}
	case queryMoney:
		return []interface{}{&p.OurParticipantsSql, &p.StatusSql, &p.ApplicationGuaranteeSql}
	default:
		return []interface{}{
			&p.RegistryNumber, &p.PurchaseSubject, &p.PurchaseSubjectAbbrId,
			&p.PurchaseTypeId, &p.CollectingDateTime, &p.ApprovalDateTime,
			&p.BiddingDateTime, &p.RegionId, &p.CustomerTypeId,
			&p.MaxPrice, &p.ApplicationGuarantee, &p.ContractGuarantee,
			&p.StatusId, &p.OurParticipants, &p.Estimation,
			&p.ETPId, &p.Winner, &p.WinnerPrice, &p.Participants,
		}
	}
}
