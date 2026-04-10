/**
 * BSON Document type - a generic key-value object.
 * This is equivalent to MongoDB's Document type.
 */
export type BsonDocument = Record<string, unknown>;

/**
 * Information about a collection.
 */
export interface CollectionInfo {
    /** The name of the collection */
    name: string;
    /** The type: "collection", "view", or "timeseries" */
    type: string;
    options?: {
        /** For views, the source collection name */
        viewOn?: string;
        /** For views, the aggregation pipeline */
        pipeline?: BsonDocument[];
        /** For timeseries, the timeseries options */
        timeseries?: {
            timeField: string;
            metaField?: string;
        };
    };
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
    aggregate(dbName: string, collName: string, pipeline: BsonDocument[]): Promise<BsonDocument[]>;
    /** Execute a find query on a collection */
    find(dbName: string, collName: string, filter: BsonDocument): Promise<BsonDocument[]>;
}
