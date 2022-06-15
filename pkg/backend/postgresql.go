package backend

import "database/sql"
import "fmt"
import "log/slog"
import "os"

import "github.com/lib/pq"

import "github.com/thierry-f-78/relp-log-collector/pkg/config"

type postgresBackend struct {
	db  *sql.DB
	log *slog.Logger
}

type postgresBatchOps struct {
	tx   *sql.Tx
	stmt *sql.Stmt
	log  *slog.Logger
}

// Connects to PostgreSQL using the provided logger and configuration
func postgresInit(inheritLog *slog.Logger) (transaction, error) {
	var dsn_parts []string
	var dsn string
	var pb *postgresBackend
	var err error

	if config.Cf.PostgreSQL == nil {
		return nil, nil
	}

	pb = &postgresBackend{}
	pb.log = inheritLog.WithGroup("postgres")

	if config.Cf.PostgreSQL.Host != "" {
		dsn_parts = append(dsn_parts, fmt.Sprintf("host=%s", config.Cf.PostgreSQL.Host))
	}

	if config.Cf.PostgreSQL.Port != 0 {
		dsn_parts = append(dsn_parts, fmt.Sprintf("port=%d", config.Cf.PostgreSQL.Port))
	}

	if config.Cf.PostgreSQL.User != "" {
		dsn_parts = append(dsn_parts, fmt.Sprintf("user=%s", config.Cf.PostgreSQL.User))
	}

	if config.Cf.PostgreSQL.Password != "" {
		dsn_parts = append(dsn_parts, fmt.Sprintf("password=%s", config.Cf.PostgreSQL.Password))
	}

	if config.Cf.PostgreSQL.Name != "" {
		dsn_parts = append(dsn_parts, fmt.Sprintf("dbname=%s", config.Cf.PostgreSQL.Name))
	}

	if config.Cf.PostgreSQL.SSLMode != "" {
		dsn_parts = append(dsn_parts, fmt.Sprintf("sslmode=%s", config.Cf.PostgreSQL.SSLMode))
	}

	dsn_parts = append(dsn_parts, fmt.Sprintf("connect_timeout=%d", int(config.Cf.PostgreSQL.Timeout)))

	pb.db, err = sql.Open("postgres", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't connect to PostgreSQL: %s\n", err.Error())
		os.Exit(1)
	}

	return pb, nil
}

func (pb *postgresBackend) newTransaction(tableName string, columns ...string) (TransactionOps, error) {
	var batch *postgresBatchOps
	var err error

	batch = &postgresBatchOps{}
	batch.log = pb.log
	batch.tx, err = pb.db.Begin()
	if err != nil {
		return nil, err
	}

	batch.stmt, err = batch.tx.Prepare(pq.CopyIn(tableName, columns...))
	if err != nil {
		return nil, err
	}

	return batch, err
}

func (ch *postgresBatchOps) Append(args ...interface{}) error {
	var err error

	_, err = ch.stmt.Exec(args...)

	return err
}

func (ch *postgresBatchOps) Flush() error {
	var err error

	err = ch.stmt.Close()
	if err != nil {
		return err
	}

	err = ch.tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
