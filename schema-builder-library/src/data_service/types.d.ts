/**
 * BSON Document type - a generic key-value object.
 * This is equivalent to MongoDB's Document type.
 */
export type BsonDocument = Record<string, unknown>;

/**
 * Options for timeseries collections.
 */
export interface TimeSeriesOptions {
    timeField: string;
    metaField?: string;
}

/**
 * Options for a collection entry, varying by collection type.
 */
export interface CollectionInfoOptions {
    /** For views, the source collection name */
    viewOn: string;
    /** For views, the aggregation pipeline */
    pipeline: BsonDocument[];
    /** For timeseries collections, the timeseries options */
    timeseries: TimeSeriesOptions;
}

/**
 * Information about a collection.
 */
export interface CollectionInfo {
    /** The name of the collection */
    name: string;
    /** The type: "collection", "view", or "timeseries" */
    type: string;
    /** Type-specific options; fields vary by collection type */
    options?: Partial<CollectionInfoOptions>;
}

/**
 * Options for aggregate queries
 */
export interface AggregateOptions {
    /** A hint for which key to use for indexing */
    keyHint: BsonDocument,
}

/**
 * SqlDataService interface for database operations.
 *
 * Implement this interface in your JavaScript/TypeScript code to provide
 * database access to the schema builder.
 */
export interface SqlDataService {
    /** List all database names */
    listDatabases(): Promise<string[]>;
    /** List all collections in a database */
    listCollections(dbName: string): Promise<CollectionInfo[]>;
    /** Execute an aggregation pipeline on a collection */
    aggregate(dbName: string, collName: string, pipeline: BsonDocument[], options: Partial<AggregateOptions>): Promise<SqlCursor>;
    /** Execute a find query on a collection */
    find(dbName: string, collName: string, filter: BsonDocument): Promise<SqlCursor>;
}

/**
 * SqlCursor interface for streaming database results
 *
 * Implement this interface in your JavaScript/TypeScript code to provide a
 * way to stream over database results
 */
export interface SqlCursor {
    /** Get the next element in this cursor. Return undefined to signal the end of the stream */
    next(): Promise<BsonDocument | undefined>;
}
