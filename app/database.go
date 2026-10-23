package main

import (
	"context"
	"errors"
	"log"
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
	if databaseURL == "" {
		log.Print("PostgreSQL configuration failed: DATABASE_URL is required")
		return s
	}
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		// Parse errors may contain the connection string, including its password.
		log.Print("PostgreSQL configuration failed: invalid DATABASE_URL")
		return s
	}
	s.pool = pool
	return s
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
