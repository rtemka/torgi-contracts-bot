package memdb

import (
	"database/sql"
	"fmt"
	"io"
	botDB "tbot/pkg/db"
	"time"
)

var mockErr = fmt.Errorf("intentional error")

var MockPurchase = botDB.PurchaseRecord{
	RegistryNumber:          "0859200001122007104",
	PurchaseSubject:         "test subject",
	PurchaseId:              1,
	PurchaseSubjectAbbr:     "ABS",
	PurchaseSubjectAbbrId:   1,
	PurchaseType:            "test type",
	PurchaseTypeId:          1,
	CollectingDateTime:      time.Unix(1657016101, 0),
	ApprovalDateTime:        time.Unix(1657016101, 0),
	ApprovalDateTimeSql:     sql.NullTime{Time: time.Unix(1657016101, 0), Valid: true},
	BiddingDateTime:         time.Unix(1657016101, 0),
	BiddingDateTimeSql:      sql.NullTime{Time: time.Unix(1657016101, 0), Valid: true},
	Region:                  "test region",
	RegionId:                1,
	CustomerType:            "test customer",
	CustomerTypeId:          1,
	MaxPrice:                100000,
	ApplicationGuarantee:    100000,
	ApplicationGuaranteeSql: sql.NullFloat64{Float64: 100000, Valid: true},
	ContractGuarantee:       0,
	ContractGuaranteeSql:    sql.NullFloat64{Float64: 100000, Valid: true},
	Status:                  "test status",
	StatusSql:               sql.NullString{},
	StatusId:                1,
	OurParticipants:         "part1, part 2",
	OurParticipantsSql:      sql.NullString{},
	Estimation:              100000,
	EstimationSql:           sql.NullFloat64{Float64: 100000, Valid: true},
	ETP:                     "test etp",
	EtpSql:                  sql.NullString{String: "test etp", Valid: true},
	ETPId:                   0,
	Winner:                  "test winner",
	WinnerSql:               sql.NullString{String: "test winner", Valid: true},
	WinnerPrice:             0,
	WinnerPriceSql:          sql.NullFloat64{Float64: 100000, Valid: true},
	Participants:            "test participants",
	ParticipantsSql:         sql.NullString{String: "test participants", Valid: true},
	QueryType:               botDB.General,
}

// MemDB is for testing purposes
type MemDB struct{ needErr bool }

func New(needErr bool) *MemDB {
	return &MemDB{
		needErr: needErr,
	}
}

func (d MemDB) Upsert(_ io.ReadCloser) error {
	if d.needErr {
		return mockErr
	}
	return nil
}

func (d MemDB) Delete(_ io.ReadCloser) error { return nil }

func (d MemDB) Query(_ int, _ ...botDB.QueryOpt) ([]botDB.PurchaseRecord, error) {
	return nil, nil
}

func (d MemDB) QueryRow(_ int64) (botDB.PurchaseRecord, error) {
	return botDB.PurchaseRecord{}, nil
}
