package drivers

import (
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"

	"wazmeow/internal/infra/config"
	"wazmeow/pkg/logger"
)

// BaseDriver contém funcionalidades comuns a todos os drivers de banco de dados
type BaseDriver struct {
	DB     *bun.DB
	Config *config.DatabaseConfig
	Logger logger.Logger
}

// GetDB retorna a instância Bun DB
func (b *BaseDriver) GetDB() *bun.DB {
	return b.DB
}

// Close fecha a conexão com o banco de dados
func (b *BaseDriver) Close() error {
	if b.DB != nil {
		if err := b.DB.Close(); err != nil {
			if b.Logger != nil {
				b.Logger.ErrorWithError("failed to close database connection", err, nil)
			}
			return err
		}
		if b.Logger != nil {
			b.Logger.Info("database connection closed")
		}
	}
	return nil
}

// Health verifica a saúde da conexão com o banco
func (b *BaseDriver) Health() error {
	if b.DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	if err := b.DB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// Stats retorna estatísticas da conexão com o banco
func (b *BaseDriver) Stats() sql.DBStats {
	if b.DB == nil {
		return sql.DBStats{}
	}
	return b.DB.DB.Stats()
}

// configureConnectionPool configura o pool de conexões do banco
func (b *BaseDriver) configureConnectionPool(sqlDB *sql.DB) {
	sqlDB.SetMaxOpenConns(b.Config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(b.Config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(b.Config.ConnMaxLifetime)

	if b.Logger != nil {
		b.Logger.InfoWithFields("connection pool configured", logger.Fields{
			"max_open_conns":    b.Config.MaxOpenConns,
			"max_idle_conns":    b.Config.MaxIdleConns,
			"conn_max_lifetime": b.Config.ConnMaxLifetime,
		})
	}
}

// logConnectionEstablished registra log de conexão estabelecida
func (b *BaseDriver) logConnectionEstablished(driverName string, extraFields logger.Fields) {
	if b.Logger != nil {
		fields := logger.Fields{
			"driver":            driverName,
			"max_open_conns":    b.Config.MaxOpenConns,
			"max_idle_conns":    b.Config.MaxIdleConns,
			"conn_max_lifetime": b.Config.ConnMaxLifetime,
		}

		// Merge extra fields
		for k, v := range extraFields {
			fields[k] = v
		}

		b.Logger.InfoWithFields("database connection established", fields)
	}
}
