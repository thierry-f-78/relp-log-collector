package backend

type transaction interface {
	newTransaction(tableName string, columns ...string) (TransactionOps, error)
}
