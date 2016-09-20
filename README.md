# sqlproxy

```
git clone git@github.com:10gen/sqlproxy.git $GOPATH/src/github.com/10gen/sqlproxy
cd $GOPATH/src/github.com/10gen/sqlproxy && ./build.sh
./bin/mongosqld --schema sample.conf
```

Connect using TCP protocol with MySQL client
```
mysql --protocol=tcp --port 3307
```
