# mongo-sql-temp

```
cd $GOROOT/src/github.com/10gen/
git clone git@github.com:erh/mongo-sql-temp.git sqlproxy
make bootstrap
make test
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
