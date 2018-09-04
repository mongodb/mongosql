# BI Connector

## Components
`mongosqld` - The proxy process.

`mongodrdl` - The schema generator.

For `mongosqld` command line options invoke:

    $ mongosqld --help

For `mongodrdl` command line options invoke:

    $ mongodrdl --help

## Downloading the BI Connector
You can download the BI Connector in the `sign` task from [this project's Evergreen waterfall page](https://evergreen.mongodb.com/waterfall/sqlproxy).

## Building the BI Connector
### Unix
Download [Go](https://golang.org/dl/) version >= 1.8.1.  

Refer to [this guide](https://github.com/golang/go/wiki/Setting-GOPATH#unix-systems) for setting `$GOPATH`.   
Note that on Linux systems, you should elect to use `~/.bashrc` instead of `~/.bash_profile`.
You should also set `$GOBIN` and `$PATH`:
```
export GOBIN="$GOPATH/bin"
export PATH="$GOBIN:$PATH"
```
After this, you can download the repo and build it as such:
```
git clone git@github.com:10gen/sqlproxy.git $GOPATH/src/github.com/10gen/sqlproxy
cd $GOPATH/src/github.com/10gen/sqlproxy
go install cmd/mongosqld/mongosqld.go
```

### Windows
Download [Go](https://golang.org/dl/) version >= 1.8.1.

Refer to [this guide](https://github.com/golang/go/wiki/Setting-GOPATH#windows) for setting `%GOPATH%`.  
You should then set `%GOBIN%` to `%GOPATH%\bin`.  
You will also need to update your `PATH` to:
- `%GOBIN%`, if it was originally blank, or
- `%GOBIN%;<other-contents>` otherwise. 

After this, you can download the repo and build it as such:
```
git clone git@github.com:10gen/sqlproxy.git "%GOPATH%\src\github.com\10gen\sqlproxy"
cd "%GOPATH%\src\github.com\10gen\sqlproxy"
go install cmd\mongosqld\mongosqld.go
```

## Running the BI Connector
Simply run the connector binary:
```
# Start the connector with automatic schema sampling
mongosqld
```

You can then connect to it using any client that communicates with MySQL's wire protocol.

For example, using the MySQL CLI's TCP protocol:
```
mysql --protocol=tcp --port 3307
```
or via its UNIX sockets:
```
# mongosqld binds to the socket at /tmp/mysql.sock by default
mysql
```

## Testing the BI Connector
There are three different categories into which the BI Connector's tests fall:
unit tests, integration tests, and configuration parameter (config) tests.

### Unit Tests
Most of the packages in this repository include unit tests.
To run the tests for a particular package, run `go test ./<package>`
or `cd <package> && go test`.
To run all of the unit tests, you can also run `make clean test-unit`.

If you are running the unit tests manually, note that the following packages
include one or more unit tests that require a MongoDB instance to be running:

- `internal/sample`
- `mongodrdl`
- `evaluator`
- `mongodb`

### Integration Tests
The BI Connector's internal integration tests verify that we return correct results for
a variety of SQL queries. Our integration tests are broken up into a number of
"suites": `blackbox`, `internal`, `tableau`, `tdvt`, and `tpch`. The
`internal` suite is written by the BI Connector dev team; the others are
third-party test suites adapted for use in our testing framework.

If you haven't already dowloaded all of our integration test data, you can do
so by running `make download-data`.

The integration tests belong to the `sqlproxy` package, and can be found in
`integration_test.go`. Each suite is a subtest of `TestIntegration`, and the
tests in a given suite are subtests of those subtests. The examples below
demonstrate how to run certain subsets of the integration tests:

```
# run all the tests in the sqlproxy package
# currently, this only includes TestIntegration
go test

# run TestIntegration and all its subtests
go test -run TestIntegration

# run all the tests in the internal integration suite
go test -run /internal

# run all tests in any suite that match the regex "where"
go test -run //where

# run test 123 from the blackbox suite
go test -run /blackbox/123
```

The integration tests also add a new flag (`-automate`) that can be passed to
`go test`. The `-automate` flag allows the user to control which parts of the
test infrastructure should be set up automatically. The following values are
currently supported:

- `none` - Do not automate infrastructure setup. This is the default if `-automate` is not supplied.
- `data` - Mongorestore data needed for the specified tests to MongoDB.

For example, the following command will run all of the tdvt tests and ensure that
all the data needed for those tests is inserted into the running MongoDB
instance:

```
go test -run /tdvt/ -automate data
```

In the future, we hope to support automating mongod, mongosqld, and various
other parts of our test infrastructure via the `-automate` flag. In the meantime,
the helper scripts in `testdata/bin` can be used to spin up those components.

### Config Tests
The BI Connector's config tests are end-to-end tests that verify various
mongosqld behaviors under many permutations of mysql client, mongosqld, and
MongoDB configuration options. Each config test is a make target in one of the
`*.mk` files found at `testdata/config/tests`. To run a config test, run
`make clean <target>` from the root of the repository.

## Customizations
If you prefer a file-based schema approach, you can generate and use a `.drdl` file:
```
mongodrdl -d <database-name> -o schema.drdl
mongosqld --schema schema.drdl
```

For more a comprehensive set of startup customizations, you can pass in a config file:
```
# This file is included in the repo
mongosqld --config testdata/resources/configs/config.yml
```

## Documentation
See the BI Connector [documentation](https://docs.mongodb.com/bi-connector/master/).
