package backend

import "context"
import "log/slog"
import "time"

import "github.com/ClickHouse/clickhouse-go/v2"
import "github.com/ClickHouse/clickhouse-go/v2/lib/driver"

import "github.com/thierry-f-78/relp-log-collector/pkg/config"

type clickhouseBackend struct {
	clickhouseConn driver.Conn
	log            *slog.Logger
}

type clickhouseBatchOps struct {
	logLines int
	batch    driver.Batch
	log      *slog.Logger
}

// Connects to ClickHouse using the provided logger and configuration
func clickhouseInit(inheritLog *slog.Logger) (transaction, error) {
	var options *clickhouse.Options
	var err error
	var ch *clickhouseBackend

	ch = &clickhouseBackend{}
	ch.log = inheritLog.WithGroup("clickhouse")

	options = &clickhouse.Options{
		Addr:  config.Cf.Clickhouse.Target,
		Debug: true,
		Debugf: func(format string, v ...any) {
			ch.log.Debug("clickhouse: "+format, v...)
		},
		DialTimeout:          time.Duration(30) * time.Second,
		MaxOpenConns:         50,
		MaxIdleConns:         5,
		ConnMaxLifetime:      time.Duration(10) * time.Minute,
		ConnOpenStrategy:     clickhouse.ConnOpenInOrder,
		BlockBufferSize:      10,
		MaxCompressionBuffer: 10240,
	}

	ch.clickhouseConn, err = clickhouse.Open(options)
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func (ch *clickhouseBackend) newTransaction(tableName string, columns ...string) (TransactionOps, error) {
	var batch *clickhouseBatchOps
	var err error

	batch = &clickhouseBatchOps{}
	batch.log = ch.log
	batch.logLines = 0
	batch.batch, err = ch.clickhouseConn.PrepareBatch(context.Background(), "INSERT INTO "+tableName)

	return batch, err
}

func (ch *clickhouseBatchOps) Append(args ...interface{}) error {
	var err error

	err = ch.batch.Append(args...)
	if err == nil {
		ch.logLines++
	}
	return err
}

func (ch *clickhouseBatchOps) Flush() error {
	var err error

	if ch.logLines > 0 {
		err = ch.batch.Send()
	} else {
		err = ch.batch.Abort()
	}

	// Ensure is never reused
	ch.batch = nil

	return err
}
