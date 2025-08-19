package feature

import "time"

type SchemaBuilder interface {
	// Управление таблицами
	CreateTable(name string, callback func(TableBuilder)) error
	AlterTable(name string, callback func(TableBuilder)) error
	DropTable(name string) error
	RenameTable(from, to string) error

	// Проверки существования
	HasTable(name string) (bool, error)
	HasColumn(table, column string) (bool, error)
	HasIndex(table, index string) (bool, error)

	// Информация о схеме
	GetTables() ([]TableInfo, error)
	GetColumns(table string) ([]ColumnInfo, error)
	GetIndexes(table string) ([]IndexInfo, error)
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
	Index(columns ...string) TableBuilder
	UniqueIndex(columns ...string) TableBuilder
	PrimaryKey(columns ...string) TableBuilder
	ForeignKey(column, refTable, refColumn string) TableBuilder

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
type MigrationManager interface {
	// Выполнение миграций
	Up() error
	Down() error
	Migrate(steps ...int) error
	Rollback(steps ...int) error

	// Статус
	Status() ([]MigrationStatus, error)
	Current() (string, error)

	// Создание миграций
	Create(name string) error

	// Сброс
	Reset() error
	Refresh() error
}

type Migration interface {
	Up(schema SchemaBuilder) error
	Down(schema SchemaBuilder) error
}
type MigrationStatus struct {
	Name      string
	Applied   bool
	AppliedAt *time.Time
	Batch     int
}
