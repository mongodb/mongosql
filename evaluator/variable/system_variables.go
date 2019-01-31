package variable

import (
	"fmt"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/internal/mysqlerrors"
	"github.com/10gen/sqlproxy/internal/procutil"
	"github.com/10gen/sqlproxy/log"
)

func getCharacterSetResults(c *Container) interface{} {
	if c.characterSetResults == "" {
		return nil
	}
	return c.characterSetResults
}

func setCharacterSetClient(c *Container, v interface{}) error {
	if v == nil {
		return invalidValueError(CharacterSetClient, nil)
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(CharacterSetClient, v)
	}

	cs, err := collation.GetCharset(collation.CharsetName(s))
	if err != nil {
		return err
	}
	c.characterSetClient = string(cs.Name)
	return nil
}

func setCharacterSetConnection(c *Container, v interface{}) error {
	if v == nil {
		return invalidValueError(CharacterSetConnection, nil)
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(CharacterSetConnection, v)
	}

	cs, err := collation.GetCharset(collation.CharsetName(s))
	if err != nil {
		return err
	}

	col, err := collation.Get(cs.DefaultCollationName)
	if err != nil {
		return err
	}

	c.characterSetConnection = string(cs.Name)
	c.collationConnection = string(col.Name)
	return nil
}

func setCharacterSetDatabase(c *Container, v interface{}) error {
	if v == nil {
		return invalidValueError(CharacterSetDatabase, nil)
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(CharacterSetDatabase, v)
	}

	cs, err := collation.GetCharset(collation.CharsetName(s))
	if err != nil {
		return err
	}

	col, err := collation.Get(cs.DefaultCollationName)
	if err != nil {
		return err
	}

	c.characterSetDatabase = string(cs.Name)
	c.collationDatabase = string(col.Name)
	return nil
}

func setCharacterSetResults(c *Container, v interface{}) error {
	if v == nil {
		c.characterSetResults = string(collation.NullCharset.Name)
		return nil
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(CharacterSetResults, v)
	}

	cs, err := collation.GetCharset(collation.CharsetName(s))
	if err != nil {
		return err
	}
	c.characterSetResults = string(cs.Name)
	return nil
}

func setCollationConnection(c *Container, v interface{}) error {
	if v == nil {
		return invalidValueError(CollationConnection, nil)
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(CollationConnection, v)
	}

	col, err := collation.Get(collation.Name(s))
	if err != nil {
		return err
	}

	cs, err := collation.GetCharset(col.CharsetName)
	if err != nil {
		return err
	}

	c.characterSetConnection = string(cs.Name)
	c.collationConnection = string(col.Name)
	return nil
}

func setCollationDatabase(c *Container, v interface{}) error {
	if v == nil {
		return invalidValueError(CollationDatabase, nil)
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(CollationDatabase, v)
	}

	col, err := collation.Get(collation.Name(s))
	if err != nil {
		return err
	}

	cs, err := collation.GetCharset(col.CharsetName)
	if err != nil {
		return err
	}

	c.characterSetDatabase = string(cs.Name)
	c.collationDatabase = string(col.Name)
	return nil
}

func setCollationServer(c *Container, v interface{}) error {
	if v == nil {
		return invalidValueError(CollationServer, nil)
	}

	s, ok := convertString(v)
	if !ok {
		return wrongTypeError(CollationServer, v)
	}

	col, err := collation.Get(collation.Name(s))
	if err != nil {
		return err
	}

	c.collationServer = string(col.Name)
	return nil
}

func setGroupConcatMaxLen(c *Container, v interface{}) error {
	i, ok := convertInt64(v)
	if !ok {
		return wrongTypeError(GroupConcatMaxLen, v)
	}

	// MySQL's minimum group_concat_max_len value is 4. When a user tries to set the
	// group_concat_max_len system variable to a value < 4, rather than throwing an error,
	// we set the value to the minimum.
	if i < 4 {
		i = 4
	}

	c.groupConcatMaxLen = i
	return nil
}

func setLogLevel(c *Container, v interface{}) error {
	i, ok := convertInt64(v)
	if !ok {
		return wrongTypeError(LogLevel, v)
	}

	// Changes the global logger's verbosity to whatever the user inputted.
	// Too high and too low values are handled in  log.SetVerbosity
	// The global logger is the parent of every component logger so this
	// changes all of their verbosity's as well
	log.SetVerbosity(log.Verbosity(i))
	normalizedVerbosity := log.NormalizeVerbosityLevel(i)
	c.logLevel = normalizedVerbosity

	return nil
}

func setWaitTimeoutSecs(c *Container, v interface{}) error {
	i, ok := convertInt64(v)
	if !ok {
		return wrongTypeError(WaitTimeoutSecs, v)
	}

	upperLimit := int64(31536000)

	if procutil.IsWindowsOS {
		upperLimit = int64(2147483)
	}

	if i < 1 || i > upperLimit {
		return mysqlerrors.Defaultf(mysqlerrors.ErWrongValueForVar,
			WaitTimeoutSecs, fmt.Sprintf("%v", i))
	}

	c.waitTimeoutSecs = i
	return nil
}
