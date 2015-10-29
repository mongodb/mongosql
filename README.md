# mongo-sql-temp

```
git clone git@github.com:erh/mongo-sql-temp.git sqlproxy
cd sqlproxy && . ./set_gopath.sh
go run main/main.go -config sample.conf
```

Connect using TCP protocol with mysql client
```
mysql --protocol=tcp
```

TODO

- ~~AS~~
   - ~~field list~~
   - ~~table list~~
- ~~dotted-field mapping/selection~~
- run all tests in select_test.go in roundtrip_test.go
- consolidate NewSQLValue and BuildMatcher
- WHERE
   - ~~math~~
   - case expression
   - exists expression
   - strval expression
   - nullcheck expression
   - subquery expression
   - value argument expressions
- ~~AGGREGATORS (avg, sum)~~
- ~~SORT~~
- ~~HAVING~~
- ~~join~~
   - ~~subquery~~
   - ~~in FROM~~
- pushdown/optimizations
   - group
   - sort
   - $lookup
