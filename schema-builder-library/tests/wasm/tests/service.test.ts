import { afterEach, beforeAll, beforeEach, describe, expect, test } from "vitest";
import { MongoClient } from "mongodb";

import { getCredentials, MongoDataService } from "./common.js";

const { URI } = getCredentials();

// Create and drop the client between tests
let CLIENT: MongoClient;
let SERVICE: MongoDataService;
beforeEach(async () => {
    CLIENT = await new MongoClient(URI).connect();
    SERVICE = new MongoDataService(CLIENT);
});
afterEach(async () => CLIENT.close());

// Tests
describe("service impl", () => {
    test("listing databases works", async () => {
        expect(await SERVICE.listDatabases()).toMatchInlineSnapshot(`
          [
            "admin",
            "config",
            "local",
            "nonuniform",
            "uniform",
          ]
        `);
    });

    test("listing collections works", async () => {
        expect((await SERVICE.listCollections("uniform")).map(coll => coll.name)).toMatchInlineSnapshot(`
          [
            "small",
            "large",
            "unit",
            "system.views",
            "__sql_schemas",
            "test_view",
          ]
        `);
    })
})
