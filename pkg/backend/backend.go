package backend

type BatchList struct {
	list []Batch
	errors Batch
}

func NewBatch() (*BatchList, error) {
	var registeredBackend *backendRef
	var err error
	var batch Batch
	var batchList *BatchList

	batchList = &BatchList{}
	for _, registeredBackend = range backendRefList {
		batch, err = registeredBackend.newBatch()
		if err != nil {
			return nil, err
		}
		batchList.list = append(batchList.list, batch)
	}
	batchList.errors, err = backendRefError.newBatch()
	if err != nil {
		return nil, err
	}

	return batchList, nil
}

func (bl *BatchList) Pick(m *Message, isNotDecoded bool) error {
	var picked bool
	var err error
	var batch Batch

	if isNotDecoded {
		picked, err = bl.errors.Pick(m)
		if err != nil {
			return err
		}
		return nil
	}

	for _, batch = range bl.list {
		picked, err = batch.Pick(m)
		if err != nil {
			return err
		}
		if picked {
			return nil
		}
	}

	return nil
}

func (bl *BatchList) Flush() error {
	var err error
	var batch Batch

	err = bl.errors.Flush()
	if err != nil {
		return err
	}

	for _, batch = range bl.list {
		err = batch.Flush()
		if err != nil {
			return err
		}
	}

	return nil
}
