package botDB

import (
	"fmt"
	"strings"
)

// opts is parameters needed to build
// query statements
type opts struct {
	tableName  string
	primaryKey string
	multiplier int
	withUpdate bool
}

func upsertStatement(opts opts, cols []string) string {
	if len(cols) == 0 {
		return ""
	}

	if opts.withUpdate {
		return fmt.Sprintf("insert into %s (%s) values %s on conflict (%s) do update %s;",
			opts.tableName, columns(cols), placeholders(len(cols), opts.multiplier), opts.primaryKey, excluded(cols))
	}
	return fmt.Sprintf("insert into %s (%s) values %s on conflict (%s) do nothing;",
		opts.tableName, columns(cols), placeholders(len(cols), opts.multiplier), opts.primaryKey)
}

func insertStatement(table string, cols []string) string {
	if len(cols) == 0 {
		return ""
	}
	return fmt.Sprintf("insert into %s (%s) values %s;", table, columns(cols), placeholders(len(cols), 1))
}

func idStatement(tableName, nameCol, idCol, name string) string {
	return fmt.Sprintf("select %s from %s where %s = %s;", idCol, tableName, nameCol, name)
}

func idNameStatement(tableName, idCol, nameCol string) string {
	return fmt.Sprintf("select %s, %s from %s;", idCol, nameCol, tableName)
}

func placeholders(count, multiplier int) string {
	var b strings.Builder
	n := 1

	for i := 0; i < multiplier; i++ {

		b.WriteString("(")

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
			b.WriteString(")")
		}
	}

	return b.String()
}

func columns(cols []string) string {
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
