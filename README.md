# BI Connector



### Components
`mongosqld` - The proxy process.

`mongodrdl` - The schema generator.

For `mongosqld` command line options invoke:

    $ mongosqld --help

For `mongodrdl` command line options invoke:

    $ mongodrdl --help

### Downloading the BI Connector
You can download the BI Connector in the `sign` task from [this project's Evergreen waterfall page](https://evergreen.mongodb.com/waterfall/sqlproxy).



### Building the BI Connector: Unix

Download [Go](https://golang.org/dl/) version >= 1.8.1.  

Refer to [this guide](https://github.com/golang/go/wiki/Setting-GOPATH#unix-systems) for setting `$GOPATH`.   
Note that on Linux systems, you should elect to use `~/.bashrc` instead of `~/.bash_profile`.
You should also set `$GOBIN` and `$PATH`:
```
export $GOBIN=$GOPATH\bin
export $PATH=$GOBIN:$PATH
```
After this, you can download the repo and build it as such:
```
git clone git@github.com:10gen/sqlproxy.git $GOPATH/src/github.com/10gen/sqlproxy
cd $GOPATH/src/github.com/10gen/sqlproxy
go install main/sqlproxy.go
```



### Building the BI Connector: Windows

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
go install main\sqlproxy.go
```



### Running the BI Connector
Simply run the connector binary:
```
# Start the connector with automatic schema sampling
sqlproxy
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


### Customizations

If you prefer a file-based schema approach, you can generate and use a `.drdl` file:
```
mongodrdl -d <database-name> -o schema.drdl
sqlproxy --schema schema.drdl
```


For more a comprehensive set of startup customizations, you can pass in a config file: 
```
# This file is included in the repo
sqlproxy --config testdata/resources/configs/sample.yml
```




### Documentation
See the BI Connector [documentation](https://docs.mongodb.com/bi-connector/master/).