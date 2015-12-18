package proxy

import (
	"fmt"
	"github.com/deafgoat/mixer/sqlparser"
)

func (c *Conn) handleAdmin(admin *sqlparser.Admin) error {
	name := string(admin.Name)
	return fmt.Errorf("admin %s not supported now", name)
}

func (c *Conn) adminUpNodeServer(values sqlparser.ValExprs) error {
	return fmt.Errorf("adminUpNodeServer not supported now")
}

func (c *Conn) adminDownNodeServer(values sqlparser.ValExprs) error {
	return fmt.Errorf("adminDownNodeServer not supported now")
}
