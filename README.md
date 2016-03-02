# sqlproxy

```
git clone git@github.com:10gen/sqlproxy.git
cd sqlproxy && . ./set_gopath.sh
go run main/main.go -config sample.conf
```

Connect using TCP protocol with MySQL client
```
mysql --protocol=tcp --port 3307
```
