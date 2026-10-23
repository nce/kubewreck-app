package main

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const databaseTimeout = 3 * time.Second

type teamStore struct {
	pool      *pgxpool.Pool
	mu        sync.Mutex
	connected bool
}

func newTeamStore(databaseURL string) *teamStore {
	s := &teamStore{}
	config, err := databaseConfig(databaseURL)
	if err != nil {
		log.Printf("PostgreSQL configuration failed: %v", err)
		return s
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Print("PostgreSQL configuration failed: unable to initialize connection pool")
		return s
	}
	s.pool = pool
	return s
}

// Precedence: DATABASE_URL, PG* environment variables, mounted Secret, defaults.
func databaseConfig(databaseURL string) (*pgxpool.Config, error) {
	connectionString := databaseURL
	if databaseURL == "" {
		var settings []string
		var host string
		for _, field := range []struct {
			env, key, parameter, fallback string
		}{
			{"PGHOST", "host", "host", ""},
			{"PGPORT", "port", "port", "5432"},
			{"PGUSER", "username", "user", "postgres"},
			{"PGPASSWORD", "postgres-password", "password", ""},
			{"PGDATABASE", "database", "dbname", "postgres"},
		} {
			value, supplied := os.LookupEnv(field.env)
			if !supplied {
				value = field.fallback
				if directory := os.Getenv("POSTGRES_SECRET_DIR"); directory != "" {
					data, err := os.ReadFile(filepath.Join(directory, field.key))
					if err != nil && !errors.Is(err, os.ErrNotExist) {
						return nil, errors.New("cannot read PostgreSQL Secret key " + field.key)
					}
					if err == nil {
						value = string(data)
					}
				}
			}
			if field.env == "PGHOST" {
				host = value
			}
			// Quote libpq values rather than interpolating credentials into a URL.
			value = strings.ReplaceAll(value, `\`, `\\`)
			value = strings.ReplaceAll(value, `'`, `\'`)
			settings = append(settings, field.parameter+"='"+value+"'")
		}
		if host == "" {
			return nil, errors.New("set DATABASE_URL, PGHOST, or mount a host key using POSTGRES_SECRET_DIR")
		}
		connectionString = strings.Join(settings, " ")
	}
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		// Driver parse errors can contain credentials, so never return them verbatim.
		return nil, errors.New("invalid PostgreSQL connection settings; check DATABASE_URL or PG* variables and Secret keys")
	}
	return config, nil
}

func (s *teamStore) teamName(ctx context.Context) (string, error) {
	if s.pool == nil {
		return "", errors.New("database is not configured")
	}
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()
	var name string
	err := s.pool.QueryRow(ctx, "SELECT name FROM teams WHERE id = 1").Scan(&name)
	s.reportConnection(err)
	return name, err
}

func (s *teamStore) reportConnection(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err == nil {
		if !s.connected {
			log.Print("PostgreSQL connection succeeded")
		}
		s.connected = true
		return
	}
	s.connected = false
	// Report useful failure categories without logging credentials or connection strings.
	var pgErr *pgconn.PgError
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		log.Print("PostgreSQL team lookup failed: team id=1 is missing; apply db/init.sql")
	case errors.As(err, &pgErr):
		log.Printf("PostgreSQL operation failed: SQLSTATE=%s (check credentials, database, and schema)", pgErr.Code)
	case errors.Is(err, context.DeadlineExceeded):
		log.Print("PostgreSQL connection/query failed: timed out after 3 seconds")
	case errors.Is(err, context.Canceled):
		log.Print("PostgreSQL operation canceled")
	default:
		log.Print("PostgreSQL connection/query failed: check database availability, network, TLS, and credentials")
	}
}
