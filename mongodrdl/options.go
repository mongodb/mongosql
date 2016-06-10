package mongodrdl

var Usage = `<options>

Export the content of a running server into .yaml files.

Specify a database with -d and a collection with -c to export the schema for only that database or collection.

See http://docs.mongodb.org/manual/reference/program/mongodrdl/ for more information.`

// OutputOptions defines the set of options for writing dump data.
type OutputOptions struct {
	CustomFilterField string `long:"customFilterField" short:"f" description:"the name of the field to use with a custom mongo filter field (defaults to no custom filter field)."`
	Out               string `long:"out" short:"o" description:"output file, or '-' for standard out (defaults to standard out)" default-mask:"-"`
}

// Name returns a human-readable group name for output options.
func (*OutputOptions) Name() string {
	return "output"
}
