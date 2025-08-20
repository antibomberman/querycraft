package querycraft

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/antibomberman/querycraft/dialect"
)

type TableInfo struct {
	Name string
}

type ColumnInfo struct {
	Name string
	Type string
}

type IndexInfo struct {
	Name    string
	Columns []string
}

type SchemaBuilder interface {
	// Управление таблицами
	CreateTable(name string, callback func(TableBuilder)) error
	AlterTable(name string, callback func(TableBuilder)) error
	DropTable(name string) error
	RenameTable(from, to string) error
	ClearTable(table string) error

	// Проверки существования
	HasTable(name string) (bool, error)
	HasColumn(table, column string) (bool, error)
	HasIndex(table, index string) (bool, error)

	// Информация о схеме
	GetTables() ([]TableInfo, error)
	GetColumns(table string) ([]ColumnInfo, error)
	GetIndexes(table string) ([]IndexInfo, error)

	WithContext(ctx context.Context) SchemaBuilder
}
type TableBuilder interface {
	// Колонки
	ID() TableBuilder                                        // auto increment primary key
	String(name string, length ...int) ColumnBuilder         // VARCHAR
	Text(name string) ColumnBuilder                          // TEXT
	Integer(name string) ColumnBuilder                       // INT
	BigInteger(name string) ColumnBuilder                    // BIGINT
	Decimal(name string, precision, scale int) ColumnBuilder // DECIMAL
	Boolean(name string) ColumnBuilder                       // BOOLEAN
	Date(name string) ColumnBuilder                          // DATE
	DateTime(name string) ColumnBuilder                      // DATETIME
	Timestamp(name string) ColumnBuilder                     // TIMESTAMP
	JSON(name string) ColumnBuilder                          // JSON

	// Специальные колонки
	Timestamps() TableBuilder  // created_at, updated_at
	SoftDeletes() TableBuilder // deleted_at

	// Индексы
	AddIndex(name string, columns ...string) TableBuilder
	UniqueIndex(columns ...string) TableBuilder
	PrimaryKey(columns ...string) TableBuilder
	ForeignKey(column, refTable, refColumn string) TableBuilder
	//HasIndex(table string, index string) (bool, error)

	// Удаление
	DropColumn(name string) TableBuilder
	DropIndex(name string) TableBuilder
	DropForeign(name string) TableBuilder
}

type ColumnBuilder interface {
	Nullable() ColumnBuilder
	NotNull() ColumnBuilder
	Default(value any) ColumnBuilder
	Unique() ColumnBuilder
	Index() ColumnBuilder
	Primary() ColumnBuilder
	AutoIncrement() ColumnBuilder
	Comment(comment string) ColumnBuilder
	After(column string) ColumnBuilder // MySQL
	First() ColumnBuilder              // MySQL
}

type schemaBuilder struct {
	db      SQLXExecutor
	dialect dialect.Dialect
	ctx     context.Context
}

func NewSchemaBuilder(db SQLXExecutor, dialect dialect.Dialect) SchemaBuilder {
	return &schemaBuilder{
		db:      db,
		dialect: dialect,
		ctx:     context.Background(),
	}
}

func (s *schemaBuilder) WithContext(ctx context.Context) SchemaBuilder {
	s.ctx = ctx
	return s
}

// Управление таблицами
func (s *schemaBuilder) CreateTable(name string, callback func(TableBuilder)) error {
	builder := newTableBuilder(s.db, s.dialect, name)
	callback(builder)

	query, args := builder.toSQL()
	_, err := s.db.ExecContext(s.ctx, query, args...)
	return err
}

func (s *schemaBuilder) AlterTable(name string, callback func(TableBuilder)) error {
	builder := newTableBuilder(s.db, s.dialect, name)
	builder.alter = true
	callback(builder)

	query, args := builder.toSQL()
	_, err := s.db.ExecContext(s.ctx, query, args...)
	return err
}

func (s *schemaBuilder) DropTable(name string) error {
	query := fmt.Sprintf("DROP TABLE %s", s.dialect.QuoteIdentifier(name))
	_, err := s.db.ExecContext(s.ctx, query)
	return err
}

func (s *schemaBuilder) RenameTable(from, to string) error {
	query := fmt.Sprintf("ALTER TABLE %s RENAME TO %s",
		s.dialect.QuoteIdentifier(from),
		s.dialect.QuoteIdentifier(to))
	_, err := s.db.ExecContext(s.ctx, query)
	return err
}

// Проверки существования
func (s *schemaBuilder) HasTable(name string) (bool, error) {
	query := s.dialect.HasTableQuery(name)

	// Используем map для получения результата, так как выражение COUNT(*) > 0 возвращает столбец без имени
	rows, err := s.db.QueryContext(s.ctx, query)
	if err != nil {
		// Если таблица не существует, некоторые драйверы могут вернуть ошибку
		// В этом случае мы возвращаем false, nil
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	defer rows.Close()

	// Получаем первую строку
	if rows.Next() {
		// Создаем слайс для значений
		values, err := rows.Columns()
		if err != nil {
			return false, err
		}

		// Создаем slice для значений
		valuePtrs := make([]any, len(values))
		valueScan := make([]any, len(values))
		for i := range values {
			valuePtrs[i] = &valueScan[i]
		}

		// Сканируем значения
		if err := rows.Scan(valuePtrs...); err != nil {
			return false, err
		}

		// Преобразуем результат в bool
		if len(valueScan) > 0 {
			// Проверяем тип значения
			switch v := valueScan[0].(type) {
			case int64:
				return v > 0, nil
			case bool:
				return v, nil
			default:
				// Если это другое значение, пытаемся преобразовать в bool
				return v != nil, nil
			}
		}
	}

	return false, nil
}

func (s *schemaBuilder) HasColumn(table, column string) (bool, error) {
	query := s.dialect.HasColumnQuery(table, column)

	// Используем map для получения результата, так как выражение COUNT(*) > 0 возвращает столбец без имени
	rows, err := s.db.QueryContext(s.ctx, query)
	if err != nil {
		// Если столбец не существует, некоторые драйверы могут вернуть ошибку
		// В этом случае мы возвращаем false, nil
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	defer rows.Close()

	// Получаем первую строку
	if rows.Next() {
		// Создаем слайс для значений
		values, err := rows.Columns()
		if err != nil {
			return false, err
		}

		// Создаем slice для значений
		valuePtrs := make([]any, len(values))
		valueScan := make([]any, len(values))
		for i := range values {
			valuePtrs[i] = &valueScan[i]
		}

		// Сканируем значения
		if err := rows.Scan(valuePtrs...); err != nil {
			return false, err
		}

		// Преобразуем результат в bool
		if len(valueScan) > 0 {
			// Проверяем тип значения
			switch v := valueScan[0].(type) {
			case int64:
				return v > 0, nil
			case bool:
				return v, nil
			default:
				// Если это другое значение, пытаемся преобразовать в bool
				return v != nil, nil
			}
		}
	}

	return false, nil
}

func (s *schemaBuilder) HasIndex(table, index string) (bool, error) {
	query := s.dialect.HasIndexQuery(table, index)

	// Используем map для получения результата, так как выражение COUNT(*) > 0 возвращает столбец без имени
	rows, err := s.db.QueryContext(s.ctx, query)
	if err != nil {
		// Если индекс не существует, некоторые драйверы могут вернуть ошибку
		// В этом случае мы возвращаем false, nil
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	defer rows.Close()

	// Получаем первую строку
	if rows.Next() {
		// Создаем слайс для значений
		values, err := rows.Columns()
		if err != nil {
			return false, err
		}

		// Создаем slice для значений
		valuePtrs := make([]any, len(values))
		valueScan := make([]any, len(values))
		for i := range values {
			valuePtrs[i] = &valueScan[i]
		}

		// Сканируем значения
		if err := rows.Scan(valuePtrs...); err != nil {
			return false, err
		}

		// Преобразуем результат в bool
		if len(valueScan) > 0 {
			// Проверяем тип значения
			switch v := valueScan[0].(type) {
			case int64:
				return v > 0, nil
			case bool:
				return v, nil
			default:
				// Если это другое значение, пытаемся преобразовать в bool
				return v != nil, nil
			}
		}
	}

	return false, nil
}

// Информация о схеме
func (s *schemaBuilder) GetTables() ([]TableInfo, error) {
	query := s.dialect.GetTablesQuery()

	rows, err := s.db.QueryContext(s.ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var table TableInfo
		if err := rows.Scan(&table.Name); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

func (s *schemaBuilder) GetColumns(table string) ([]ColumnInfo, error) {
	query := s.dialect.GetColumnsQuery(table)

	rows, err := s.db.QueryContext(s.ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var column ColumnInfo
		if err := rows.Scan(&column.Name, &column.Type); err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
}

func (s *schemaBuilder) GetIndexes(table string) ([]IndexInfo, error) {
	query := s.dialect.GetIndexesQuery(table)

	rows, err := s.db.QueryContext(s.ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []IndexInfo
	for rows.Next() {
		var index IndexInfo
		var columnsStr string

		if err := rows.Scan(&index.Name, &columnsStr); err != nil {
			return nil, err
		}

		// Разбиваем строку с колонками на срез
		if columnsStr != "" {
			index.Columns = strings.Split(columnsStr, ",")
			// Удаляем пробелы вокруг имен колонок
			for i, col := range index.Columns {
				index.Columns[i] = strings.TrimSpace(col)
			}
		}

		indexes = append(indexes, index)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return indexes, nil
}
func (s *schemaBuilder) ClearTable(table string) error {
	query := s.dialect.TruncateTableSQL(table)
	_, err := s.db.ExecContext(s.ctx, query)
	return err
}

// TableBuilder implementation
type tableBuilder struct {
	db      SQLXExecutor
	dialect dialect.Dialect

	tableName string
	columns   []columnDefinition
	indexes   []indexDefinition
	commands  []string

	alter bool // true for ALTER TABLE, false for CREATE TABLE
}

type columnDefinition struct {
	name      string
	dataType  string
	modifiers []string
	after     string
	first     bool
}

type indexDefinition struct {
	name    string
	columns []string
	unique  bool
	primary bool
	foreign *foreignDefinition
}

type foreignDefinition struct {
	column    string
	refTable  string
	refColumn string
	onDelete  string
	onUpdate  string
}

func newTableBuilder(db SQLXExecutor, dialect dialect.Dialect, tableName string) *tableBuilder {
	return &tableBuilder{
		db:        db,
		dialect:   dialect,
		tableName: tableName,
		columns:   make([]columnDefinition, 0),
		indexes:   make([]indexDefinition, 0),
		commands:  make([]string, 0),
	}
}

// Колонки
func (t *tableBuilder) ID() TableBuilder {
	t.columns = append(t.columns, columnDefinition{
		name:      "id",
		dataType:  t.dialect.GetIDColumnType(),
		modifiers: []string{"PRIMARY KEY", "AUTO_INCREMENT"},
	})
	return t
}

func (t *tableBuilder) String(name string, length ...int) ColumnBuilder {
	l := 255
	if len(length) > 0 && length[0] > 0 {
		l = length[0]
	}

	t.columns = append(t.columns, columnDefinition{
		name:     name,
		dataType: fmt.Sprintf("VARCHAR(%d)", l),
	})
	return t
}

func (t *tableBuilder) Text(name string) ColumnBuilder {
	t.columns = append(t.columns, columnDefinition{
		name:     name,
		dataType: "TEXT",
	})
	return t
}

func (t *tableBuilder) Integer(name string) ColumnBuilder {
	t.columns = append(t.columns, columnDefinition{
		name:     name,
		dataType: "INT",
	})
	return t
}

func (t *tableBuilder) BigInteger(name string) ColumnBuilder {
	t.columns = append(t.columns, columnDefinition{
		name:     name,
		dataType: "BIGINT",
	})
	return t
}

func (t *tableBuilder) Decimal(name string, precision, scale int) ColumnBuilder {
	t.columns = append(t.columns, columnDefinition{
		name:     name,
		dataType: fmt.Sprintf("DECIMAL(%d, %d)", precision, scale),
	})
	return t
}

func (t *tableBuilder) Boolean(name string) ColumnBuilder {
	t.columns = append(t.columns, columnDefinition{
		name:     name,
		dataType: "BOOLEAN",
	})
	return t
}

func (t *tableBuilder) Date(name string) ColumnBuilder {
	t.columns = append(t.columns, columnDefinition{
		name:     name,
		dataType: "DATE",
	})
	return t
}

func (t *tableBuilder) DateTime(name string) ColumnBuilder {
	t.columns = append(t.columns, columnDefinition{
		name:     name,
		dataType: "DATETIME",
	})
	return t
}

func (t *tableBuilder) Timestamp(name string) ColumnBuilder {
	t.columns = append(t.columns, columnDefinition{
		name:     name,
		dataType: "TIMESTAMP",
	})
	return t
}

func (t *tableBuilder) JSON(name string) ColumnBuilder {
	t.columns = append(t.columns, columnDefinition{
		name:     name,
		dataType: "JSON",
	})
	return t
}

// Специальные колонки
func (t *tableBuilder) Timestamps() TableBuilder {
	t.DateTime("created_at").NotNull()
	t.DateTime("updated_at").NotNull()
	return t
}

func (t *tableBuilder) SoftDeletes() TableBuilder {
	t.DateTime("deleted_at").Nullable()
	return t
}

// Модификаторы колонок
func (t *tableBuilder) Nullable() ColumnBuilder {
	if len(t.columns) > 0 {
		t.columns[len(t.columns)-1].modifiers = append(t.columns[len(t.columns)-1].modifiers, "NULL")
	}
	return t
}

func (t *tableBuilder) NotNull() ColumnBuilder {
	if len(t.columns) > 0 {
		t.columns[len(t.columns)-1].modifiers = append(t.columns[len(t.columns)-1].modifiers, "NOT NULL")
	}
	return t
}

func (t *tableBuilder) Default(value any) ColumnBuilder {
	if len(t.columns) > 0 {
		var defaultValue string
		switch v := value.(type) {
		case string:
			defaultValue = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
		case bool:
			if v {
				defaultValue = "true"
			} else {
				defaultValue = "false"
			}
		default:
			defaultValue = fmt.Sprintf("%v", v)
		}
		t.columns[len(t.columns)-1].modifiers = append(t.columns[len(t.columns)-1].modifiers, fmt.Sprintf("DEFAULT %s", defaultValue))
	}
	return t
}

func (t *tableBuilder) Unique() ColumnBuilder {
	if len(t.columns) > 0 {
		columnName := t.columns[len(t.columns)-1].name
		t.indexes = append(t.indexes, indexDefinition{
			name:    fmt.Sprintf("%s_%s_unique", t.tableName, columnName),
			columns: []string{columnName},
			unique:  true,
		})
	}
	return t
}

func (t *tableBuilder) Index() ColumnBuilder {
	if len(t.columns) > 0 {
		columnName := t.columns[len(t.columns)-1].name
		t.indexes = append(t.indexes, indexDefinition{
			name:    fmt.Sprintf("%s_%s_index", t.tableName, columnName),
			columns: []string{columnName},
			unique:  false,
		})
	}
	return t
}

func (t *tableBuilder) Primary() ColumnBuilder {
	if len(t.columns) > 0 {
		columnName := t.columns[len(t.columns)-1].name
		t.indexes = append(t.indexes, indexDefinition{
			name:    fmt.Sprintf("%s_%s_primary", t.tableName, columnName),
			columns: []string{columnName},
			primary: true,
		})
	}
	return t
}

func (t *tableBuilder) AutoIncrement() ColumnBuilder {
	if len(t.columns) > 0 {
		t.columns[len(t.columns)-1].modifiers = append(t.columns[len(t.columns)-1].modifiers, "AUTO_INCREMENT")
	}
	return t
}

func (t *tableBuilder) Comment(comment string) ColumnBuilder {
	// Comments are typically handled differently in different databases
	// For now, we'll just ignore this in the SQL generation
	return t
}

func (t *tableBuilder) After(column string) ColumnBuilder {
	if len(t.columns) > 0 {
		t.columns[len(t.columns)-1].after = column
	}
	return t
}

func (t *tableBuilder) First() ColumnBuilder {
	if len(t.columns) > 0 {
		t.columns[len(t.columns)-1].first = true
	}
	return t
}

// Индексы
func (t *tableBuilder) AddIndex(name string, columns ...string) TableBuilder {
	if len(columns) > 0 {
		t.indexes = append(t.indexes, indexDefinition{
			name:    name,
			columns: columns,
			unique:  false,
		})
	}
	return t
}

func (t *tableBuilder) UniqueIndex(columns ...string) TableBuilder {
	if len(columns) > 0 {
		indexName := fmt.Sprintf("%s_%s_unique", t.tableName, strings.Join(columns, "_"))
		t.indexes = append(t.indexes, indexDefinition{
			name:    indexName,
			columns: columns,
			unique:  true,
		})
	}
	return t
}

func (t *tableBuilder) PrimaryKey(columns ...string) TableBuilder {
	if len(columns) > 0 {
		indexName := fmt.Sprintf("%s_%s_primary", t.tableName, strings.Join(columns, "_"))
		t.indexes = append(t.indexes, indexDefinition{
			name:    indexName,
			columns: columns,
			primary: true,
		})
	}
	return t
}

func (t *tableBuilder) ForeignKey(column, refTable, refColumn string) TableBuilder {
	foreign := &foreignDefinition{
		column:    column,
		refTable:  refTable,
		refColumn: refColumn,
	}

	t.indexes = append(t.indexes, indexDefinition{
		name:    fmt.Sprintf("%s_%s_foreign", t.tableName, column),
		columns: []string{column},
		foreign: foreign,
	})

	return t
}

// Удаление
func (t *tableBuilder) DropColumn(name string) TableBuilder {
	if t.alter {
		t.commands = append(t.commands, fmt.Sprintf("DROP COLUMN %s", t.dialect.QuoteIdentifier(name)))
	}
	return t
}

func (t *tableBuilder) DropIndex(name string) TableBuilder {
	if t.alter {
		t.commands = append(t.commands, fmt.Sprintf("DROP INDEX %s", t.dialect.QuoteIdentifier(name)))
	}
	return t
}

func (t *tableBuilder) DropForeign(name string) TableBuilder {
	if t.alter {
		t.commands = append(t.commands, fmt.Sprintf("DROP FOREIGN KEY %s", t.dialect.QuoteIdentifier(name)))
	}
	return t
}

// Generate SQL
func (t *tableBuilder) toSQL() (string, []any) {
	if t.alter {
		return t.toAlterSQL()
	}
	return t.toCreateSQL()
}

func (t *tableBuilder) toCreateSQL() (string, []any) {
	var columnDefs []string
	var args []any

	for _, col := range t.columns {
		def := t.dialect.QuoteIdentifier(col.name) + " " + col.dataType
		if len(col.modifiers) > 0 {
			def += " " + strings.Join(col.modifiers, " ")
		}
		columnDefs = append(columnDefs, def)
	}

	// Add index definitions
	for _, idx := range t.indexes {
		if idx.primary {
			columnDefs = append(columnDefs, fmt.Sprintf("PRIMARY KEY (%s)",
				strings.Join(quoteIdentifiers(t.dialect, idx.columns), ", ")))
		} else if idx.unique {
			columnDefs = append(columnDefs, fmt.Sprintf("UNIQUE KEY %s (%s)",
				t.dialect.QuoteIdentifier(idx.name),
				strings.Join(quoteIdentifiers(t.dialect, idx.columns), ", ")))
		}
	}

	// Add foreign key constraints
	for _, idx := range t.indexes {
		if idx.foreign != nil {
			columnDefs = append(columnDefs, fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s)",
				t.dialect.QuoteIdentifier(idx.foreign.column),
				t.dialect.QuoteIdentifier(idx.foreign.refTable),
				t.dialect.QuoteIdentifier(idx.foreign.refColumn)))
		}
	}

	query := fmt.Sprintf("CREATE TABLE %s (%s)",
		t.dialect.QuoteIdentifier(t.tableName),
		strings.Join(columnDefs, ", "))

	return query, args
}

func (t *tableBuilder) toAlterSQL() (string, []any) {
	var alterParts []string
	var args []any

	// Add column definitions
	for _, col := range t.columns {
		def := "ADD COLUMN " + t.dialect.QuoteIdentifier(col.name) + " " + col.dataType
		if len(col.modifiers) > 0 {
			def += " " + strings.Join(col.modifiers, " ")
		}
		alterParts = append(alterParts, def)
	}

	// Add index definitions
	for _, idx := range t.indexes {
		if idx.primary {
			alterParts = append(alterParts, fmt.Sprintf("ADD PRIMARY KEY (%s)",
				strings.Join(quoteIdentifiers(t.dialect, idx.columns), ", ")))
		} else if idx.unique {
			alterParts = append(alterParts, fmt.Sprintf("ADD UNIQUE KEY %s (%s)",
				t.dialect.QuoteIdentifier(idx.name),
				strings.Join(quoteIdentifiers(t.dialect, idx.columns), ", ")))
		} else {
			alterParts = append(alterParts, fmt.Sprintf("ADD INDEX %s (%s)",
				t.dialect.QuoteIdentifier(idx.name),
				strings.Join(quoteIdentifiers(t.dialect, idx.columns), ", ")))
		}
	}

	// Add foreign key constraints
	for _, idx := range t.indexes {
		if idx.foreign != nil {
			alterParts = append(alterParts, fmt.Sprintf("ADD FOREIGN KEY (%s) REFERENCES %s(%s)",
				t.dialect.QuoteIdentifier(idx.foreign.column),
				t.dialect.QuoteIdentifier(idx.foreign.refTable),
				t.dialect.QuoteIdentifier(idx.foreign.refColumn)))
		}
	}

	// Add commands
	for _, cmd := range t.commands {
		alterParts = append(alterParts, cmd)
	}

	query := fmt.Sprintf("ALTER TABLE %s %s", t.dialect.QuoteIdentifier(t.tableName), strings.Join(alterParts, ", "))
	return query, args
}

func quoteIdentifiers(dialect dialect.Dialect, identifiers []string) []string {
	quoted := make([]string, len(identifiers))
	for i, id := range identifiers {
		quoted[i] = dialect.QuoteIdentifier(id)
	}
	return quoted
}
