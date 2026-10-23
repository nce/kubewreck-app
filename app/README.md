<p align="center" width="100%">
    <kbd>
        <img src="./docs/logo.png" width="400" />
    </kbd>
</p>

# Kubewreck App

This is a demo application which renders an HTML page with a team name from PostgreSQL and
instructions on how to upgrade the application in kubernetes.

In our `kubewreck` challenges, we tinker with kubernetes components/resources
and maliciously block the straight forward upgrade path.
After the teams manages to mitigate all issues, the `v2` Image
displays a new version of the page.


## PostgreSQL setup

Provision a PostgreSQL database and configure its connection using the required
`DATABASE_URL` environment variable. This includes the host, port, database,
username, password, and TLS settings. For a local development database:

```sh
export DATABASE_URL='postgres://kubewreck:local-password@localhost:5432/kubewreck?sslmode=disable'
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f app/db/init.sql
cd app
go run .
```

Open http://localhost:8080. Use `sslmode=verify-full` with the appropriate CA
configuration when connecting to a remote database that requires verified TLS.
URL-encode special characters in credentials. The application reads the environment
at startup; it does not load `.env` files automatically.

The schema has a single `teams` row (`id = 1`), initially named `opdt`.
The initialization script is safe to re-run and preserves an existing name.
Provision the schema before starting the app; its runtime database user only
needs SELECT permission on `teams`.

Change the greeting without rebuilding the application:

```sql
UPDATE teams SET name = 'new team name' WHERE id = 1;
```

Each page load queries PostgreSQL with a three-second timeout. If configuration
is missing/invalid, the database is unreachable, the schema is missing, or no team
row exists, the page displays an error and returns HTTP 503. There is no fallback
team name. The HTTP server stays running during an outage and retries on page
reload; after PostgreSQL recovers, the greeting returns automatically. Configuration
changes require restarting the application. `/version` and `/latest` remain available
independently of the database.

Logs report the startup connection check, lookup failures, and recovery, without
printing credentials or the connection URL. PostgreSQL errors include their
SQLSTATE; connection failures include a diagnostic category.

## Tests

```sh
cd app
go test ./...
```

## Build

Build App and Image via:

```
task build
```

Push Image to ghcr:
```
task push
```
