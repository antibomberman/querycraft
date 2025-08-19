package dialect

type Dialect interface {
	QuoteIdentifier(name string) string
	Placeholder(index int) string
	LimitOffset(limit, offset int) string
	OnConflictDoNothing() string
	OnConflictDoUpdate(columns []string) string
	LastInsertID() string
	DateTimeFormat() string
	SupportsReturning() bool
}
