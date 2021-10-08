# BI Connector

## Runbook
The [`runbook/`](./runbook) directory contains markdown files with
step-by-step instructions for a number of BI Connector processes:

- [Releasing the BI Connector](./runbook/release.md)
- [Adding or Removing Platforms](./runbook/platforms.md)
- [Adding or Removing MongoDB Version Support](./runbook/mongodb-versions.md)

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
If you use Go version 1.11 or 1.12 with `$GO111MODULE=on`, or if you use Go version >= 1.13,
you must set `$GOFLAGS=-mod=vendor`. You should also set `$GOBIN` and `$PATH`:
```
export GOFLAGS=-mod=vendor
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

If you use Go version 1.11 or 1.12 with `%GO111MODULE%` set to `on`, or if you use Go
version >= 1.13, you must set `%GOFLAGS%` to `-mod=vendor`.

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
unit tests, integration tests, and end-to-end (e2e) tests.

### Test Automation Concerns
When a new major version of MongoDB is released, the `LATEST` variable at the top of
`testdata/bin/download-mongodb.sh` must be updated.

### Unit Tests
Most of the packages in this repository include unit tests.
To run all unit tests, run `go test -v ./...`
or for an individual package `cd <package> && go test -v`.
To run all of the unit tests, you can also run `make clean test-unit`.

If you are running the unit tests manually, note that the following packages
include one or more unit tests that require a MongoDB instance to be running:

- `internal/sample`
- `mongodrdl`
- `evaluator`
- `mongodb`

The relevant test files are suffixed by "_integration_test.go".
By default, these tests will not run when you invoke `go test -v ./...`. To run
these module integration tests (together with the other unit tests), run:
```
go test -v ./... -tags=integration -run $(find . -name '*integration_test.go' | grep -v './vendor' | grep -v './testdata' | xargs -L1 dirname | uniq | sed 's/\.$//')
```

### Integration Tests
The BI Connector's internal integration tests verify that we return correct results for
a variety of SQL queries. Our integration tests are broken up into a number of
"suites": `blackbox`, `internal`, `tableau`, `tdvt`, and `tpch`. The
`internal` suite is written by the BI Connector dev team; the others are
third-party test suites adapted for use in our testing framework.

If you haven't already downloaded all of our integration test data, you can do
so by running `make download-data`.

Before running the integration tests, be sure to startup `mongosqld`:
```
mongosqld -vv
```

and `mongod`:
```
mongod
```

The entry point for our integration tests can be found in
`integration_test.go`. Each suite is a subtest of `TestIntegration`, and the
tests in a given suite are subtests of those subtests. The examples below
demonstrate how to run certain subsets of the integration tests:

```
# run all unit and integration tests
go test -v -tags=integration ./...

# run all integration tests
go test -v -tags=integration

# run all the tests in the internal integration suite
go test -v -tags=integration -run /internal

# run all tests in any suite that match the regex "where"
go test -v -tags=integration -run //where

# run test 123 from the blackbox integration suite
go test -v -tags=integration -run /blackbox/123
```

The integration tests also add a flag (`-automate`) that can be passed to
`go test`. The `-automate` flag allows the user to control which parts of the
test infrastructure should be set up automatically. The following values are
currently supported:

- `none` - Do not automate infrastructure setup. This is the default if `-automate` is not supplied.
- `data,schema` - Mongorestore data needed for the specified tests to MongoDB and sample the data.
- `data` - Mongorestore data needed for the specified tests to MongoDB without sampling.

For example, the following command will run all of the tdvt tests and ensure that
all the data needed for those tests is inserted into the running MongoDB
instance:

```
go test -v -tags=integration -run /tdvt/ -automate data,schema
```

In the future, we hope to support automating mongod, mongosqld, and various
other parts of our test infrastructure via the `-automate` flag. In the meantime,
the helper scripts in `testdata/bin` can be used to spin up those components.

### e2e Tests
The BI Connector's end-to-end tests verify various mongosqld behaviors under many
permutations of mysql client, mongosqld, and MongoDB configuration options. Each
e2e test is a make target in one of the `*.mk` files found at `testdata/e2e/tests`.
To run an e2e test, run `make clean <target>` from the root of the repository.

## Customizations
If you prefer a file-based schema approach, you can generate and use a `.drdl` file:
```
mongodrdl -d <database-name> -o schema.drdl
mongosqld --schema schema.drdl
```

For more a comprehensive set of startup customizations, you can pass in a config file:
```
# This file is included in the repo, and bundled in our release package.
mongosqld --config release/distsrc/example-mongosqld-config.yml
```

## Dependencies
This repository uses [Go modules](https://github.com/golang/go/wiki/Modules) for dependency
management, and uses vendored dependencies for building and testing. Occasionally, contributors
may need to add or update dependencies. If you need to add or update a dependency and you use Go
version 1.11 or 1.12, you must set `GO111MODULE=on` in your environment. As noted above, you
must also set `GOFLAGS=-mod=vendor`.

#### Adding a new dependency
If you need to add a new dependency to the project, you can simply add import statements to
the relevant `.go` code. There is no need to `go get`.

Typically, when you build the code (via `go build`, `go test`, etc) the new dependency will be
identified automatically by the Go build system and downloaded to the module cache, if necessary.
However, because this repository uses vendored dependencies, you **must** revendor before building:
```
go mod vendor
```
This command will update the `go.mod` and `go.sum` files and will update the `vendor/` directory
with the new dependency.

#### Updating a dependency
There are two main ways you may need to update a dependency: using previously unused packages
from an existing dependency, or updating the version of an existing dependency.

To use a previously unused package, all you need to do is import it as necessary in the source
code, and then `go mod vendor` to update the `vendor/` directory. The modules' vendor feature is
clever and does not vendor unused packages, which is why you must revendor even when using an
existing dependency. This is only necessary if the package does not already exist in the `vendor/`
directory, which you can easily verify.

To update the version of an existing dependency, you need to update the `go.mod` file. Find the
relevant dependency in the `require` list and update the version, and then `go mod vendor` to update
the `vendor/` directory. The new version can be a specific release version (i.e. v0.1.2), or can
be a specific git commit (see the `go.mod` file for examples). You can also specify "master",
which will be replaced with the relevant version info when you revendor.

## Documentation
See the BI Connector [documentation](https://docs.mongodb.com/bi-connector/master/).
