// Package storage реализует хранилище метрик с использованием PostgreSQL.
package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lenarlenar/go-my-metrics-service/internal/log"
	"github.com/lenarlenar/go-my-metrics-service/internal/model"
	"github.com/lenarlenar/go-my-metrics-service/internal/server/flags"
)

// DBStorage представляет хранилище метрик, использующее PostgreSQL.
type DBStorage struct {
	DB          *sql.DB
	databaseDSN string
}

// NewDBStorage создает новое хранилище на базе PostgreSQL, проверяет соединение
// и создает таблицу при необходимости.
func NewDBStorage(config flags.Config) *DBStorage {
	storage := &DBStorage{
		databaseDSN: config.DatabaseDSN,
	}

	if err := storage.Ping(); err == nil {
		if err := storage.CreateTable(); err != nil {
			panic(err)
		}
	} else {
		panic(err)
	}

	return storage
}

// CreateTable создает таблицу metrics, если она ещё не существует.
func (m *DBStorage) CreateTable() error {
	_, err := m.DB.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS metrics (
		id SERIAL PRIMARY KEY,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		value DOUBLE PRECISION,
		delta BIGINT
	);`)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

// GetMetrics возвращает все метрики из базы данных в виде map.
func (m *DBStorage) GetMetrics() map[string]model.Metrics {
	rows, err := m.DB.QueryContext(context.Background(), `SELECT * FROM metrics`)
	if err == nil {
		metrics := make(map[string]model.Metrics)
		for rows.Next() {
			var m model.Metrics
			var id interface{}
			err = rows.Scan(&id, &m.MType, &m.ID, &m.Value, &m.Delta)
			if err != nil {
				panic(err)
			}
			metrics[m.ID] = m
		}
		defer rows.Close()
		err = rows.Err()
		if err != nil {
			panic(err)
		}
		return metrics
	} else {
		panic(err)
	}
}

// SetGauge сохраняет значение метрики типа gauge в базу данных.
func (m *DBStorage) SetGauge(n string, v float64) {
	_, err := m.DB.ExecContext(context.Background(), `INSERT INTO metrics (type, name, value, delta)
	VALUES ($1, $2, $3, $4)`,
		"gauge", n, v, nil)
	if err != nil {
		log.I().Warnf("ошибка при попытке добавить в бд метрику типа gauge: %w", err)
	}
}

// AddCounter увеличивает значение метрики counter или создает новую, если она отсутствует.
func (m *DBStorage) AddCounter(n string, v int64) {
	row := m.DB.QueryRowContext(context.Background(), `SELECT delta FROM metrics
			WHERE type = 'counter' AND name = $1`, n)
	var oldValue int64
	err := row.Scan(&oldValue)
	if err == nil {
		newDelta := oldValue + v
		_, err := m.DB.ExecContext(context.Background(), `UPDATE metrics
			SET delta = $1
			WHERE type = 'counter' AND name = $2`, newDelta, n)
		if err != nil {
			log.I().Warnf("ошибка при попытке обновить в бд метрику типа counter: %w", err)
		}
	} else if err == sql.ErrNoRows {
		_, err := m.DB.ExecContext(context.Background(), `INSERT INTO metrics (type, name, value, delta)
		VALUES ($1, $2, $3, $4)`,
			"counter", n, nil, v)
		if err != nil {
			log.I().Warnf("ошибка при попытке добавить в бд метрику типа counter: %w", err)
		}
	} else {
		panic(err)
	}
}

// Ping проверяет соединение с базой данных и инициализирует объект DB.
func (m *DBStorage) Ping() error {
	db, err := sql.Open("postgres", m.databaseDSN)
	m.DB = db
	if err != nil {
		return fmt.Errorf("ошибка при попытке подключиться к базе данных: %w", err)
	}

	return db.Ping()
}
