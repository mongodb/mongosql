# Sampler in Rust!!

The goals of this project are:
- Partition a collection based on _id, much like mongosync
- On each partition, run the following algorithm
  1. Sample 100 documents
  2. Calculate their schema
  3. Find documents that _do not match_ the schema
  4. Calculate the schema for any returned documents
  5. Merge the schemas
  6. Repeat 3-5 until no new documents are returned
- Merge each partition schema into a single schema
