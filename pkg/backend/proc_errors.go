package backend

// register errors as internal plugin
func errorsInit() *backendRef {
	return &backendRef{
		newBatch: errorsNewBatch,
	}
}

type backendErrors struct {
	batch TransactionOps
}

func errorsNewBatch() (Batch, error) {
	var bs *backendErrors
	var err error

	bs = &backendErrors{}
	bs.batch, err = NewTransaction("error", "date", "data")

	return bs, err
}

func (bs *backendErrors) Pick(m *Message) (bool, error) {
	var err error

	err = bs.batch.Append(
		m.Date,
		m.Data,
	)
	return true, err
}

func (bs *backendErrors) Flush() error {
	var err error

	err = bs.batch.Flush()

	return err
}
