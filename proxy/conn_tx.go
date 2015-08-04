package proxy

func (c *Conn) isInTransaction() bool {
	return false
}

func (c *Conn) isAutoCommit() bool {
	return true
}

func (c *Conn) handleBegin() error {
	return nil
}

func (c *Conn) handleCommit() (err error) {
	return nil
}

func (c *Conn) handleRollback() (err error) {
	return nil
}

func (c *Conn) commit() (err error) {
	return nil
}

func (c *Conn) rollback() (err error) {
	return nil
}

func (c *Conn) needBeginTx() bool {
	return false
}
