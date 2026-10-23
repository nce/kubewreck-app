package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDatabaseConfiguration(t *testing.T) {
	for _, name := range []string{"PGHOST", "PGPORT", "PGUSER", "PGPASSWORD", "PGDATABASE", "POSTGRES_SECRET_DIR"} {
		t.Setenv(name, "")
		if err := os.Unsetenv(name); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PGSSLMODE", "disable")
	secret := t.TempDir()
	password := "special '@:/?#\\ password\n"
	for key, value := range map[string]string{
		"host": "my-postgres.default.svc", "port": "5433", "username": "postgres",
		"database": "postgres", "postgres-password": password,
	} {
		if err := os.WriteFile(filepath.Join(secret, key), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("POSTGRES_SECRET_DIR", secret)
	config, err := databaseConfig("")
	if err != nil {
		t.Fatal(err)
	}
	c := config.ConnConfig
	if c.Host != "my-postgres.default.svc" || c.Port != 5433 || c.User != "postgres" || c.Database != "postgres" || c.Password != password {
		t.Fatal("mounted Secret values were not preserved")
	}

	t.Setenv("PGHOST", "env-host")
	t.Setenv("PGUSER", "env-user")
	t.Setenv("PGPASSWORD", "env-password")
	t.Setenv("PGPORT", "5434")
	t.Setenv("PGDATABASE", "env-db")
	config, err = databaseConfig("")
	if err != nil {
		t.Fatal(err)
	}
	c = config.ConnConfig
	if c.Host != "env-host" || c.Port != 5434 || c.User != "env-user" || c.Database != "env-db" || c.Password != "env-password" {
		t.Fatal("PG environment variables must override mounted Secret values")
	}

	config, err = databaseConfig("postgres://url-user:url-password@url-host:5435/url-db?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	c = config.ConnConfig
	if c.Host != "url-host" || c.Port != 5435 || c.User != "url-user" || c.Database != "url-db" || c.Password != "url-password" {
		t.Fatal("DATABASE_URL must take precedence")
	}
	if _, err := databaseConfig("postgres://%invalid"); err == nil {
		t.Fatal("invalid DATABASE_URL must not fall back to other settings")
	}
}

func TestPostgresDefaults(t *testing.T) {
	t.Setenv("POSTGRES_SECRET_DIR", "")
	t.Setenv("PGHOST", "my-postgres")
	for _, name := range []string{"PGPORT", "PGUSER", "PGPASSWORD", "PGDATABASE"} {
		t.Setenv(name, "")
		if err := os.Unsetenv(name); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PGSSLMODE", "disable")
	config, err := databaseConfig("")
	if err != nil {
		t.Fatal(err)
	}
	c := config.ConnConfig
	if c.Port != 5432 || c.User != "postgres" || c.Database != "postgres" {
		t.Fatal("unexpected defaults for a standard PostgreSQL installation")
	}
	t.Setenv("PGPORT", "invalid-port")
	if _, err := databaseConfig(""); err == nil {
		t.Fatal("invalid port must fail configuration")
	}
}
