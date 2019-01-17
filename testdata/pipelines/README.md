# Pushdown Metrics Pipelines

This is a collection of aggregation pipelines that can be used to create views suitable for charting in a MongoDB cluster holding BI Connector pushdown metrics data.   

These views can be used to answer questions such as:

* Most frequent pushdown failure reasons
* Most frequently used scalar functions
* Percentage of queries using unions or subqueries
* min/max/avg query latency

## Getting Started

Create a Stitch project in Atlas, configure a database and collection for BIC metrics, and configure BI Connector to send pushdown metrics to Stitch.  For example, the following configuration, with the appropriate STITCH_APP_ID, is used to send data to the bic-prod cluster in Atlas.

```
## Pushdown metrics logging
setParameter:
  metrics_backend: stitch
metrics:
    stitchURL: "https://webhooks.mongodb-stitch.com/api/client/v2.0/app/<STITCH_APP_ID>/service/atlas/incoming_webhook/query-post-test"
```

Create views for each of the pipelines in this folder on the metrics collection using Mongoshell.  You can then add the views as data sources in Charts.

Happy Charting!
