package botDB

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var ErrNoRows = sql.ErrNoRows

// BotDB responsible for database management
type BotDB struct {
	db      *sql.DB
	records []PurchaseRecord
	tk      tablesKeeper
	refMap  refTablesMap
}

// NewBotDB is database BotDB manager constructor.
// Expects established database connection
func NewBotDB(db *sql.DB) *BotDB {
	return &BotDB{
		db:      db,
		records: nil,
		tk:      newTables(),
		refMap:  nil,
	}
}

// botDbError represents information
// about error condition due working
// with bot database
type botDbError struct {
	Context string
	Err     error
}

func (be *botDbError) Error() string {
	return fmt.Sprintf("%s: %v", be.Context, be.Err)
}

func newBotDbError(operation, stmt string, err error, args ...any) *botDbError {
	ctx := fmt.Sprintf("op=%s stmt=%s args=%v", operation, stmt, args)
	return &botDbError{Context: ctx, Err: err}
}

// tablesKeeper provides information about available DB tables
type tablesKeeper interface {
	table(name string) table // returns table for the giving name
	tables() []table         // returns available DB tables
}

// Single table interface. Provides key
// information about concrete table
type table interface {
	name() string                  // Table name
	primaryKeyCol(tableOpt) string // Table primary key
	nameKeyCol() string            // Table significant name column
	refTables() []table            // Reference tables (foreign key tables)
	columns(tableOpt) []string     // Table columns
	joinOn(table) string           // Returns key that joins this table with provided table
}

// map of maps with values of reference table.
// It goes like this: "tableName":map["NameColumnKey":PrimaryKey]
type refTablesMap map[string]map[string]int64

// OpenDB establish connection to the database
func OpenDB(DbParams string) (*sql.DB, error) {

	db, err := sql.Open("postgres", DbParams)
	if err != nil {
		return nil, err
	}

	return db, db.Ping()
}

// Upsert reading from incoming update source
// and try to perform an insert/update operation
func (m *BotDB) Upsert(rc io.ReadCloser) error {

	// unmarshalling incoming data
	// and put them inside BotDB
	err := json.NewDecoder(rc).Decode(&m.records)
	if err != nil {
		return err
	}

	// deferring clean up operations in case
	// of early returns due to error occurrences
	defer func() {
		m.records = nil
		m.refMap = nil
	}()

	if len(m.records) == 0 {
		return fmt.Errorf("the length of incoming records is zero")
	}

	// work to be done before updating
	if err := m.prepareUpdate(); err != nil {
		return err
	}

	return m.upsrt()
}

// prepareUpdate sets up a reference map for BotDB,
// foreign keys for the record
func (m *BotDB) prepareUpdate() error {
	// setting up reference tables map
	err := m.setRefMaps()
	if err != nil {
		return err
	}

	// setting up foreign keys from reference tables map
	for i := range m.records {
		err = m.setForeignKeys(&m.records[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// core upsert operation
func (m *BotDB) upsrt() error {
	// get main table
	t := m.tk.table(purchTableName)

	// construct options for building an
	// query statement
	opts := stmtOpts{
		tableName:   t.name(),
		conflictKey: t.primaryKeyCol(primaryKey),
		multiplier:  len(m.records),
		withUpdate:  true,
		// get table columns that taking part in update
		cols: t.columns(upsert),
	}

	// building an upsert query
	stmt := upsertStatement(opts)

	// get arguments for the query
	args := m.buildArgs()

	// transaction
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.Exec(stmt, args...)
	if err != nil {
		return newBotDbError("BotDB: upsrt", stmt, err, args...)
	}

	return tx.Commit()
}

func (m *BotDB) setForeignKeys(p *PurchaseRecord) error {

	// gives the record reference table map
	// to set foreign keys fields
	// it returns back an update map in case
	// there is new data for foreign tables
	um := p.setForeignKeys(m.refMap)

	// if there is no new data -> we done
	if len(um) == 0 {
		return nil
	}

	// otherwise we update iternal refMap
	err := m.updateRefMap(um)
	if err != nil {
		return err
	}

	// then with the updated refMap in hand
	// we go again
	return m.setForeignKeys(p)
}

// buildArgs takes every record in a BotDB,
// build arguments for each and then append them to
// one big arguments slice
func (m *BotDB) buildArgs() []interface{} {

	args := make([]interface{}, 0, len(m.records)*(purchTableColsCount-1))

	for i := range m.records {
		args = append(args, m.records[i].args(upsert)...)
	}

	return args
}

// setRefMaps builds a map of tables
// that our main table is referecing at
func (m *BotDB) setRefMaps() error {

	// get tables that our main table is referecing at
	refTables := m.tk.table(purchTableName).refTables()

	// plug them in to the BotDB
	m.refMap = make(refTablesMap, len(refTables))

	// fill them with data
	for i := range refTables {
		err := m.fillMap(refTables[i])
		if err != nil {
			return newBotDbError("BotDB: setRefMaps", "", err)
		}
	}

	return nil
}

// fills the BotDBs refMap with provided table data
func (m *BotDB) fillMap(t table) error {
	var (
		id   int64
		name string
	)

	// build statement for querying table id and name info from db
	q := selectWhereStmt(stmtOpts{tableName: t.name(), cols: t.columns(upsert)})

	rows, err := m.db.Query(q)
	if err != nil {
		return newBotDbError("BotDB: fillMap: Query", q, err)
	}

	defer rows.Close()

	m.refMap[t.name()] = make(map[string]int64)

	for rows.Next() {
		err := rows.Scan(&id, &name)
		if err != nil {
			return newBotDbError("BotDB: fillMap: Scan", q, err)
		}
		m.refMap[t.name()][name] = id
	}

	return rows.Err()
}

// updateRefMap executes when incoming records comes with
// new data for the referencing tables.
// It updates DB and then BotDB internal refMap with that data
func (m *BotDB) updateRefMap(um map[string]string) error {

	var id int64

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for k, v := range um {

		// get the table
		t := m.tk.table(k)

		q := fmt.Sprintf("insert into %s (%s) values ($1) returning %s;",
			t.name(), t.nameKeyCol(), t.primaryKeyCol(primaryKey))

		err := tx.QueryRow(q, v).Scan(&id)
		if err != nil {
			return newBotDbError("BotDB: updateRefMap", q, err, v)
		}

		// plug inserted id in refMap
		m.refMap[t.name()][t.nameKeyCol()] = id
	}

	return tx.Commit()
}

// Delete is for removing old records from DB only.
// This should be done from time to time so that the database
// does not grow in size too much due to restrictions of heroku platform.
// No functionality beyond that is provided.
func (m *BotDB) Delete(_ io.ReadCloser) error {

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.Exec(purchDeleteStatement)
	if err != nil {
		return newBotDbError("BotDB: Delete", purchDeleteStatement, err)
	}

	return tx.Commit()
}

// Query performs select operations from
// database based on provided query options
func (m *BotDB) Query(daysLimit int, qopts ...QueryOpt) ([]PurchaseRecord, error) {

	var recs []PurchaseRecord
	var r PurchaseRecord

	// get main table
	t := m.tk.table(purchTableName)

	// range over provided query options
	for _, q := range qopts {

		opts := q.stmtOpts(daysLimit, t) // build statement options
		stmt := selectWhereStmt(opts)    // build statement
		rows, err := m.db.Query(stmt)
		if err != nil {
			return nil, newBotDbError("BotDB: Query", stmt, err)
		}

		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(r.args(q.tableOpt())...)
			if err != nil {
				return nil, newBotDbError("BotDB: Query Scan", "", err, r.args(q.tableOpt())...)
			}
			// we add specific query option to the record
			// this needed for the record to properly
			// build string info about itself
			r.QueryType = q
			recs = append(recs, r)
		}

		if err = rows.Err(); err != nil {
			return nil, err
		}

		rows.Close()
	}

	return recs, nil
}

// QueryRow looks for one specific row by id
func (m *BotDB) QueryRow(id int64) (PurchaseRecord, error) {
	var r PurchaseRecord
	if id == 0 {
		return r, fmt.Errorf("invalid identifier %d", id)
	}

	// get main table
	t := m.tk.table(purchTableName)

	opts := stmtOpts{
		tableName:   t.name(),
		fromClause:  buildFromClause(t, left),
		whereClause: fmt.Sprintf("where %s = %d", t.primaryKeyCol(secondaryKey), id),
		cols:        t.columns(query),
	}

	stmt := selectWhereStmt(opts)
	err := m.db.QueryRow(stmt).Scan(r.args(query)...)
	if err != nil {
		if err == sql.ErrNoRows {
			return r, ErrNoRows
		}
		return r, newBotDbError("BotDB: QueryRow", stmt, err, r.args(query)...)
	}

	// we add specific query option to the record
	// this needed for the record to properly
	// build string info about itself
	r.QueryType = General

	return r, nil
}

// QueryOpt is the parameter for
// the Read/Query operation
type QueryOpt int

// Read/Query operation parameter
const (
	_ QueryOpt = iota
	General
	Today
	Future
	Past
	TodayAuction
	TodayGo
	FutureAuction
	FutureGo
	FutureMoney
)

// String returns string representation of queryOpt
func (q QueryOpt) String() string {
	return []string{"", "", "*Сегодня*\n\n", "*Впереди*\n\n", "*Результаты*\n\n",
		"*Аукционы* ⚔️\n\n", "*Заявки* 🏃\n\n", "*Аукционы* ⚔️\n\n",
		"*Заявки* 🏃\n\n", "*Обеспечения заявок* 💰\n\n"}[q]
}

// tableOpt returns tableOpt option based on self
func (q QueryOpt) tableOpt() tableOpt {
	switch q {
	case FutureMoney:
		return queryMoney
	default:
		return query
	}
}

// stmtOpts builds stmtOpts based on self
func (q QueryOpt) stmtOpts(daysLimit int, t table) stmtOpts {
	switch q {
	case FutureMoney:
		return stmtOpts{
			tableName:   t.name(),
			fromClause:  buildFromClause(t, left),
			whereClause: q.whereClause(daysLimit),
			groupBy:     []string{ourParticipants, statusName},
			cols:        t.columns(queryMoney),
		}
	default:
		return stmtOpts{
			tableName:   t.name(),
			fromClause:  buildFromClause(t, left),
			whereClause: q.whereClause(daysLimit),
			orderBy:     []string{biddingColumn},
			cols:        t.columns(query),
		}
	}
}

// whereClause builds where clause based on self
func (q QueryOpt) whereClause(daysLimit int) string {
	var b strings.Builder

	switch q {

	case Today:
		pd := plusDays(time.Now().Weekday())
		b.WriteString(fmt.Sprintf("where (%s in ('%s', '%s')", statusName, statusAuction, statusAuction2))
		b.WriteString(fmt.Sprintf(" and date_trunc('day', %s) = current_date::timestamp)", biddingColumn))
		b.WriteString(fmt.Sprintf("  or (%s in ('%s', '%s')", statusName, statusGo, statusEstim))
		b.WriteString(fmt.Sprintf(" and date_trunc('day', %s) = (current_date+%d)::timestamp)", collectingColumn, pd))

	case Future:
		b.WriteString(fmt.Sprintf("where (%s in ('%s', '%s')", statusName, statusAuction, statusAuction2))
		if daysLimit > 0 {
			b.WriteString(fmt.Sprintf(" and %s between (current_date+1)::timestamp and (current_date+%d)::timestamp)", biddingColumn, daysLimit))
			b.WriteString(fmt.Sprintf("or (%s in ('%s', '%s')", statusName, statusGo, statusEstim))
			b.WriteString(fmt.Sprintf(" and %s between (current_date+1)::timestamp and (current_date+%d)::timestamp)", collectingColumn, daysLimit))
			return b.String()
		}
		b.WriteString(fmt.Sprintf(" and %s >= (current_date+1)::timestamp)", biddingColumn))
		b.WriteString(fmt.Sprintf("or (%s in ('%s', '%s')", statusName, statusGo, statusEstim))
		b.WriteString(fmt.Sprintf(" and %s >= (current_date+1)::timestamp)", collectingColumn))

	case Past:
		if daysLimit > 0 {
			return fmt.Sprintf("where (%s in ('%s', '%s') and %s between (current_date-%d)::timestamp and current_date::timestamp)",
				statusName, statusWin, statusLost, biddingColumn, daysLimit)
		}
		return fmt.Sprintf("where (%s in ('%s', '%s') and %s < current_date::timestamp)",
			statusName, statusWin, statusLost, biddingColumn)

	case TodayAuction:
		b.WriteString(fmt.Sprintf("where (%s in ('%s', '%s')", statusName, statusAuction, statusAuction2))
		b.WriteString(fmt.Sprintf(" and date_trunc('day', %s) = current_date::timestamp)", biddingColumn))

	case TodayGo:
		pd := plusDays(time.Now().Weekday())
		b.WriteString(fmt.Sprintf("where (%s in ('%s', '%s')", statusName, statusGo, statusEstim))
		b.WriteString(fmt.Sprintf(" and date_trunc('day', %s) = (current_date+%d)::timestamp)", collectingColumn, pd))

	case FutureAuction:
		b.WriteString(fmt.Sprintf("where (%s in ('%s', '%s')", statusName, statusAuction, statusAuction2))
		if daysLimit > 0 {
			b.WriteString(fmt.Sprintf(" and %s between (current_date+1)::timestamp and (current_date+%d)::timestamp)", biddingColumn, daysLimit))
			return b.String()
		}
		b.WriteString(fmt.Sprintf(" and %s >= (current_date+1)::timestamp)", biddingColumn))

	case FutureGo:
		b.WriteString(fmt.Sprintf("where (%s in ('%s', '%s')", statusName, statusGo, statusEstim))
		if daysLimit > 0 {
			b.WriteString(fmt.Sprintf(" and %s between (current_date+1)::timestamp and (current_date+%d)::timestamp)", collectingColumn, daysLimit))
			return b.String()
		}
		b.WriteString(fmt.Sprintf(" and %s >= (current_date+1)::timestamp)", collectingColumn))

	case FutureMoney:
		b.WriteString(fmt.Sprintf("where (%s in ('%s', '%s')", statusName, statusGo, statusEstim))
		if daysLimit > 0 {
			b.WriteString(fmt.Sprintf(" and %s between (current_date+1)::timestamp and (current_date+%d)::timestamp)", collectingColumn, daysLimit))
			return b.String()
		}
		b.WriteString(fmt.Sprintf(" and %s >= (current_date+1)::timestamp)", collectingColumn))
	}

	return b.String()
}

// plusDays returns amount of days that
// need to be added to provided weekday to get next
// workday
func plusDays(d time.Weekday) int {
	switch d {
	case time.Friday:
		return 3
	case time.Saturday:
		return 2
	default:
		return 1
	}
}
