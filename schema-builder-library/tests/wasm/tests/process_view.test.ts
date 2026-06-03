import { afterEach, beforeAll, beforeEach, describe, expect, test } from "vitest";
import { MongoClient, type Document } from "mongodb";
import { init, process_view } from "schema-builder-library";

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
describe("view schema generation", () => {
    const VIEW_NAME = "test_view";

    test(`generates a valid schema`, async () => {
        await expect(toEJSON(await process_view(SERVICE, "uniform", VIEW_NAME) as Document))
            .toMatchFileSnapshot(`../snaps/uniform.${VIEW_NAME}.schema.json`);
    });
});
