package botDB

import (
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

// PurchaseRecord represents incoming data that needs
// to be inserted/updated against DB
type PurchaseRecord struct {
	RegistryNumber        string `json:"registry_number"`
	PurchaseSubject       string `json:"purchase_subject"`
	PurchaseId            int64  `json:"purchase_id"`
	PurchaseSubjectAbbr   string `json:"purchase_abbr,omitempty"`
	PurchaseSubjectAbbrId int64
	PurchaseType          string `json:"purchase_type"`
	PurchaseTypeId        int64
	// CollectingDateTime    string `json:"collecting_datetime"`
	// ApprovalDateTime      string `json:"approval_datetime,omitempty"`
	// BiddingDateTime       string `json:"bidding_datetime,omitempty"`
	CollectingDateTime   time.Time `json:"collecting_datetime"`
	ApprovalDateTime     time.Time `json:"approval_datetime,omitempty"`
	BiddingDateTime      time.Time `json:"bidding_datetime,omitempty"`
	Region               string    `json:"region"`
	RegionId             int64
	CustomerType         string `json:"customer_type,omitempty"`
	CustomerTypeId       int64
	MaxPrice             float64 `json:"max_price"`
	ApplicationGuarantee float64 `json:"application_guarantee,omitempty"`
	ContractGuarantee    float64 `json:"contract_guarantee,omitempty"`
	Status               string  `json:"status"`
	StatusId             int64
	OurParticipants      string  `json:"our_participants,omitempty"`
	Estimation           float64 `json:"estimation,omitempty"`
	ETP                  string  `json:"etp,omitempty"`
	ETPId                int64
	Winner               string  `json:"winner,omitempty"`
	WinnerPrice          float64 `json:"winner_price,omitempty"`
	Participants         string  `json:"participants,omitempty"`
	queryOpt             QueryOpt
}

func (p *PurchaseRecord) truncNum() string {
	if len(p.RegistryNumber) < 3 {
		return ""
	}
	return p.RegistryNumber[len(p.RegistryNumber)-3:]
}

func (p *PurchaseRecord) String() string {

	switch p.queryOpt {

	case TodayAuction:
		return p.auctionString()

	case Future, FutureAuction, TodayGo, FutureGo:
		return p.participateString()

	case Today:
		if p.Status == statusGo {
			return p.participateString()
		}
		return p.auctionString()

	case FutureMoney:
		return p.moneyString()

	case Past:
		return p.pastString()

	default:
		return p.generalString()

	}

}

func (p *PurchaseRecord) generalString() string {

	tc := p.CollectingDateTime.Format("02.01.2006 15:04")
	tb := p.BiddingDateTime.Format("02.01.2006 15:04")

	return fmt.Sprintf("[%d] %s\n%s %s\n🔝 %.2f\n⏳ %v\n⏰ %v\n💸 %.2f\nСтатус: %s\nПлощадка: %s\n\n",
		p.PurchaseId, p.RegistryNumber, p.Region, p.PurchaseSubjectAbbr,
		p.MaxPrice, tc, tb, p.ApplicationGuarantee, p.Status, p.ETP)
}

func (p *PurchaseRecord) auctionString() string {

	t := p.BiddingDateTime.Format("15:04")

	return fmt.Sprintf("[%d] %s %s %s\n⏰: %v | ⬇️: %.2f\n\n",
		p.PurchaseId, p.Region, p.truncNum(), p.PurchaseSubjectAbbr, t, p.Estimation)
}

func (p *PurchaseRecord) participateString() string {

	t := p.CollectingDateTime.Format("02.01.2006 15:04")

	return fmt.Sprintf("[%d] %s %s %s\n⏳: %v\n\n",
		p.PurchaseId, p.Region, p.truncNum(), p.PurchaseSubjectAbbr, t)
}

func (p *PurchaseRecord) pastString() string {
	res := '🏆'
	if p.Status == statusLost {
		res = '❌'
	}

	return fmt.Sprintf("[%d] %s %s %s --> %c \n\n",
		p.PurchaseId, p.Region, p.truncNum(), p.PurchaseSubjectAbbr, res)
}

func (p *PurchaseRecord) moneyString() string {
	s := fmt.Sprintf("%s %s \n💸 %.2f\n\n",
		p.Region, p.PurchaseSubjectAbbr, p.ApplicationGuarantee)

	if p.Region == "" {
		s = fmt.Sprintf("%s 💸 %.2f\n\n",
			p.PurchaseSubjectAbbr, p.ApplicationGuarantee)
	}

	return s
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
// supposed to taking a part in insert/update operation
func (p *PurchaseRecord) args() []interface{} {
	return []interface{}{
		p.RegistryNumber,
		p.PurchaseSubject,
		p.PurchaseSubjectAbbrId,
		p.PurchaseTypeId,
		p.CollectingDateTime,
		p.ApprovalDateTime,
		p.BiddingDateTime,
		p.RegionId,
		p.CustomerTypeId,
		p.MaxPrice,
		p.ApplicationGuarantee,
		p.ContractGuarantee,
		p.StatusId,
		p.OurParticipants,
		p.Estimation,
		p.ETPId,
		p.Winner,
		p.WinnerPrice,
		p.Participants,
	}
}
