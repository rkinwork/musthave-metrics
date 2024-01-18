package storage

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/rkinwork/musthave-metrics/internal/config"
	"github.com/rkinwork/musthave-metrics/internal/logger"
	"go.uber.org/zap"
	"sync"
)

type PgSaver struct {
	Conn      *pgx.Conn
	Schema    string
	TableName string
	IMetricRepository
	once sync.Once
}

func (ps *PgSaver) CreateTable(ctx context.Context) error {
	// Use the sync.Once member to ensure CreateTable only called once
	var createErr error
	ps.once.Do(func() {
		tx, err := ps.Conn.Begin(ctx)
		if err != nil {
			createErr = err
			return
		}

		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx,
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

		createErr = tx.Commit(ctx)
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
	columns := []string{"id", "type", "delta", "value"}

	values := make([][]interface{}, len(metrics))
	for i, metric := range metrics {
		values[i] = []interface{}{metric.ID, metric.MType, metric.Delta, metric.Value}
	}

	tx, err := ps.Conn.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `TRUNCATE TABLE public.metrics`); err != nil {
		return err
	}

	_, err = tx.CopyFrom(ctx, pgx.Identifier{ps.Schema, ps.TableName}, columns, pgx.CopyFromRows(values))
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (ps *PgSaver) Load(ctx context.Context) error {

	if err := ps.Ping(ctx); err != nil {
		return err
	}

	if err := ps.CreateTable(ctx); err != nil {
		return err
	}
	rows, err := ps.Conn.Query(ctx, `SELECT id, mtype, delta, mvalue FROM public.metrics`)
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

func (ps *PgSaver) Ping(ctx context.Context) error {
	return ps.Conn.Ping(ctx)
}

func (ps *PgSaver) Close(ctx context.Context) error {
	return ps.Conn.Close(ctx)
}

func NewPgSaver(cfg *config.Config, repository IMetricRepository) (*PgSaver, error) {
	connConfig, err := pgx.ParseConfig(cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}
	conn, err := pgx.ConnectConfig(context.Background(), connConfig)
	if err != nil {
		return nil, err
	}
	return &PgSaver{
		Conn:              conn,
		Schema:            "public",
		TableName:         "metrics",
		IMetricRepository: repository,
	}, nil
}
