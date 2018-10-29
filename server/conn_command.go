package server

import (
	"context"
	"fmt"

	"github.com/10gen/mongo-go-driver/mongo/private/ops"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/variable"
	"github.com/10gen/sqlproxy/log"
	"github.com/10gen/sqlproxy/mongodb"
	"github.com/10gen/sqlproxy/schema"
)

func (c *conn) getCommandHandler() *commandHandler {
	return &commandHandler{
		conn: c,
		lg:   c.Logger(log.ControlComponent),
	}
}

// commandHandler implements the evaluator.CommandHandler interface.
type commandHandler struct {
	conn *conn
	lg   log.Logger
}

func (ch *commandHandler) isAdminUser() bool {
	user := ch.conn.user
	source := ch.conn.source
	return ch.conn.server.isAdminUser(user, source)
}

func (ch *commandHandler) Aggregate(ctx context.Context, db, col string, pipeline interface{}) (ops.Cursor, error) {
	return ch.conn.session.Aggregate(ctx, db, col, pipeline)
}

func (ch *commandHandler) Alter(ctx context.Context, alts []*schema.Alteration) error {
	info := ch.conn.variables.MongoDBInfo
	if info.IsSecurityEnabled() {
		if !(info.IsAllowedSampleSource(mongodb.InsertPrivilege|mongodb.UpdatePrivilege) || ch.isAdminUser()) {
			return fmt.Errorf(
				"must have `insert` and `update` privileges for the " +
					"'sample source' or be admin user in order to alter tables")
		}
	}

	sch, err := ch.conn.server.Alter(ctx, alts)
	if err != nil {
		return err
	}
	return ch.conn.updateCatalog(ctx, sch)
}

func (ch *commandHandler) Count(ctx context.Context, db, col string) (int, error) {
	return ch.conn.session.Count(ctx, db, col)
}

func (ch *commandHandler) Kill(ctx context.Context, targetConnID uint32, killScope evaluator.KillScope) error {
	user := ch.conn.user
	if ch.conn.variables.MongoDBInfo.IsSecurityEnabled() {
		ok, err := ch.conn.server.isProcessOwner(user, targetConnID)
		if err != nil {
			return err
		}
		if !ok {
			return mysqlerrors.Defaultf(mysqlerrors.ErKillDeniedError, targetConnID)
		}
	}

	return ch.conn.server.Kill(ctx, ch.conn.connectionID, targetConnID, killScope)
}

func (ch *commandHandler) Resample(ctx context.Context) error {
	info := ch.conn.variables.MongoDBInfo
	if info.IsSecurityEnabled() {
		if !(info.IsAllowedSampleSource(mongodb.UpdatePrivilege|mongodb.InsertPrivilege) || ch.isAdminUser()) {
			return fmt.Errorf("must have " +
				"`insert` and `update` privileges on " +
				"the 'sample source' or be admin user in order to flush sample")
		}

		// In Clustered Write Mode and Standalone Mode we ensure that the user
		// can read all collections that will be sampled. This allows a DBA to give
		// privileges to flush sample to a trusted user who is not the single admin user,
		// and the privileges make sense from the perspective that the user is allowed
		// to see all tables. In Clustered Read Mode, flushing is not allowed, but that
		// is caught in the resample implementation.
		if ch.conn.catalog.HasAuthRestrictedNamespaces() {
			// Do not print out the namespaces the user does not have access to;
			// that would be a minor security breach when namespace names are
			// sensitive.
			return fmt.Errorf("must have " +
				"`find` privileges on the 'sample source' in order to flush sample")
		}
	}

	ch.lg.Infof(log.Always, "sample refresh initiated")

	sch, err := ch.conn.server.Resample(ctx)
	if err != nil {
		return err
	}

	ch.lg.Infof(log.Always, "sample refresh completed")
	return ch.conn.updateCatalog(ctx, sch)
}

func (ch *commandHandler) RotateLogs() error {
	info := ch.conn.variables.MongoDBInfo
	if info.IsSecurityEnabled() {
		if !ch.isAdminUser() {
			return fmt.Errorf("only admin user can flush logs")
		}
	}

	return ch.conn.server.RotateLogs()
}

func (ch *commandHandler) Set(name variable.Name, scope variable.Scope, kind variable.Kind, value interface{}) error {
	err := ch.SetScopeAuthorized(scope)
	if err != nil {
		return err
	}
	return ch.conn.variables.Set(name, scope, kind, value)
}

func (ch *commandHandler) SetScopeAuthorized(scope variable.Scope) error {
	info := ch.conn.variables.MongoDBInfo
	if info.IsSecurityEnabled() {
		if scope == variable.GlobalScope && !ch.isAdminUser() {
			return fmt.Errorf("only admin user can set global variables")
		}
	}
	return nil
}
