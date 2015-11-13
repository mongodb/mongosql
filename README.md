# sqlproxy

```
git clone git@github.com:10gen/sqlproxy.git
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
- Run select_test.go as integration tests
- Add support for arrays
- Add support for TIMESTAMP parsing and processing
- Finalize data type mapping between MongoDB and MySQL types
- Support all expressions in 'dual' queries.
- Add authentication support
- Add SSL support
- Add support for disk support
- Add query_planner.go tests
- Generate config on the fly (use connection stream to infer)
- Stream results to client
- WHERE
   - ~~math~~
   - ~~case expression~~
   - exists expression
   - strval expression
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
   - limit
   - $lookup
