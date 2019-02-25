package variable

import (
	"fmt"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator/values"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/log"
)

func getCharacterSetResults(c *Container) values.SQLValue {
	if c.characterSetResults.String() == "" {
		return nil
	}
	return c.characterSetResults
}

func setCharacterSetClient(c *Container, v values.SQLValue) error {
	if v.IsNull() {
		return invalidValueError(CharacterSetClient, nil)
	}

	s, err := convertSQLVarchar(CharacterSetResults, v)
	if err != nil {
		return err
	}

	cs, err := collation.GetCharset(collation.CharsetName(s.String()))
	if err != nil {
		return err
	}
	c.characterSetClient = values.NewSQLVarchar(values.VariableSQLValueKind, string(cs.Name))
	return nil
}

func setCharacterSetConnection(c *Container, v values.SQLValue) error {
	if v.IsNull() {
		return invalidValueError(CharacterSetConnection, nil)
	}

	s, err := convertSQLVarchar(CharacterSetConnection, v)
	if err != nil {
		return err
	}

	cs, err := collation.GetCharset(collation.CharsetName(s.String()))
	if err != nil {
		return err
	}

	col, err := collation.Get(cs.DefaultCollationName)
	if err != nil {
		return err
	}

	c.characterSetConnection = values.NewSQLVarchar(values.VariableSQLValueKind, string(cs.Name))
	c.collationConnection = values.NewSQLVarchar(values.VariableSQLValueKind, string(col.Name))
	return nil
}

func setCharacterSetDatabase(c *Container, v values.SQLValue) error {
	if v.IsNull() {
		return invalidValueError(CharacterSetDatabase, nil)
	}

	s, err := convertSQLVarchar(CharacterSetDatabase, v)
	if err != nil {
		return err
	}

	cs, err := collation.GetCharset(collation.CharsetName(s.String()))
	if err != nil {
		return err
	}

	col, err := collation.Get(cs.DefaultCollationName)
	if err != nil {
		return err
	}

	c.characterSetDatabase = values.NewSQLVarchar(values.VariableSQLValueKind, string(cs.Name))
	c.collationDatabase = values.NewSQLVarchar(values.VariableSQLValueKind, string(col.Name))
	return nil
}

func setCharacterSetResults(c *Container, v values.SQLValue) error {
	if v.IsNull() {
		c.characterSetResults = values.NewSQLNull(values.VariableSQLValueKind)
		return nil
	}

	s, err := convertSQLVarchar(CharacterSetResults, v)
	if err != nil {
		return err
	}

	cs, err := collation.GetCharset(collation.CharsetName(s.String()))
	if err != nil {
		return err
	}
	c.characterSetResults = values.NewSQLVarchar(values.VariableSQLValueKind, string(cs.Name))
	return nil
}

func setCollationConnection(c *Container, v values.SQLValue) error {
	if v.IsNull() {
		return invalidValueError(CollationConnection, nil)
	}

	s, err := convertSQLVarchar(CollationConnection, v)
	if err != nil {
		return err
	}

	col, err := collation.Get(collation.Name(s.String()))
	if err != nil {
		return err
	}

	cs, err := collation.GetCharset(col.CharsetName)
	if err != nil {
		return err
	}

	c.characterSetConnection = values.NewSQLVarchar(values.VariableSQLValueKind, string(cs.Name))
	c.collationConnection = values.NewSQLVarchar(values.VariableSQLValueKind, string(col.Name))
	return nil
}

func setCollationDatabase(c *Container, v values.SQLValue) error {
	if v.IsNull() {
		return invalidValueError(CollationDatabase, nil)
	}

	s, err := convertSQLVarchar(CollationDatabase, v)
	if err != nil {
		return err
	}

	col, err := collation.Get(collation.Name(s.String()))
	if err != nil {
		return err
	}

	cs, err := collation.GetCharset(col.CharsetName)
	if err != nil {
		return err
	}

	c.characterSetDatabase = values.NewSQLVarchar(values.VariableSQLValueKind, string(cs.Name))
	c.collationDatabase = values.NewSQLVarchar(values.VariableSQLValueKind, string(col.Name))
	return nil
}

func setCollationServer(c *Container, v values.SQLValue) error {
	if v.IsNull() {
		return invalidValueError(CollationServer, nil)
	}

	s, err := convertSQLVarchar(CollationServer, v)
	if err != nil {
		return err
	}

	col, err := collation.Get(collation.Name(s.String()))
	if err != nil {
		return err
	}

	c.collationServer = values.NewSQLVarchar(values.VariableSQLValueKind, string(col.Name))
	return nil
}

func setGroupConcatMaxLen(c *Container, v values.SQLValue) error {
	i, err := convertSQLInt64(GroupConcatMaxLen, v)
	if err != nil {
		return err
	}
	ii := i.Value().(int64)

	// MySQL's minimum group_concat_max_len value is 4. When a user tries to set the
	// group_concat_max_len system variable to a value < 4, rather than throwing an error,
	// we set the value to the minimum.
	if ii < 4 {
		ii = 4
	}

	c.groupConcatMaxLen = values.NewSQLInt64(values.VariableSQLValueKind, ii)
	return nil
}

func setLogLevel(c *Container, v values.SQLValue) error {
	i, err := convertSQLInt64(LogLevel, v)
	if err != nil {
		return err
	}
	ii := values.Int64(i)

	// Changes the global logger's verbosity to whatever the user inputted.
	// Too high and too low values are handled in  log.SetVerbosity
	// The global logger is the parent of every component logger so this
	// changes all of their verbosity's as well
	log.SetVerbosity(log.Verbosity(ii))
	normalizedVerbosity := log.NormalizeVerbosityLevel(ii)
	c.logLevel = values.NewSQLInt64(values.VariableSQLValueKind, normalizedVerbosity)

	return nil
}

func setWaitTimeoutSecs(c *Container, v values.SQLValue) error {
	i, err := convertSQLInt64(WaitTimeoutSecs, v)
	if err != nil {
		return err
	}
	ii := values.Int64(i)

	upperLimit := int64(31536000)

	if procutil.IsWindowsOS {
		upperLimit = int64(2147483)
	}

	if ii < 1 || ii > upperLimit {
		return mysqlerrors.Defaultf(mysqlerrors.ErWrongValueForVar,
			WaitTimeoutSecs, fmt.Sprintf("%v", ii))
	}

	c.waitTimeoutSecs = values.NewSQLInt64(values.VariableSQLValueKind, ii)
	return nil
}
