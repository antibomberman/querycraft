package dialect

import (
	"fmt"
	"strings"
)

type MySQLDialect struct{}

func (d *MySQLDialect) PlaceholderFormat() string {
	return "?"
}

func (d *MySQLDialect) Rebind(sql string) string {
	return sql // MySQL uses ? placeholders, no rebinding needed
}

func (d *MySQLDialect) SelectLimit(limit int) string {
	return fmt.Sprintf("LIMIT %d", limit)
}

func (d *MySQLDialect) SelectOffset(offset int) string {
	return fmt.Sprintf("OFFSET %d", offset)
}

func (d *MySQLDialect) SelectOrderBy(column string, desc bool) string {
	if desc {
		return fmt.Sprintf("ORDER BY %s DESC", d.QuoteIdentifier(column))
	}
	return fmt.Sprintf("ORDER BY %s", d.QuoteIdentifier(column))
}

func (d *MySQLDialect) InsertIgnore() string {
	return "INSERT IGNORE"
}

func (d *MySQLDialect) InsertReplace() string {
	return "REPLACE"
}

func (d *MySQLDialect) InsertOnConflict(columns []string, updateColumns []string, updateExcluded []string) string {
	if len(updateColumns) == 0 && len(updateExcluded) == 0 {
		return "ON DUPLICATE KEY UPDATE"
	}

	var updates []string
	for _, col := range updateColumns {
		updates = append(updates, fmt.Sprintf("%s = VALUES(%s)", d.QuoteIdentifier(col), d.QuoteIdentifier(col)))
	}

	for _, col := range updateExcluded {
		updates = append(updates, fmt.Sprintf("%s = VALUES(%s)", d.QuoteIdentifier(col), d.QuoteIdentifier(col)))
	}

	return fmt.Sprintf("ON DUPLICATE KEY UPDATE %s", strings.Join(updates, ", "))
}

func (d *MySQLDialect) UpdateLimit(limit int) string {
	return fmt.Sprintf("LIMIT %d", limit)
}

func (d *MySQLDialect) DeleteLimit(limit int) string {
	return fmt.Sprintf("LIMIT %d", limit)
}

func (d *MySQLDialect) Upsert(columns []string, values []any, conflictColumns []string, updateColumns []string) (string, []any) {
	// For MySQL, this is implemented as INSERT ... ON DUPLICATE KEY UPDATE
	// This will be handled in the UpsertBuilder implementation
	return "", nil
}

func (d *MySQLDialect) BulkInsert(table string, columns []string, values []any, batchSize int) (string, []any) {
	// This will be handled in the BulkBuilder implementation
	return "", nil
}

func (d *MySQLDialect) BulkUpdate(table string, columns []string, values []any, keyColumn string) (string, []any) {
	// This will be handled in the BulkBuilder implementation
	return "", nil
}

func (d *MySQLDialect) BulkDelete(table string, conditions []map[string]any) (string, []any) {
	if len(conditions) == 0 {
		return "", nil
	}

	// Build WHERE clauses for each condition
	var whereClauses []string
	var args []any

	for _, condition := range conditions {
		var parts []string
		for column, value := range condition {
			parts = append(parts, fmt.Sprintf("%s = %s", d.QuoteIdentifier(column), d.PlaceholderFormat()))
			args = append(args, value)
		}
		whereClauses = append(whereClauses, fmt.Sprintf("(%s)", strings.Join(parts, " AND ")))
	}

	whereClause := strings.Join(whereClauses, " OR ")
	query := fmt.Sprintf("DELETE FROM %s WHERE %s", d.QuoteIdentifier(table), whereClause)

	return query, args
}

func (d *MySQLDialect) QuoteIdentifier(name string) string {
	return fmt.Sprintf("`%s`", strings.ReplaceAll(name, "`", "``"))
}
