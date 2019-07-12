package mongoutil

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/operation"
)

// ExecuteWithDeployment executes the command cmd on the database db
// using the driver.Deployment d and writes the response to result.
func ExecuteWithDeployment(
	ctx context.Context,
	db string,
	d driver.Deployment,
	cmd bson.D,
	result interface{},
) error {
	doc, err := bson.Marshal(&cmd)
	if err != nil {
		return fmt.Errorf("unable to marshal command: %v", err)
	}

	c := operation.NewCommand(doc).Database(db).Deployment(d)
	if err = c.Execute(ctx); err != nil {
		return fmt.Errorf("unable to execute command: %v", err)
	}

	if err = bson.Unmarshal(c.Result(), result); err != nil {
		return fmt.Errorf("unable to unmarshal command result: %v", err)
	}

	return nil
}

// ExecuteWithConnection executes the command cmd on the database db using
// the driver.Connection c in a driver.SingleConnectionDeployment, and it
// writes the response to result.
func ExecuteWithConnection(
	ctx context.Context,
	db string,
	c driver.Connection,
	cmd bson.D,
	result interface{},
) error {
	d := driver.SingleConnectionDeployment{C: c}
	return ExecuteWithDeployment(ctx, db, d, cmd, result)
}
