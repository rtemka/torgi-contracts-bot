package botDB

// Purchase Table column
const (
	purchTableColsCount  = 20
	purchTableName       = "purchase_registry"
	registryNumber       = "registry_number"
	purchaseID           = "purchase_id"
	purchaseSubject      = "purchase_subject"
	purchaseABBR         = "purchase_string_code"
	purchaseTypeColumn   = "purchase_type_id"
	collectingColumn     = "collecting"
	approvalColumn       = "approval_date"
	biddingColumn        = "bidding"
	regionColumn         = "region_id"
	customerTypeColumn   = "customer_type_id"
	maxPrice             = "max_price"
	applicationGuarantee = "application_guarantee"
	contractGuarantee    = "contract_guarantee"
	statusColumn         = "status_id"
	ourParticipants      = "our_participants"
	estimationColumn     = "estimation"
	etpColumn            = "etp_id"
	winnerColumn         = "winner"
	winnerPrice          = "winner_price"
	participantsColumn   = "participants"
)

// Customer Types Table column
const (
	custTableColsCount = 2
	custTableName      = "customer_types"
	customerTypeID     = "customer_type_id"
	customerTypeName   = "customer_type_name"
)

// Purchase Types Table column
const (
	purchTypeTableColsCount = 2
	purchTypeTableName      = "purchase_types"
	purchaseTypeID          = "purchase_type_id"
	purchaseTypeName        = "purchase_type_name"
)

// Regions Table column
const (
	regionTableColsCount = 2
	regionTableName      = "regions"
	regionID             = "region_id"
	regionName           = "region_name"
)

// ETP Table column
const (
	etpTableColsCount = 2
	etpTableName      = "etp"
	etpID             = "etp_id"
	etpName           = "etp_name"
)

// Status Table column
const (
	statusTableColsCount = 2
	statusTableName      = "statuses"
	statusID             = "status_id"
	statusName           = "status_name"
)

// Purchase string code Table column
const (
	purchaseStringCodeTableColsCount = 3
	purchaseStringCodeTableName      = "purchase_string_codes"
	purchaseStringCode               = "purchase_string_code"
	purchaseStringCodeName           = "purchase_string_code_name"
)

// Delete statement for cleaning up space in DB
const (
	purchDeleteStatement = `delete from ` + purchTableName +
		` where ` + biddingColumn + ` < current_date - interval '2 months'; vacuum full;`
)

// Struct for holding up database tables logic
type tables struct{}

func newTables() tables { return tables{} }

// table returns table for the giving name
func (t tables) table(name string) table {
	switch name {
	case custTableName:
		return &custTypeTable{}
	case regionTableName:
		return &regionTable{}
	case etpTableName:
		return &etpTable{}
	case purchTypeTableName:
		return &purchTypeTable{}
	case purchaseStringCodeTableName:
		return &purchStrCodeTable{}
	case statusTableName:
		return &statusTable{}
	case purchTableName:
		return &purchaseTable{}
	}

	return nil
}

// tables returns available DB tables
func (t tables) tables() []table {
	return []table{
		&purchaseTable{},
		&etpTable{},
		&purchStrCodeTable{},
		&custTypeTable{},
		&regionTable{},
		&purchTypeTable{},
		&statusTable{},
	}
}

// tables types
type purchaseTable struct{}
type etpTable struct{}
type regionTable struct{}
type statusTable struct{}
type purchStrCodeTable struct{}
type purchTypeTable struct{}
type custTypeTable struct{}

func (t etpTable) name() string {
	return etpTableName
}

func (t etpTable) primaryKeyCol() string {
	return etpID
}

func (t etpTable) secondaryKeyCol() string { return "" }

func (t etpTable) nameKeyCol() string {
	return etpName
}

func (t etpTable) columns() []string {
	return []string{
		etpID,
		etpName,
	}
}

func (t etpTable) refTables() []table { return nil }

func (t purchTypeTable) name() string {
	return purchTypeTableName
}

func (t purchTypeTable) primaryKeyCol() string {
	return purchaseTypeID
}

func (t purchTypeTable) secondaryKeyCol() string { return "" }

func (t purchTypeTable) nameKeyCol() string {
	return purchaseTypeName
}

func (t purchTypeTable) columns() []string {
	return []string{
		purchaseTypeID,
		purchaseTypeName,
	}
}

func (t purchTypeTable) refTables() []table { return nil }

func (t regionTable) name() string {
	return regionTableName
}

func (t regionTable) primaryKeyCol() string {
	return regionID
}

func (t regionTable) secondaryKeyCol() string { return "" }

func (t regionTable) nameKeyCol() string {
	return regionName
}

func (t regionTable) columns() []string {
	return []string{
		regionID,
		regionName,
	}
}

func (t regionTable) refTables() []table { return nil }

func (t custTypeTable) name() string {
	return custTableName
}

func (t custTypeTable) primaryKeyCol() string {
	return customerTypeID
}

func (t custTypeTable) secondaryKeyCol() string { return "" }

func (t custTypeTable) nameKeyCol() string {
	return customerTypeName
}

func (t custTypeTable) columns() []string {
	return []string{
		customerTypeID,
		customerTypeName,
	}
}

func (t custTypeTable) refTables() []table { return nil }

func (t statusTable) name() string {
	return statusTableName
}

func (t statusTable) primaryKeyCol() string {
	return statusID
}

func (t statusTable) secondaryKeyCol() string { return "" }

func (t statusTable) nameKeyCol() string {
	return statusName
}

func (t statusTable) columns() []string {
	return []string{
		statusID,
		statusName,
	}
}

func (t statusTable) refTables() []table { return nil }

func (t purchStrCodeTable) name() string {
	return purchaseStringCodeTableName
}

func (t purchStrCodeTable) primaryKeyCol() string {
	return purchaseStringCode
}

func (t purchStrCodeTable) secondaryKeyCol() string { return "" }

func (t purchStrCodeTable) nameKeyCol() string {
	return purchaseStringCodeName
}

func (t purchStrCodeTable) columns() []string {
	return []string{
		purchaseStringCode,
		purchaseStringCodeName,
	}
}

func (t purchStrCodeTable) refTables() []table { return nil }

func (t purchaseTable) name() string {
	return purchTableName
}

func (t purchaseTable) primaryKeyCol() string {
	return registryNumber
}

func (t purchaseTable) secondaryKeyCol() string { return purchaseID }

func (t purchaseTable) nameKeyCol() string {
	return purchaseSubject
}

func (t purchaseTable) refTables() []table {
	return []table{
		&etpTable{},
		&purchStrCodeTable{},
		&custTypeTable{},
		&regionTable{},
		&purchTypeTable{},
		&statusTable{},
	}
}

func (t purchaseTable) columns() []string {
	return []string{
		//this is generated by DB
		//purchaseID,
		registryNumber,
		purchaseSubject,
		purchaseABBR,
		purchaseTypeColumn,
		collectingColumn,
		approvalColumn,
		biddingColumn,
		regionColumn,
		customerTypeColumn,
		maxPrice,
		applicationGuarantee,
		contractGuarantee,
		statusColumn,
		ourParticipants,
		estimationColumn,
		etpColumn,
		winnerColumn,
		winnerPrice,
		participantsColumn,
	}
}
