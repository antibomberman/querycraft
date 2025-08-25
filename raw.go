package querycraft

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// RawQuery - интерфейс для сырых SQL запросов
type Raw interface {
	//struct
	One(dest any) error
	//Array struct
	All(dest any) error
	Row() (map[string]any, error)
	Rows() ([]map[string]any, error)
	Exec() (sql.Result, error)

	// Утилиты
	WithContext(ctx context.Context) Raw
	Args() []any
	Query() string
	PrintSQL() Raw
}

type rawQuery struct {
	db     SQLXExecutor
	ctx    context.Context
	query  string
	args   []any
	logger Logger

	// Print SQL flag
	printSQL bool
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
	// Print SQL if needed
	if r.printSQL {
		// Simple placeholder replacement for debugging
		formattedSQL := r.query
		for _, arg := range r.args {
			formattedSQL = strings.Replace(formattedSQL, "?", fmt.Sprintf("'%v'", arg), 1)
		}
		fmt.Println(formattedSQL)
	}

	// Log query if logger is set
	var start time.Time
	if r.logger != nil {
		start = time.Now()
	}

	err := r.db.GetContext(r.ctx, dest, r.query, r.args...)

	// Log query execution
	if r.logger != nil {
		duration := time.Since(start)
		r.logger.LogQuery(r.ctx, r.query, r.args, duration, err)
	}

	return err
}

func (r *rawQuery) All(dest any) error {
	// Print SQL if needed
	if r.printSQL {
		// Simple placeholder replacement for debugging
		formattedSQL := r.query
		for _, arg := range r.args {
			formattedSQL = strings.Replace(formattedSQL, "?", fmt.Sprintf("'%v'", arg), 1)
		}
		fmt.Println(formattedSQL)
	}

	// Log query if logger is set
	var start time.Time
	if r.logger != nil {
		start = time.Now()
	}

	err := r.db.SelectContext(r.ctx, dest, r.query, r.args...)

	// Log query execution
	if r.logger != nil {
		duration := time.Since(start)
		r.logger.LogQuery(r.ctx, r.query, r.args, duration, err)
	}

	return err
}

func (r *rawQuery) Row() (map[string]any, error) {
	// Print SQL if needed
	if r.printSQL {
		// Simple placeholder replacement for debugging
		formattedSQL := r.query
		for _, arg := range r.args {
			formattedSQL = strings.Replace(formattedSQL, "?", fmt.Sprintf("'%v'", arg), 1)
		}
		fmt.Println(formattedSQL)
	}

	// Log query if logger is set
	var start time.Time
	if r.logger != nil {
		start = time.Now()
	}

	rows, err := r.db.QueryxContext(r.ctx, r.query, r.args...)
	if err != nil {
		// Log query execution
		if r.logger != nil {
			duration := time.Since(start)
			r.logger.LogQuery(r.ctx, r.query, r.args, duration, err)
		}
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		row := make(map[string]any)
		if err := rows.MapScan(row); err != nil {
			// Log query execution
			if r.logger != nil {
				duration := time.Since(start)
				r.logger.LogQuery(r.ctx, r.query, r.args, duration, err)
			}
			return nil, err
		}

		// Log query execution
		if r.logger != nil {
			duration := time.Since(start)
			r.logger.LogQuery(r.ctx, r.query, r.args, duration, nil)
		}

		return convertByteArrayToString(row), nil
	}

	// Log query execution
	if r.logger != nil {
		duration := time.Since(start)
		r.logger.LogQuery(r.ctx, r.query, r.args, duration, sql.ErrNoRows)
	}

	return nil, sql.ErrNoRows
}

func (r *rawQuery) Rows() ([]map[string]any, error) {
	// Print SQL if needed
	if r.printSQL {
		// Simple placeholder replacement for debugging
		formattedSQL := r.query
		for _, arg := range r.args {
			formattedSQL = strings.Replace(formattedSQL, "?", fmt.Sprintf("'%v'", arg), 1)
		}
		fmt.Println(formattedSQL)
	}

	// Log query if logger is set
	var start time.Time
	if r.logger != nil {
		start = time.Now()
	}

	rows, err := r.db.QueryxContext(r.ctx, r.query, r.args...)
	if err != nil {
		// Log query execution
		if r.logger != nil {
			duration := time.Since(start)
			r.logger.LogQuery(r.ctx, r.query, r.args, duration, err)
		}
		return nil, err
	}
	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		row := make(map[string]any)
		if err := rows.MapScan(row); err != nil {
			// Log query execution
			if r.logger != nil {
				duration := time.Since(start)
				r.logger.LogQuery(r.ctx, r.query, r.args, duration, err)
			}
			return nil, err
		}
		results = append(results, convertByteArrayToString(row))
	}

	// Log query execution
	if r.logger != nil {
		duration := time.Since(start)
		r.logger.LogQuery(r.ctx, r.query, r.args, duration, nil)
	}

	return results, nil
}

func (r *rawQuery) Exec() (sql.Result, error) {
	// Print SQL if needed
	if r.printSQL {
		// Simple placeholder replacement for debugging
		formattedSQL := r.query
		for _, arg := range r.args {
			formattedSQL = strings.Replace(formattedSQL, "?", fmt.Sprintf("'%v'", arg), 1)
		}
		fmt.Println(formattedSQL)
	}

	// Log query if logger is set
	var start time.Time
	if r.logger != nil {
		start = time.Now()
	}

	result, err := r.db.ExecContext(r.ctx, r.query, r.args...)

	// Log query execution
	if r.logger != nil {
		duration := time.Since(start)
		r.logger.LogQuery(r.ctx, r.query, r.args, duration, err)
	}

	return result, err
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

func (r *rawQuery) PrintSQL() Raw {
	r.printSQL = true
	return r
}

func (r *rawQuery) setLogger(logger Logger) {
	r.logger = logger
}
