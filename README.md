# mongo-sql-temp

```
git clone git@github.com:erh/mongo-sql-temp.git sqlproxy
. ./set_gopath.sh
go run cmd/bi-connector/main.go -config sample.conf
```

Connect using TCP protocol with mysql client
```
mysql --protocol=tcp
```

TODO

- AS
   - field list
   - table list
- WHERE
   - math
- AGGREGATORS (avg, sum)
- SORT
- HAVING
- join
   - subquery
   - in FROM   
- pushdown
   - group
   - sort
   - $lookup
