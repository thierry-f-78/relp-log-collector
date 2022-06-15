package backend

import "database/sql"
import "fmt"
import "log/slog"
import "strconv"
import "testing"
import "time"

import "github.com/thierry-f-78/relp-log-collector/pkg/config"

func TestTransaction(t *testing.T) {
	var err error
	var b *backendRef
	var bl *BatchList
	var i int
	var sqliteDb *sql.DB
	var rows *sql.Rows
	var lines int

	// init SQLITE with in memory table
	config.Cf.SQLite = &config.SQLite{
		Path: ":memory:",
	}
	sqliteInit(slog.Default())
	db, err = sqliteInit(slog.Default())
	if err != nil {
		t.Fatalf("Can't init SQLite: %s", err.Error())
		return
	}
	sqliteDb = db.(*sqliteBackend).db
	_, err = sqliteDb.Exec(`
		CREATE TABLE syslog (
			"date"       DATETIME,
			"facility"   INTEGER,
			"severity"   INTEGER,
			"hostname"   TEXT,
			"process"    TEXT,
			"pid"        INTEGER,
			"data"       TEXT
		);

		CREATE TABLE errors (
			"date"       DATETIME,
			"data"       TEXT
		);
	`)
	if err != nil {
		t.Fatalf("Can't init SQLite: %s", err.Error())
		return
	}

	// init log processing backend syslog and error
	b = syslogInit()
	backendRefList = append(backendRefList, b)

	// Start new batch
	bl, err = NewBatch()
	if err != nil {
		t.Fatalf("Can't create new batch: %s", err.Error())
	}

	// Inject message
	for i = 0; i < 100; i++ {
		err = bl.Pick(&Message{
			Date:     time.Now().Add(time.Second * time.Duration(i)),
			Facility: 3,
			Severity: 3,
			Hostname: "host" + strconv.Itoa(i),
			Process:  "myself",
			Pid:      1,
			Data:     "test number " + strconv.Itoa(i),
		})
		if err != nil {
			t.Fatalf("Can't pick message: %s", err.Error())
		}
	}

	// Flush messages
	err = bl.Flush()
	if err != nil {
		t.Fatalf("Can't flush messages: %s", err.Error())
	}

	// Verification
	rows, err = sqliteDb.Query("SELECT * FROM syslog")
	if err != nil {
		t.Fatalf("Error selecting data from syslog table: %s", err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var date string
		var facility, severity, pid int
		var hostname, process, data string

		lines++

		err = rows.Scan(&date, &facility, &severity, &hostname, &process, &pid, &data)
		if err != nil {
			t.Fatalf("Error scanning data from syslog table: %s", err.Error())
		}

		// Dipslayed on ly in error cases
		fmt.Printf("Date: %s, Facility: %d, Severity: %d, Hostname: %s, Process: %s, PID: %d, Data: %s\n",
			date, facility, severity, hostname, process, pid, data)
	}
	err = rows.Err()
	if err != nil {
		t.Fatalf("Error selecting data from syslog table: %s", err.Error())
	}

	if lines != 100 {
		t.Fatalf("Wrong number of lines: %d", lines)
	}
}
