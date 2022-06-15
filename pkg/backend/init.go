package backend

import "fmt"
import "log/slog"

import "github.com/thierry-f-78/relp-log-collector/pkg/config"

type backendRef struct {
	newBatch func() (Batch, error)
}

var backendRefList []*backendRef
var backendRefError *backendRef

var db transaction

func Init(inheritLog *slog.Logger) error {
	var err error

	// Register internal logs proc. syslog is always at the final position to catch all logs
	// errors is specif to handle unparsable logs
	backendRefList = append(backendRefList, syslogInit())
	backendRefError = errorsInit()

	// Init database connection. If more than one database is configured, only
	// the first in the below list is initialized.
	switch {
	case config.Cf.Clickhouse != nil:
		fmt.Printf("Init Clickhouse\n")
		db, err = clickhouseInit(inheritLog)
	case config.Cf.PostgreSQL != nil:
		fmt.Printf("Init PostgreSQL\n")
		db, err = postgresInit(inheritLog)
	case config.Cf.SQLite != nil:
		fmt.Printf("Init SQLite\n")
		db, err = sqliteInit(inheritLog)
	}
	if err != nil {
		return err
	}

	return nil
}
