package querycraft

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/antibomberman/querycraft/dialect"
)

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

	// Работа с миграциями
	RegisterMigration(name string, migration Migration) error
	GetMigrations() map[string]Migration
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

type migrationManager struct {
	db         SQLXExecutor
	dialect    dialect.Dialect
	migrations map[string]Migration
	ctx        context.Context
}

func NewMigrationManager(db SQLXExecutor, dialect dialect.Dialect) MigrationManager {
	return &migrationManager{
		db:         db,
		dialect:    dialect,
		migrations: make(map[string]Migration),
		ctx:        context.Background(),
	}
}

func (m *migrationManager) WithContext(ctx context.Context) MigrationManager {
	m.ctx = ctx
	return m
}

// Регистрация миграций
func (m *migrationManager) RegisterMigration(name string, migration Migration) error {
	m.migrations[name] = migration
	return nil
}

func (m *migrationManager) GetMigrations() map[string]Migration {
	return m.migrations
}

// Выполнение миграций
func (m *migrationManager) Up() error {
	// Создаем таблицу миграций, если она не существует
	if err := m.createMigrationsTable(); err != nil {
		return err
	}

	// Получаем список уже примененных миграций
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	// Определяем следующий batch номер
	batch, err := m.getNextBatchNumber()
	if err != nil {
		return err
	}

	// Применяем непримененные миграции
	for name, migration := range m.migrations {
		if !contains(applied, name) {
			schema := NewSchemaBuilder(m.db, m.dialect)
			if err := migration.Up(schema); err != nil {
				return fmt.Errorf("error applying migration %s: %w", name, err)
			}

			// Записываем в таблицу миграций
			if err := m.recordMigration(name, batch); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *migrationManager) Down() error {
	// Получаем список всех примененных миграций
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	// Откатываем миграции в обратном порядке
	for i := len(applied) - 1; i >= 0; i-- {
		name := applied[i]
		migration, exists := m.migrations[name]
		if !exists {
			return fmt.Errorf("migration %s not found", name)
		}

		schema := NewSchemaBuilder(m.db, m.dialect)
		if err := migration.Down(schema); err != nil {
			return fmt.Errorf("error rolling back migration %s: %w", name, err)
		}

		// Удаляем запись из таблицы миграций
		if err := m.removeMigrationRecord(name); err != nil {
			return err
		}
	}

	return nil
}

func (m *migrationManager) Migrate(steps ...int) error {
	stepCount := 1
	if len(steps) > 0 {
		stepCount = steps[0]
	}

	// Создаем таблицу миграций, если она не существует
	if err := m.createMigrationsTable(); err != nil {
		return err
	}

	// Получаем список уже примененных миграций
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	// Определяем следующий batch номер
	batch, err := m.getNextBatchNumber()
	if err != nil {
		return err
	}

	// Применяем непримененные миграции
	appliedCount := 0
	for name, migration := range m.migrations {
		if !contains(applied, name) && appliedCount < stepCount {
			schema := NewSchemaBuilder(m.db, m.dialect)
			if err := migration.Up(schema); err != nil {
				return fmt.Errorf("error applying migration %s: %w", name, err)
			}

			// Записываем в таблицу миграций
			if err := m.recordMigration(name, batch); err != nil {
				return err
			}

			appliedCount++
		}
	}

	return nil
}

func (m *migrationManager) Rollback(steps ...int) error {
	stepCount := 1
	if len(steps) > 0 {
		stepCount = steps[0]
	}

	// Получаем список последних примененных миграций по batch номеру
	lastBatch, err := m.getLastBatchNumber()
	if err != nil {
		return err
	}

	applied, err := m.getAppliedMigrationsByBatch(lastBatch)
	if err != nil {
		return err
	}

	// Откатываем миграции в обратном порядке
	rollbackCount := 0
	for i := len(applied) - 1; i >= 0 && rollbackCount < stepCount; i-- {
		name := applied[i]
		migration, exists := m.migrations[name]
		if !exists {
			return fmt.Errorf("migration %s not found", name)
		}

		schema := NewSchemaBuilder(m.db, m.dialect)
		if err := migration.Down(schema); err != nil {
			return fmt.Errorf("error rolling back migration %s: %w", name, err)
		}

		// Удаляем запись из таблицы миграций
		if err := m.removeMigrationRecord(name); err != nil {
			return err
		}

		rollbackCount++
	}

	return nil
}

// Статус
func (m *migrationManager) Status() ([]MigrationStatus, error) {
	// Создаем таблицу миграций, если она не существует
	if err := m.createMigrationsTable(); err != nil {
		return nil, err
	}

	// Получаем список уже примененных миграций
	applied, err := m.getMigrationStatuses()
	if err != nil {
		return nil, err
	}

	// Создаем статусы для всех миграций
	var statuses []MigrationStatus
	for name := range m.migrations {
		status := MigrationStatus{
			Name:    name,
			Applied: false,
		}

		// Проверяем, применена ли миграция
		for _, appliedStatus := range applied {
			if appliedStatus.Name == name {
				status.Applied = true
				status.AppliedAt = appliedStatus.AppliedAt
				status.Batch = appliedStatus.Batch
				break
			}
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

func (m *migrationManager) Current() (string, error) {
	lastBatch, err := m.getLastBatchNumber()
	if err != nil {
		return "", err
	}

	applied, err := m.getAppliedMigrationsByBatch(lastBatch)
	if err != nil {
		return "", err
	}

	if len(applied) > 0 {
		return applied[len(applied)-1], nil
	}

	return "", nil
}

// Создание миграций
func (m *migrationManager) Create(name string) error {
	// Эта функция должна создавать файлы миграций, но для простоты
	// мы просто возвращаем nil
	return nil
}

// Сброс
func (m *migrationManager) Reset() error {
	// Откатываем все миграции
	if err := m.Down(); err != nil {
		return err
	}

	// Удаляем таблицу миграций
	if err := m.dropMigrationsTable(); err != nil {
		return err
	}

	return nil
}

func (m *migrationManager) Refresh() error {
	// Сброс и повторное применение всех миграций
	if err := m.Reset(); err != nil {
		return err
	}

	if err := m.Up(); err != nil {
		return err
	}

	return nil
}

// Вспомогательные методы
func (m *migrationManager) createMigrationsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS migrations (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL UNIQUE,
		batch INT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := m.db.ExecContext(m.ctx, query)
	return err
}

func (m *migrationManager) dropMigrationsTable() error {
	query := "DROP TABLE IF EXISTS migrations"
	_, err := m.db.ExecContext(m.ctx, query)
	return err
}

func (m *migrationManager) getAppliedMigrations() ([]string, error) {
	query := "SELECT name FROM migrations ORDER BY created_at ASC"
	rows, err := m.db.QueryContext(m.ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	return names, nil
}

func (m *migrationManager) getAppliedMigrationsByBatch(batch int) ([]string, error) {
	query := "SELECT name FROM migrations WHERE batch = ? ORDER BY created_at ASC"
	rows, err := m.db.QueryContext(m.ctx, query, batch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	return names, nil
}

func (m *migrationManager) getMigrationStatuses() ([]MigrationStatus, error) {
	query := "SELECT name, batch, created_at FROM migrations ORDER BY created_at ASC"
	rows, err := m.db.QueryContext(m.ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []MigrationStatus
	for rows.Next() {
		var status MigrationStatus
		var appliedAt time.Time
		if err := rows.Scan(&status.Name, &status.Batch, &appliedAt); err != nil {
			return nil, err
		}
		status.Applied = true
		status.AppliedAt = &appliedAt
		statuses = append(statuses, status)
	}

	return statuses, nil
}

func (m *migrationManager) getNextBatchNumber() (int, error) {
	var batch int
	query := "SELECT COALESCE(MAX(batch), 0) + 1 FROM migrations"
	err := m.db.GetContext(m.ctx, &batch, query)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return batch, nil
}

func (m *migrationManager) getLastBatchNumber() (int, error) {
	var batch int
	query := "SELECT COALESCE(MAX(batch), 0) FROM migrations"
	err := m.db.GetContext(m.ctx, &batch, query)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return batch, nil
}

func (m *migrationManager) recordMigration(name string, batch int) error {
	query := "INSERT INTO migrations (name, batch) VALUES (?, ?)"
	_, err := m.db.ExecContext(m.ctx, query, name, batch)
	return err
}

func (m *migrationManager) removeMigrationRecord(name string) error {
	query := "DELETE FROM migrations WHERE name = ?"
	_, err := m.db.ExecContext(m.ctx, query, name)
	return err
}

// Вспомогательная функция для проверки наличия элемента в срезе
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
