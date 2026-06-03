import type { SqlDataService, SqlCursor, BsonDocument, AggregateOptions } from "schema-builder-library";

import { EJSON, type Document } from "bson";
import { type AggregationCursor, type FindCursor, MongoClient } from "mongodb";

/**
 * Convert native BSON documents to EJSON format for WASM.
 * This ensures BSON types like ObjectId, Date, etc. are properly serialized
 * as EJSON objects that can be deserialized by Rust's bson crate.
 */
export function toEJSON(doc: Document): BsonDocument {
    // EJSON.serialize converts BSON types to their EJSON representation
    // e.g., ObjectId("...") -> { "$oid": "..." }
    //
    // Note: We set relaxed to `false` here because we want to keep the types in
    // their native BSON form since we are sending documents across the WASM
    // boundary, which has no notion of JS types. Without this, the native JS
    // types for numerical ints was being propagates instead, leading to aggregations
    // referencing non-existant types.
    return EJSON.serialize(doc, { relaxed: false }) as BsonDocument;
}

/**
 * Convert EJSON format from WASM back to native BSON documents.
 * This ensures EJSON objects like { "$oid": "..." } are converted back
 * to native BSON types like ObjectId for the MongoDB driver.
 */
export function fromEJSON(doc: BsonDocument): Document {
    // EJSON.deserialize converts EJSON representation back to BSON types
    // e.g., { "$oid": "..." } -> ObjectId("...")
    return EJSON.deserialize(doc as Document) as Document;
}

/**
 * MongoDB data service using the typescript MongoDB client
 */
export class MongoDataService implements SqlDataService {
    client: MongoClient;

    constructor(client: MongoClient) {
        this.client = client;
    }

    async listDatabases(): Promise<string[]> {
        // The result shape of calling the `listDatabases` command.
        interface ListDbResult {
            databases: [{ name: string }],
        }

        let dbs = await this.client
            .db("admin")
            .command({ listDatabases: 1, nameOnly: true }) as ListDbResult;
        return dbs.databases.map(db => db.name);
    }

    async listCollections(dbName: string): Promise<any[]> {
        const cursor = this.client.db(dbName).listCollections();
        const results = [];

        for await (const collection of cursor) {
            results.push(collection);
        }

        return results;
    }

    async aggregate(dbName: string, collName: string, pipeline: BsonDocument[], options: Partial<AggregateOptions>): Promise<SqlCursor> {
        const nativePipeline = pipeline.map(fromEJSON);
        let cursor = this.client
            .db(dbName)
            .collection(collName)
            .aggregate(nativePipeline, {
                hint: options.keyHint as any,

                // ### SUPER IMPORTANT ###
                // This flag ensures that the node driver doesn't randomly narrow long types into
                // integers. If this is set to the default of `true` aggregate pipelines which try
                // to narrow the result set using a schema will never complete.
                promoteLongs: false,
            });

        return new EJSONCursor(cursor);
    }

    async find(dbName: string, collName: string, filter: BsonDocument): Promise<SqlCursor> {
        const nativeFilter = fromEJSON(filter);
        let cursor = this.client.db(dbName).collection(collName).find(nativeFilter);

        return new EJSONCursor(cursor);
    }
}

/**
 * Cursor over MongoSQL entries
 */
class EJSONCursor implements SqlCursor {
    cursor: AggregationCursor | FindCursor;

    constructor(cursor: SqlCursor) {
        this.cursor = cursor;
    }

    async next(): Promise<BsonDocument | null> {
        let next = (await this.cursor.next()) as Document;
        if (!next) {
            return null;
        }

        return toEJSON(next);
    }
}

/**
 * Test Credentials
 */
export interface Credentials {
    URI: string,
}

/**
 * Get the credentials for this test from the following environment variables:
 *
 * - WASM_TEST_URI: The URI, with connection information, to the MongoDB instance
 *
 * Note that this uses the following defaults if not overwritten:
 *
 * - mongodb://localhost:27017
 */
export function getCredentials(): Credentials {
    return {
        URI: process.env["WASM_TEST_URI"] ?? "mongodb://localhost:27017",
    };
}
