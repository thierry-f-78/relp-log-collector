package backend

// register syslog as internal plugin
func syslogInit() *backendRef {
	return &backendRef{
		newBatch: syslogNewBatch,
	}
}

type backendSyslog struct {
	batch TransactionOps
}

func syslogNewBatch() (Batch, error) {
	var bs *backendSyslog
	var err error

	bs = &backendSyslog{}
	bs.batch, err = NewTransaction("syslog", "date", "facility", "severity", "hostname", "process", "pid", "data")

	return bs, err
}

func (bs *backendSyslog) Pick(m *Message) (bool, error) {
	var err error

	err = bs.batch.Append(
		m.Date,
		m.Facility,
		m.Severity,
		m.Hostname,
		m.Process,
		m.Pid,
		m.Data,
	)
	return true, err
}

func (bs *backendSyslog) Flush() error {
	var err error

	err = bs.batch.Flush()

	return err
}
