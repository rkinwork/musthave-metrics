package storage

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rkinwork/musthave-metrics/internal/config"
	"github.com/rkinwork/musthave-metrics/internal/logger"
	"go.uber.org/zap"
	"sync"
)

type PgSaver struct {
	Conn      *sql.DB
	Schema    string
	TableName string
	IMetricRepository
	once sync.Once
}

func (ps *PgSaver) CreateTable(ctx context.Context) error {
	// Use the sync.Once member to ensure CreateTable only called once
	var createErr error
	ps.once.Do(func() {
		tx, err := ps.Conn.BeginTx(ctx, nil)
		if err != nil {
			createErr = err
			return
		}

		defer tx.Rollback()

		_, err = tx.ExecContext(ctx,
			`CREATE TABLE IF NOT EXISTS public.metrics (
				 id text not null,
				 mtype text not null,
				 delta integer,
				 mvalue double precision
			 )`)

		if err != nil {
			createErr = err
			return
		}

		createErr = tx.Commit()
	})

	return createErr
}

func (ps *PgSaver) Save(ctx context.Context) error {
	if err := ps.Ping(ctx); err != nil {
		return err
	}
	if err := ps.CreateTable(ctx); err != nil {
		return err
	}

	metrics := ps.GetAllMetrics()

	values := make([][]interface{}, len(metrics))
	for i, metric := range metrics {
		values[i] = []interface{}{metric.ID, metric.MType, metric.Delta, metric.Value}
	}

	tx, err := ps.Conn.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `TRUNCATE TABLE public.metrics`); err != nil {
		return err
	}

	for i := range metrics {
		// execute your insert query here.
		_, err = tx.ExecContext(ctx, "INSERT INTO public.metrics(id, mtype, delta, mvalue) VALUES($1, $2, $3, $4)", metrics[i].ID, metrics[i].MType, metrics[i].Delta, metrics[i].Value)
		if err != nil {
			return tx.Rollback() // don't forget to rollback here.
		}
	}

	return tx.Commit()
}

func (ps *PgSaver) Load(ctx context.Context) error {

	if err := ps.Ping(ctx); err != nil {
		return err
	}

	if err := ps.CreateTable(ctx); err != nil {
		return err
	}
	rows, err := ps.Conn.QueryContext(ctx, `SELECT id, mtype, delta, mvalue FROM public.metrics`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var metrics []Metrics
	for rows.Next() {
		var metric Metrics
		err = rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
		if err != nil {
			return err
		}
		metrics = append(metrics, metric)
	}

	if rows.Err() != nil {
		return rows.Err()
	}

	for _, metric := range metrics {
		if _, err := ps.Set(&metric); err != nil {
			logger.Log.Error("error while setting metric", zap.Error(err))
		}
	}

	return nil
}

func (ps *PgSaver) Ping(_ context.Context) error {
	return ps.Conn.Ping()
}

func (ps *PgSaver) Close(_ context.Context) error {
	return ps.Conn.Close()
}

func NewPgSaver(cfg *config.Config, repository IMetricRepository) (*PgSaver, error) {
	conn, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		panic(err)
	}
	return &PgSaver{
		Conn:              conn,
		Schema:            "public",
		TableName:         "metrics",
		IMetricRepository: repository,
	}, nil
}
