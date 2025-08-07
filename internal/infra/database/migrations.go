package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type Migration struct {
	ID      int
	Name    string
	UpSQL   string
	DownSQL string
}

var migrations = []Migration{
	{
		ID:    1,
		Name:  "create_sessions_table",
		UpSQL: createSessionsTableSQL,
	},
	{
		ID:    2,
		Name:  "create_migrations_table",
		UpSQL: createMigrationsTableSQL,
	},
}

const createSessionsTableSQL = `
-- PostgreSQL version
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'sessions') THEN
        CREATE TABLE sessions (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            status TEXT DEFAULT 'disconnected',
            webhook TEXT DEFAULT '',
            jid TEXT DEFAULT '',
            qrcode TEXT DEFAULT '',
            events TEXT DEFAULT '',
            proxy_url TEXT DEFAULT '',
            device_name TEXT DEFAULT 'WazMeow',
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            last_connected_at TIMESTAMP,
            s3_enabled BOOLEAN DEFAULT FALSE,
            s3_endpoint TEXT DEFAULT '',
            s3_region TEXT DEFAULT '',
            s3_bucket TEXT DEFAULT '',
            s3_access_key TEXT DEFAULT '',
            s3_secret_key TEXT DEFAULT '',
            s3_path_style BOOLEAN DEFAULT TRUE,
            s3_public_url TEXT DEFAULT '',
            media_delivery TEXT DEFAULT 'base64',
            s3_retention_days INTEGER DEFAULT 30
        );
    END IF;
END $$;
`

const createMigrationsTableSQL = `
-- PostgreSQL version
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'migrations') THEN
        CREATE TABLE migrations (
            id INTEGER PRIMARY KEY,
            name TEXT NOT NULL,
            applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    END IF;
END $$;
`

// InitializeSchema inicializa o schema do banco de dados
func InitializeSchema(db *sqlx.DB) error {
	// Criar tabela de migrações se não existir
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Obter migrações já aplicadas
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Aplicar migrações pendentes
	for _, migration := range migrations {
		if _, ok := applied[migration.ID]; !ok {
			if err := applyMigration(db, migration); err != nil {
				return fmt.Errorf("failed to apply migration %d: %w", migration.ID, err)
			}
		}
	}

	return nil
}

func createMigrationsTable(db *sqlx.DB) error {
	var tableExists bool
	var err error

	switch db.DriverName() {
	case "postgres":
		err = db.Get(&tableExists, `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables 
				WHERE table_name = 'migrations'
			)`)
	case "sqlite":
		err = db.Get(&tableExists, `
			SELECT EXISTS (
				SELECT 1 FROM sqlite_master 
				WHERE type='table' AND name='migrations'
			)`)
	default:
		return fmt.Errorf("unsupported database driver: %s", db.DriverName())
	}

	if err != nil {
		return fmt.Errorf("failed to check migrations table existence: %w", err)
	}

	if tableExists {
		return nil
	}

	_, err = db.Exec(`
		CREATE TABLE migrations (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}

func getAppliedMigrations(db *sqlx.DB) (map[int]struct{}, error) {
	applied := make(map[int]struct{})
	var rows []struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	err := db.Select(&rows, "SELECT id, name FROM migrations ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}

	for _, row := range rows {
		applied[row.ID] = struct{}{}
	}

	return applied, nil
}

func applyMigration(db *sqlx.DB, migration Migration) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if migration.ID == 1 {
		// Handle sessions table creation differently per database
		if db.DriverName() == "sqlite" {
			err = createSessionsTableSQLite(tx)
		} else {
			_, err = tx.Exec(migration.UpSQL)
		}
	} else {
		_, err = tx.Exec(migration.UpSQL)
	}

	if err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record the migration
	if _, err = tx.Exec(`
        INSERT INTO migrations (id, name) 
        VALUES ($1, $2)`, migration.ID, migration.Name); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	log.Info().Int("migration_id", migration.ID).Str("name", migration.Name).Msg("Migration applied successfully")
	return nil
}

func createSessionsTableSQLite(tx *sqlx.Tx) error {
	var exists int
	err := tx.Get(&exists, `
        SELECT COUNT(*) FROM sqlite_master
        WHERE type='table' AND name='sessions'`)
	if err != nil {
		return err
	}

	if exists == 0 {
		_, err = tx.Exec(`
			CREATE TABLE sessions (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				status TEXT DEFAULT 'disconnected',
				webhook TEXT DEFAULT '',
				jid TEXT DEFAULT '',
				qrcode TEXT DEFAULT '',
				events TEXT DEFAULT '',
				proxy_url TEXT DEFAULT '',
				device_name TEXT DEFAULT 'WazMeow',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				last_connected_at TIMESTAMP,
				s3_enabled BOOLEAN DEFAULT 0,
				s3_endpoint TEXT DEFAULT '',
				s3_region TEXT DEFAULT '',
				s3_bucket TEXT DEFAULT '',
				s3_access_key TEXT DEFAULT '',
				s3_secret_key TEXT DEFAULT '',
				s3_path_style BOOLEAN DEFAULT 1,
				s3_public_url TEXT DEFAULT '',
				media_delivery TEXT DEFAULT 'base64',
				s3_retention_days INTEGER DEFAULT 30
			)`)
		return err
	}
	return nil
}
