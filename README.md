# sqlproxy

```
git clone git@github.com:10gen/sqlproxy.git $GOPATH/src/github.com/10gen/sqlproxy
cd $GOPATH/src/github.com/10gen/sqlproxy && ./build.sh
./bin/mongosqld --schema sample.conf
```
<sub>need go version >= 1.5*</sub>

Connect using TCP protocol with MySQL client
```
mysql --protocol=tcp --port 3307
```
Connect using UNIX socket with MySQL client
```
mysql --socket /tmp/mongosqld-3307.sock
```
