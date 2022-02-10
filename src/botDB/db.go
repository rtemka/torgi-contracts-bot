package botDB

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"

	_ "github.com/lib/pq"
)

// Model responsible for database management
type Model struct {
	db      *sql.DB
	records []PurchaseRecord
	tk      tablesKeeper
	refMap  refTablesMap
}

// NewModel is database model manager constructor.
// Expects established database connection
func NewModel(db *sql.DB) *Model {
	return &Model{
		db:      db,
		records: nil,
		tk:      newTables(),
		refMap:  nil,
	}
}

// tablesKeeper provides information about available DB tables
type tablesKeeper interface {
	table(name string) table // returns table for the giving name

	tables() []table // returns available DB tables
}

// Single table interface. Provides key
// information about concrete table
type table interface {
	name() string // Table name

	primaryKeyCol() string // Table primary key

	nameKeyCol() string // Table most significant name column

	refTables() []table // Reference tables (foreign key tables)

	columns() []string // Table columns
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

	// if err = db.Ping(); err != nil {
	// 	return nil, err
	// }

	return db, db.Ping()
}

// Upsert reading from incoming update source
// and try to perform an insert/update operation
func (m *Model) Upsert(rc io.ReadCloser) error {

	// unmarshalling incoming data
	// and put them inside Model
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
		return fmt.Errorf(
			"error while processing database update: the length of incoming records is zero")
	}

	// work to be done before updating
	if err := m.prepareUpdate(); err != nil {
		return err
	}

	return m.upsrt()
}

// prepareUpdate
func (m *Model) prepareUpdate() error {
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
func (m *Model) upsrt() error {
	// get main table
	t := m.tk.table(purchTableName)

	// construct options for building an
	// query statement
	opts := opts{
		tableName:  t.name(),
		primaryKey: t.nameKeyCol(),
		multiplier: len(m.records),
		withUpdate: true,
	}

	// get table columns that taking part in update
	cols := t.columns()

	// building an upsert query
	q := upsertStatement(opts, cols)

	// get arguments for the query
	args := m.buildArgs()

	// transaction
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.Exec(q, args)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (m *Model) setForeignKeys(p *PurchaseRecord) error {

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

// buildArgs takes every record in a Model,
// build arguments for each and then append them to
// one big arguments slice
func (m *Model) buildArgs() []interface{} {

	args := make([]interface{}, 0, len(m.records)*(purchTableColsCount-1))

	for i := range m.records {
		args = append(args, m.records[i].args()...)
	}

	return args
}

// setRefMaps builds a map of tables
// that our main table is referecing at
func (m *Model) setRefMaps() error {

	// get tables that our main table is referecing at
	refTables := m.tk.table(purchTableName).refTables()

	// plug them in to the Model
	m.refMap = make(refTablesMap, len(refTables))

	// fill them with data
	for i := range refTables {
		err := m.fillMap(refTables[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// fills the models refMap with provided table data
func (m *Model) fillMap(t table) error {
	var (
		id   int64
		name string
	)
	// build statement for querying table id and name info from db
	q := idNameStatement(t.name(), t.primaryKeyCol(), t.nameKeyCol())

	rows, err := m.db.Query(q)
	if err != nil {
		return err
	}

	defer rows.Close()

	m.refMap[t.name()] = make(map[string]int64)

	for rows.Next() {
		err := rows.Scan(&id, &name)
		if err != nil {
			return err
		}
		m.refMap[t.name()][name] = id
	}

	if err = rows.Err(); err != nil {
		return err
	}

	rows.Close()

	return nil
}

// updateRefMap executes when incoming records comes with
// new data for the referencing tables.
// It updates DB and then Model internal refMap with that data
func (m *Model) updateRefMap(um map[string]string) error {

	var id int64

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for k, v := range um {

		// get the table
		t := m.tk.table(k)

		q := fmt.Sprintf("insert into %s (%s) values ($1);", t.name(), t.nameKeyCol())

		res, err := tx.Exec(q, v)
		if err != nil {
			return err
		}

		id, err = res.LastInsertId()
		if err != nil {
			return err
		}

		// plug inserted id in refMap
		m.refMap[t.name()][t.nameKeyCol()] = id
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Delete is for removing old records from DB only.
// This should be done from time to time so that the database
// does not grow in size too much due to restrictions of heroku platform.
// No functionality beyond that is provided.
func (m *Model) Delete() error {

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.Exec(purchDeleteStatement)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (m *Model) Query(opt QueryOpt, daysCut int) ([]PurchaseRecord, error) {

	return nil, nil
}

func (m *Model) QueryRow(id int64) (PurchaseRecord, error) {

	return PurchaseRecord{}, nil
}

// QueryOpt is the parameter for
// the Read operation
type QueryOpt int

// Read operation parameter
const (
	_ QueryOpt = iota
	Today
	TodayAuction
	TodayGo
	Future
	FutureAuction
	FutureGo
	FutureMoney
	Past
)
