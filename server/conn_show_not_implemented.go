package server

import (
	"strings"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/mysqlerrors"
	"github.com/10gen/sqlproxy/parser"
	"github.com/10gen/sqlproxy/schema"
	"github.com/10gen/sqlproxy/variable"
)

func (c *conn) handleShowNotImplemented(sql string, stmt *parser.Show) error {

	var err error
	var r *Resultset
	switch strings.ToLower(stmt.Section) {
	case "binary logs", "master logs":
		r, err = c.buildEmptyResultset(
			[]string{"Log_name", "File_size"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLInt},
		)
	case "binlog events", "relaylog events":
		r, err = c.buildEmptyResultset(
			[]string{"Log_name", "Pos", "Event_type", "Server_id", "End_log_pos", "Info"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLInt, schema.SQLVarchar, schema.SQLInt,
				schema.SQLInt, schema.SQLVarchar},
		)
	case "create database", "create schema":
		r, err = c.buildEmptyResultset(
			[]string{"Database", "Create Database"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLVarchar},
		)
	case "create event":
		r, err = c.buildEmptyResultset(
			[]string{"Event", "sql_mode", "time_zone", "Create Event", "character_set_client",
				"collation_connection", "Database Collation"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar},
		)
	case "create function":
		r, err = c.buildEmptyResultset(
			[]string{"Function", "sql_mode", "time_zone", "Create Function", "character_set_client",
				"collation_connection", "Database Collation"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar},
		)
	case "create procedure":
		r, err = c.buildEmptyResultset(
			[]string{"Procedure", "sql_mode", "time_zone", "Create Procedure",
				"character_set_client", "collation_connection", "Database Collation"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar},
		)
	case "create trigger":
		r, err = c.buildEmptyResultset(
			[]string{"Trigger", "sql_mode", "time_zone", "SQL Original Statement",
				"character_set_client", "collation_connection", "Database Collation", "Created"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar,
				schema.SQLTimestamp},
		)
	case "create user":
		r, err = c.buildEmptyResultset(
			[]string{"CREATE USER for " + stmt.Modifier},
			[]schema.SQLType{schema.SQLVarchar},
		)
	case "create view":
		r, err = c.buildEmptyResultset(
			[]string{"View", "Create View", "character_set_client", "collation_connection"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar,
				schema.SQLVarchar},
		)
	case "engine":
		r, err = c.buildEmptyResultset(
			[]string{"Type", "Name", "Status"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar},
		)
	case "engines":
		r, err = c.buildEmptyResultset(
			[]string{"Engine", "Support", "Comment", "Transactions", "XA", "Savepoints"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar},
		)
	case "errors":
		r, err = c.buildEmptyResultset(
			[]string{"Level", "Code", "Message"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLInt, schema.SQLVarchar},
		)
	case "count(*) errors":
		r, err = c.buildEmptyResultset(
			[]string{"@@session.error_count"},
			[]schema.SQLType{schema.SQLInt},
		)
	case "events":
		r, err = c.buildEmptyResultset(
			[]string{"Db", "Name", "Definer", "Time zone",
				"Type", "Execute at", "Interval value", "Interval field",
				"Starts", "Ends", "Status", "Originator",
				"character_set_client", "collation_connection", "Database Collation"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar,
				schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLTimestamp, schema.SQLInt, schema.SQLVarchar,
				schema.SQLTimestamp, schema.SQLTimestamp, schema.SQLVarchar, schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar},
		)
	case "function code":
		err = mysqlerrors.Defaultf(mysqlerrors.ErFeatureDisabled, "function code")
	case "function status":
		r, err = c.buildEmptyResultset(
			[]string{"Db", "Name", "Type", "Definer",
				"Modifier", "Created", "Security_type", "Comment",
				"character_set_client", "collation_connection", "Database Collation"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar,
				schema.SQLVarchar,
				schema.SQLTimestamp, schema.SQLTimestamp, schema.SQLVarchar, schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar},
		)
	case "grants":
		r, err = c.buildEmptyResultset(
			[]string{"Grants for " + stmt.Modifier},
			[]schema.SQLType{schema.SQLVarchar},
		)
	case "index", "indexes", "keys":
		r, err = c.buildEmptyResultset(
			[]string{"Table", "Non_unique", "Key_name", "Seq_in_index",
				"Column_name", "Collation", "Cardinality", "Sub_part",
				"Packed", "Null", "Index_type", "Comment",
				"Index_comment"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLBoolean, schema.SQLVarchar,
				schema.SQLInt,
				schema.SQLVarchar, schema.SQLVarchar, schema.SQLInt, schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar,
				schema.SQLVarchar},
		)
	case "master status":
		r, err = c.buildEmptyResultset(
			[]string{"File", "Position", "Binlog_Do_DB", "Binlog_Ignore_DB", "Executed_Gtid_Set"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLInt, schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLVarchar},
		)
	case "open tables":
		r, err = c.buildEmptyResultset(
			[]string{"Database", "Table", "In_use", "Name_locked"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLVarchar, schema.SQLInt,
				schema.SQLBoolean},
		)
	case "plugins":
		r, err = c.buildEmptyResultset(
			[]string{"Name", "Status", "Type", "Library", "License"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLVarchar},
		)
	case "privileges":
		r, err = c.buildEmptyResultset(
			[]string{"Privilege", "Context", "Comment"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar},
		)
	case "procedure code":
		err = mysqlerrors.Defaultf(mysqlerrors.ErFeatureDisabled, "procedure code")
	case "procedure status":
		r, err = c.buildEmptyResultset(
			[]string{"Db", "Name", "Type", "Definer",
				"Modifier", "Created", "Security_type", "Comment",
				"character_set_client", "collation_connection", "Database Collation"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLTimestamp, schema.SQLTimestamp, schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar},
		)
	case "profile":
		r, err = c.buildEmptyResultset(
			[]string{"Status", "Duration"},
			[]schema.SQLType{schema.SQLInt, schema.SQLFloat},
		)
	case "profiles":
		r, err = c.buildEmptyResultset(
			[]string{"Query_ID", "Duration", "Query"},
			[]schema.SQLType{schema.SQLInt, schema.SQLFloat, schema.SQLVarchar},
		)
	case "slave hosts":
		r, err = c.buildEmptyResultset(
			[]string{"Server_id", "Host", "Port", "Master_id", "Slave_UUID"},
			[]schema.SQLType{schema.SQLInt, schema.SQLVarchar, schema.SQLInt, schema.SQLInt,
				schema.SQLVarchar},
		)
	case "slave status":
		r, err = c.buildEmptyResultset(
			[]string{"Slave_IO_State", "Master_Host", "Master_User", "Master_Port",
				"Connect_Retry", "Master_Log_File", "Read_Master_Log_Pos", "Relay_Log_File",
				"Relay_Log_Pos", "Relay_Master_Log_File", "Slave_IO_Running", "Slave_SQL_Running",
				"Replicate_Do_DB", "Replicate_Ignore_DB", "Replicate_Do_Table",
				"Replicate_Ignore_Table", "Replicate_Wild_Do_Table", "Replicate_Wild_Ignore_Table",
				"Last_Errno", "Last_Error", "Skip_Counter", "Exec_Master_Log_Pos",
				"Relay_Log_Space", "Until_Condition", "Until_Log_File", "Until_Log_Pos",
				"Master_SSL_Allowed", "Master_SSL_CA_File", "Master_SSL_CA_Path", "Master_SSL_Cert",
				"Master_SSL_Cipher", "Master_SSL_Key", "Seconds_Behind_Master",
				"Master_SSL_Verify_Server_Cert", "Last_IO_Errno", "Last_IO_Error",
				"Last_SQL_Errno", "Last_SQL_Error", "Replica_Ignore_Server_Ids", "Master_Server_Id",
				"Master_UUID", "Master_Info_File", "SQL_Delay", "SQL_Remaining_Delay",
				"Slave_SQL_Running_State", "Master_Retry_Count", "Master_Bind",
				"Last_IO_Error_Timestamp", "Last_SQL_Error_Timestamp", "Master_SSL_Crl",
				"Master_SSL_Crlpath", "Retrieved_Gtid_Set", "Executed_Gtid_Set", "Auto_Position",
				"Replicate_Rewrite_DB", "Channel_name"},
			[]schema.SQLType{
				schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar, schema.SQLInt,
				schema.SQLInt, schema.SQLVarchar, schema.SQLInt, schema.SQLVarchar,
				schema.SQLInt, schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLVarchar, schema.SQLInt, schema.SQLVarchar,
				schema.SQLInt, schema.SQLInt, schema.SQLInt, schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar,
				schema.SQLInt, schema.SQLVarchar, schema.SQLInt, schema.SQLVarchar,
				schema.SQLInt, schema.SQLVarchar, schema.SQLVarchar, schema.SQLInt,
				schema.SQLVarchar, schema.SQLVarchar, schema.SQLInt, schema.SQLInt,
				schema.SQLVarchar, schema.SQLInt, schema.SQLVarchar, schema.SQLTimestamp,
				schema.SQLTimestamp, schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLInt, schema.SQLVarchar, schema.SQLVarchar},
		)
	case "table status":
		r, err = c.buildEmptyResultset(
			[]string{"Name", "Engine", "Version", "Row_format",
				"Rows", "Avg_row_length", "Data_length", "Max_data_length",
				"Index_length", "Data_free", "Auto_increment", "Create_time",
				"Update_time", "Check_time", "Collation", "Checksum",
				"Create_options", "Comment"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLInt, schema.SQLInt, schema.SQLInt,
				schema.SQLInt, schema.SQLInt, schema.SQLInt, schema.SQLBoolean,
				schema.SQLTimestamp, schema.SQLTimestamp, schema.SQLTimestamp, schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar},
		)
	case "triggers":
		r, err = c.buildEmptyResultset(
			[]string{"Trigger", "Event", "Table", "Statement",
				"Timing", "Created", "sql_mode", "Definer",
				"character_set_client", "collation_connection", "Database Collation"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLVarchar, schema.SQLTimestamp, schema.SQLVarchar,
				schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar, schema.SQLVarchar},
		)
	case "warnings":
		r, err = c.buildEmptyResultset(
			[]string{"Level", "Code", "Message"},
			[]schema.SQLType{schema.SQLVarchar, schema.SQLInt, schema.SQLVarchar},
		)
	case "count(*) warnings":
		r, err = c.buildEmptyResultset(
			[]string{"@@session.warning_count"},
			[]schema.SQLType{schema.SQLInt},
		)
	default:
		return mysqlerrors.Newf(mysqlerrors.ErNotSupportedYet, "no support for show (%s) for now",
			sql)
	}

	if err != nil {
		return err
	}

	return c.writeResultset(r)
}

func (c *conn) buildEmptyResultset(names []string, types []schema.SQLType) (*Resultset, error) {

	col, err := collation.Get(
		c.variables.GetCharset(variable.CharacterSetResults).DefaultCollationName)
	if err != nil {
		return nil, err
	}

	r := &Resultset{}

	valueKind := evaluator.GetSQLValueKind(c.Variables())
	for i := range names {
		field := &Field{
			Name:    util.Slice(names[i]),
			Charset: uint16(col.ID),
		}

		zeroValue := evaluator.SQLTypeToEvalType(types[i]).ZeroValue(valueKind)
		err = formatHeaderField(c.variables, field, zeroValue)
		if err != nil {
			return nil, err
		}

		r.Fields = append(r.Fields, field)
	}

	return r, nil
}

func (c *conn) writeResultset(r *Resultset) error {
	c.affectedRows = int64(-1)

	columnLen := putLengthEncodedInt(uint64(len(r.Fields)))

	data := make([]byte, 4, 1024)

	data = append(data, columnLen...)
	if err := c.writePacket(data); err != nil {
		return err
	}

	for _, v := range r.Fields {
		data = data[0:4]
		data = append(data, v.Dump(c.variables.GetCharset(variable.CharacterSetResults))...)
		if err := c.writePacket(data); err != nil {
			return err
		}
	}

	status := c.status()
	if err := c.writeEOF(status); err != nil {
		return err
	}

	for _, v := range r.RowDatas {
		data = data[0:4]
		data = append(data, v...)
		if err := c.writePacket(data); err != nil {
			return err
		}
	}

	return c.writeEOF(status)
}
