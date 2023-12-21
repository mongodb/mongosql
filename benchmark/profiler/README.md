# Profiler  
## Queries
Add queries in the `src/config_loader/queries` directory, one query per file.   
Format:
```yaml
db: database
query: |
  SELECT
      *
  FROM
      t1
```
**db**: Matches the name of the catalog in the `catalogs/` directory
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