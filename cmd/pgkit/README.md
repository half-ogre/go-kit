# pgkit

A command-line toolkit for managing PostgreSQL databases, including running migrations, creating databases, and dropping databases.

## Installation

Install the CLI tool:

```bash
go install github.com/half-ogre/go-kit/cmd/pgkit@latest
# Or from the go-kit repository: make install-pgkit
```

## Usage

### Connection String

All commands require a PostgreSQL connection string, which can be provided via:
- The `--db` flag
- The `DATABASE_URL` environment variable

Example connection string: `postgres://user:pass@localhost/dbname?sslmode=disable`

### Commands

#### Migrate

Run SQL migration files against your database:

```bash
# Run migrations from a directory
pgkit migrate --dir ./migrations

# Specify database connection
pgkit migrate --db "postgres://user:pass@localhost/dbname?sslmode=disable" --dir ./migrations

# Or use DATABASE_URL environment variable
export DATABASE_URL="postgres://user:pass@localhost/dbname?sslmode=disable"
pgkit migrate --dir ./migrations
```

**Features:**
- Tracks applied migrations in a `pgkit_migrations` table
- Alphabetical ordering of migration files
- Safe idempotent execution (already-applied migrations are skipped)

#### Status

Show all applied migrations:

```bash
# Show migration status
pgkit status

# With explicit database connection
pgkit status --db "postgres://user:pass@localhost/dbname?sslmode=disable"
```

#### Create

Create a new PostgreSQL database:

```bash
# Create a database named 'mydb'
pgkit create mydb

# Database name is inferred from connection string if not provided
pgkit create --db "postgres://user:pass@localhost/newdb?sslmode=disable"
```

#### Drop

Drop (delete) a PostgreSQL database:

```bash
# Drop a database (with confirmation prompt)
pgkit drop mydb

# Skip confirmation prompt
pgkit drop mydb --force

# Database name is inferred from connection string if not provided
pgkit drop --db "postgres://user:pass@localhost/olddb?sslmode=disable" --force
```

## Migration Files

Migration files should:
- Be SQL files (`.sql` extension)
- Be named in alphabetical order (e.g., `001_initial.sql`, `002_add_users.sql`)
- Contain valid PostgreSQL SQL

Example migration file (`001_create_users.sql`):

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## How Migrations Work

The `migrate` command:
1. Creates a `pgkit_migrations` table (if it doesn't exist) to track applied migrations
2. Reads all `.sql` files from the specified directory
3. Sorts files alphabetically
4. For each file, checks if it has been applied
5. If not applied, executes the SQL and records the migration in the `pgkit_migrations` table
6. Skips already-applied migrations, making it safe to run repeatedly

## License

See LICENSE.md in the root of this repository.
