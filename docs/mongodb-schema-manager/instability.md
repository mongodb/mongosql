# Schema Instability

The mongodb-schema-manager looks for what we call schema instability. Schema instability occurs when
too many fields appear that differ from document to document in the collection. This is a general
design pattern some MongoDB users use where the documents are treated as hash maps instead of as
structured data. When this occurs, the mongodb-schema-manager will cap the number of fields in the
collection schema and produce a warning. This design pattern is not acceptable for SQL because each
top level field is treated as a column in SQL.
