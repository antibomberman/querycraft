package querycraft

import (
	"context"
	"database/sql"
)

// RawQuery - интерфейс для сырых SQL запросов
type Raw interface {
	// Выполнение с разными результатами
	One(dest any) error
	All(dest any) error
	Row() (map[string]any, error)
	Rows() ([]map[string]any, error)
	Exec() (sql.Result, error)

	// Утилиты
	WithContext(ctx context.Context) Raw
	Args() []any
	Query() string
}

type rawQuery struct {
	db    SQLXExecutor
	ctx   context.Context
	query string
	args  []any
}

func NewRaw(db SQLXExecutor, query string, args ...any) Raw {
	return &rawQuery{
		db:    db,
		ctx:   context.Background(),
		query: query,
		args:  args,
	}
}

func (r *rawQuery) One(dest any) error {
	return r.db.GetContext(r.ctx, dest, r.query, r.args...)
}

func (r *rawQuery) All(dest any) error {
	return r.db.SelectContext(r.ctx, dest, r.query, r.args...)
}

func (r *rawQuery) Row() (map[string]any, error) {
	rows, err := r.db.QueryxContext(r.ctx, r.query, r.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		row := make(map[string]any)
		if err := rows.MapScan(row); err != nil {
			return nil, err
		}
		return row, nil
	}

	return nil, sql.ErrNoRows
}

func (r *rawQuery) Rows() ([]map[string]any, error) {
	rows, err := r.db.QueryxContext(r.ctx, r.query, r.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		row := make(map[string]any)
		if err := rows.MapScan(row); err != nil {
			return nil, err
		}
		results = append(results, row)
	}

	return results, nil
}

func (r *rawQuery) Exec() (sql.Result, error) {
	return r.db.ExecContext(r.ctx, r.query, r.args...)
}

func (r *rawQuery) WithContext(ctx context.Context) Raw {
	r.ctx = ctx
	return r
}

func (r *rawQuery) Args() []any {
	return r.args
}

func (r *rawQuery) Query() string {
	return r.query
}
