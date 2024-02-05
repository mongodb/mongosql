# Profiler
The `profiler` is used to compile SQL queries. We use the `profiler` in the
Memory Usage Profiler evergreen build variant, using `heaptrack`, to check the
memory usage of the queries in the `src/config_loader/queries` directory.

## Queries
Add queries in the `src/config_loader/queries` directory, one query per file.   
Format:
```yaml
db: database
skip_reason: string
query: |
  SELECT
      *
  FROM
      t1
```
**db**: Matches the name of the catalog in the `catalogs/` directory

**skip_reason**: *(optional)* A string specifying the reason this test is skipped.
Should be associated with a SQL ticket.

**query**: Format is a multi-line string

## Catalogs
Add catalogs in the `src/config_loader/catalogs` directory.
Format:
```json
{
  "catalog_schema": {
    <schema>
  }
}
```

## Running
Profiler takes one argument, which is the query to run.  
If not specified, default is `sample_analytics`.

```bash
$ profiler query1
```

## Important Notes Related to Leaked Memory
Note that the Memory Usage Profiler tasks mostly all report leaked memory. The
source of these leaks is known to be our lazy_static variables. We use lazy_static
variables to avoid creating large data structures in memory until absolutely
necessary. It is known and expected for this memory to not be cleaned up for the
rest of the duration of the program -- once it is finally instantiated, it is kept
alive for reuse. The OS reclaims this memory as soon as the program terminates.

Given this, we have decided to accept a small memory leak threshold -- the
`mem_usage_mem_leak_limit` variable in the evergreen project. Of course, this
leaves us in the vulnerable position of not knowing when non-lazy_static memory
is leaked. The threshold is a low as can be to account for all lazy_static data.

Some ideas to consider in case leaks become a more apparent issue in the future are
to create longer-running tasks that run multiple queries over a long time period to
see if more memory is leaked, or to manually invoke garbage collection periodically
to clean up these values. Another alternative is to stop using lazy_static entirely
since ADF is a hosted service that is always up. In that environment, the odds of
every lazy_static variable being instantiated is 1, so perhaps lazy_static is not
beneficial. These are future work ideas to consider in case we see that leaked memory
causes problems in production for ADF.
