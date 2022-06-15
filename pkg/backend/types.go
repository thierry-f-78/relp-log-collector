package backend

import "time"

// this struct is the representation of syslog message.
type Message struct {
	Date     time.Time
	Facility int
	Severity int
	Hostname string
	Process  string
	Pid      int
	Data     string
}

// Plugin must implement Init() function with the following type
type InitFunction func() error

// Plugin must implement NewBatch() function with the following type
type NewBatchFunction func() (Batch, error)

// NewBatch must return these interface
type Batch interface {
	Pick(*Message) (bool, error)
	Flush() error
}

// NewBatch could use this function to initialise backend
// database transaction
func NewTransaction(table string, columns ...string) (TransactionOps, error) {
	return db.newTransaction(table, columns...)
}

type TransactionOps interface {
	Append(...interface{}) error
	Flush() error
}
