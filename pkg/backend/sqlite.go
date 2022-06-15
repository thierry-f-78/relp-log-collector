package backend

import "database/sql"
import "fmt"
import "log/slog"
import "os"
import "strings"

import _ "github.com/mattn/go-sqlite3"

import "github.com/thierry-f-78/relp-log-collector/pkg/config"

type sqliteBackend struct {
	db  *sql.DB
	log *slog.Logger
}

type sqliteBatchOps struct {
	tx   *sql.Tx
	stmt *sql.Stmt
	log  *slog.Logger
}

// Connects to SQLite using the provided logger and configuration
func sqliteInit(inheritLog *slog.Logger) (transaction, error) {
	var dsn string
	var sb *sqliteBackend
	var err error

	if config.Cf.SQLite == nil {
		return nil, nil
	}

	sb = &sqliteBackend{}
	sb.log = inheritLog.WithGroup("sqlite")

	if config.Cf.SQLite.Path != "" {
		dsn = config.Cf.SQLite.Path
	} else {
		return nil, fmt.Errorf("SQLite path is not provided")
	}

	sb.db, err = sql.Open("sqlite3", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't connect to SQLite: %s\n", err.Error())
		os.Exit(1)
	}

	return sb, nil
}

func (sb *sqliteBackend) newTransaction(tableName string, columns ...string) (TransactionOps, error) {
	var batch *sqliteBatchOps
	var err error
	var columnPlaceholders []string
	var query string
	var i int

	batch = &sqliteBatchOps{}
	batch.log = sb.log
	batch.tx, err = sb.db.Begin()
	if err != nil {
		return nil, err
	}

	columnPlaceholders = make([]string, len(columns))
	for i = range columns {
		columnPlaceholders[i] = "?"
	}
	query = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, joinColumns(columns), strings.Join(columnPlaceholders, ","))
	batch.stmt, err = batch.tx.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("Can't prepare SQLite query %q: %s", query, err.Error())
	}

	return batch, err
}

func joinColumns(columns []string) string {
	var columnsEscaped []string
	var col string

	for _, col = range columns {
		columnsEscaped = append(columnsEscaped, fmt.Sprintf("\"%s\"", col))
	}
	return strings.Join(columnsEscaped, ", ")
}

func (ch *sqliteBatchOps) Append(args ...interface{}) error {
	var err error

	_, err = ch.stmt.Exec(args...)
	if err != nil {
		return fmt.Errorf("Can't execute SQLite query: %s", err.Error())
	}

	return nil
}

func (ch *sqliteBatchOps) Flush() error {
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
