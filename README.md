# BI Connector

### Components
`mongosqld` - The proxy process.
`mongodrdl` - The schema generator.

For `mongosqld` command line options invoke:

    $ mongosqld --help

For `mongodrdl` command line options invoke:

    $ mongodrdl --help

### Building the BI Connector
You can download the BI Connector from the `dist` task at https://evergreen.mongodb.com/waterfall/sqlproxy
or build it (requires [Go](https://golang.org/dl/) version >= 1.5) from source using:
```
git clone git@github.com:10gen/sqlproxy.git $GOPATH/src/github.com/10gen/sqlproxy
cd $GOPATH/src/github.com/10gen/sqlproxy && ./build.sh
```

### Running the BI Connector
First start the connector using:
```
# start mongosqld with the sample DRDL file in this repo
./bin/mongosqld --schema sample.conf 
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

### Documentation
See the BI Connector [documentation](https://docs.mongodb.com/bi-connector/master/).
