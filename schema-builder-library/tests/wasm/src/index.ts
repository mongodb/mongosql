import type { SqlDataService, SqlCursor, BsonDocument, AggregateOptions } from "schema-builder-library";

import { EJSON, type Document } from "bson";
import { MongoClient } from "mongodb";
import { init, process_collection } from "schema-builder-library";

const URI = "mongodb://localhost:27017";
const DB = "sample_mflix";
const COLLECTION = "movies";

/**
 * Convert native BSON documents to EJSON format for WASM.
 * This ensures BSON types like ObjectId, Date, etc. are properly serialized
 * as EJSON objects that can be deserialized by Rust's bson crate.
 */
function toEJSON(doc: Document): BsonDocument {
    // EJSON.serialize converts BSON types to their EJSON representation
    // e.g., ObjectId("...") -> { "$oid": "..." }
    // Canonical (relaxed: false) preserves Int32 vs Int64 distinction
    // across the wasm boundary; without it ints widen to Int64 in Rust.
    return EJSON.serialize(doc, { relaxed: false }) as BsonDocument;
}

/**
 * Convert EJSON format from WASM back to native BSON documents.
 * This ensures EJSON objects like { "$oid": "..." } are converted back
 * to native BSON types like ObjectId for the MongoDB driver.
 */
function fromEJSON(doc: BsonDocument): Document {
    // EJSON.deserialize converts EJSON representation back to BSON types
    // e.g., { "$oid": "..." } -> ObjectId("...")
    return EJSON.deserialize(doc as Document) as Document;
}

/**
 * MongoDB data service using the typescript MongoDB client
 */
class MongoDataService implements SqlDataService {
    client: MongoClient;

    constructor(client: MongoClient) {
        this.client = client;
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
        let cursor = this.client.db(dbName).collection(collName).aggregate(nativePipeline, options.keyHint ? { hint: options.keyHint as any } : {});

        return new Cursor(cursor);
    }
}

/**
 * Cursor over MongoSQL entries
 */
class Cursor implements SqlCursor {
    cursor: SqlCursor;

    constructor(cursor: SqlCursor) {
        this.cursor = cursor;
    }

    async next(): Promise<BsonDocument | null> {
        let next = await this.cursor.next();
        if (next == null) {
            return null;
        }

        return toEJSON(next);
    }

}

async function main() {
    init();

    console.log("Connecting to client...");
    let client = await new MongoClient(URI).connect();
    console.log("> Connected");

    let service = new MongoDataService(client);
    try {
        console.log(`Deriving schema for ${DB}.${COLLECTION}...`);

        let start = new Date().getTime();
        let schema = await process_collection(service, DB, COLLECTION);
        let end = new Date().getTime();

        console.log(`> Schema (took ${(end - start) / 1000}s):`, EJSON.stringify(schema, undefined, 2));
    } catch (e) {
        console.error("Schema derivation failed:", e);
    } finally {
        await client.close();
    }
}

await main();
