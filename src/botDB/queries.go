package botDB

import (
	"fmt"
	"strings"
)

// stmtOpts is parameters needed to build
// query statements
type stmtOpts struct {
	tableName   string
	conflictKey string // the key which we expect to conflict when upserting
	fromClause  string
	whereClause string
	groupBy     []string // group by columns
	orderBy     []string // order by columns
	limit       int
	cols        []string
	multiplier  int
	withUpdate  bool
}

// joinOpt is the parameter
// needed to alter FROM clause building process
type joinOpt int

// join option parameter
const (
	_ joinOpt = iota
	inner
	left
	right
)

func upsertStatement(opts stmtOpts) string {
	if len(opts.cols) == 0 {
		return ""
	}

	// do update or do nothing when conflict is encountered
	if opts.withUpdate {
		return fmt.Sprintf("insert into %s (%s) values %s on conflict (%s) do update %s;",
			opts.tableName, columns(opts.cols...), placeholders(len(opts.cols), opts.multiplier), opts.conflictKey, excluded(opts.cols))
	}
	return fmt.Sprintf("insert into %s (%s) values %s on conflict (%s) do nothing;",
		opts.tableName, columns(opts.cols...), placeholders(len(opts.cols), opts.multiplier), opts.conflictKey)
}

func selectWhereStmt(opts stmtOpts) string {
	var order, limit, group string
	if opts.limit > 0 {
		limit = fmt.Sprintf("limit %d", opts.limit)
	}
	if len(opts.orderBy) != 0 {
		order = fmt.Sprintf("order by %s", columns(opts.orderBy...))
	}
	if len(opts.groupBy) != 0 {
		group = fmt.Sprintf("group by %s", columns(opts.groupBy...))
	}

	if opts.fromClause != "" {
		return fmt.Sprintf("select (%s) from %s %s %s %s %s;",
			columns(opts.cols...), opts.fromClause, opts.whereClause, group, order, limit)
	}
	return fmt.Sprintf("select (%s) from %s %s %s %s %s;",
		columns(opts.cols...), opts.tableName, opts.whereClause, group, order, limit)
}

func buildFromClause(t table, join joinOpt) string {
	var b strings.Builder
	var j string
	switch join {
	case inner:
		j = "inner join"
	case left:
		j = "left join"
	case right:
		j = "right join"
	default:
		j = "inner join"
	}

	refs := t.refTables()

	b.WriteString(t.name())

	for _, rt := range refs {
		b.WriteString(fmt.Sprintf(" %s %s on %s.%s = %s.%s", j,
			rt.name(), rt.name(), rt.joinOn(t), t.name(), t.joinOn(rt)))
	}

	return b.String()
}

func placeholders(count, multiplier int) string {
	var b strings.Builder
	n := 1

	for i := 0; i < multiplier; i++ {

		b.WriteRune('(')

		for j := 0; j < count; j++ {

			if j == 0 {
				b.WriteString(fmt.Sprintf("$%d", n))
				n++
				continue
			}

			b.WriteString(fmt.Sprintf(", $%d", n))
			n++
		}

		if i != multiplier-1 {
			b.WriteString("), ")
		} else {
			b.WriteRune(')')
		}
	}

	return b.String()
}

func columns(cols ...string) string {
	var b strings.Builder
	for i := range cols {
		if i == 0 {
			b.WriteString(cols[i])
			continue
		}
		b.WriteString(", ")
		b.WriteString(cols[i])
	}
	return b.String()
}

func excluded(cols []string) string {
	var b strings.Builder
	for i := range cols {
		if i == 0 {
			b.WriteString(fmt.Sprintf("set %s = excluded.%s", cols[i], cols[i]))
			continue
		}
		b.WriteString(fmt.Sprintf(", %s = excluded.%s", cols[i], cols[i]))
	}
	return b.String()
}
