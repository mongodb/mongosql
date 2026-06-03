import { afterEach, beforeAll, beforeEach, describe, expect, test, vi } from "vitest";
import { MongoClient, type Document } from "mongodb";
import { init, process_collection } from "schema-builder-library";

import { getCredentials, MongoDataService, toEJSON } from "./common.js";

const { URI } = getCredentials();

// We need the WASM context to be ready before we run anything
beforeAll(init);

// Create and drop the client between tests
let CLIENT: MongoClient;
let SERVICE: MongoDataService;
beforeEach(async () => {
    CLIENT = await new MongoClient(URI).connect();
    SERVICE = new MongoDataService(CLIENT);
});
afterEach(async () => CLIENT.close());

// Tests
describe("collection schema generation", () => {
    test(`generates a valid schema on single entry`, async () => {
        await expect(toEJSON(await process_collection(SERVICE, "uniform", "unit") as Document))
            .toMatchFileSnapshot(`../snaps/uniform.unit.schema.json`);
    });

    test(`generates a valid schema on small collection`, async () => {
        await expect(toEJSON(await process_collection(SERVICE, "uniform", "small") as Document))
            .toMatchFileSnapshot(`../snaps/uniform.small.schema.json`);
    });

    test(`does not sample more than it needs to`, async () => {
        const aggregateSpy = vi.spyOn(SERVICE, "aggregate");
        expect(aggregateSpy).not.toHaveBeenCalled();

        // Verify that we only ran the aggregation pipeline exactly 5 times:
        // - Get the size of the collection
        // - Get the starting index
        // - Get the end index
        // - Get the first few elements
        // - No longer get any other elements since they'll be filtered out by the schema
        await process_collection(SERVICE, "uniform", "small");
        expect(aggregateSpy).toHaveBeenCalledTimes(5);
    });
});
