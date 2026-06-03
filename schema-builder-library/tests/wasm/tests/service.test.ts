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
        let dbs = await SERVICE.listDatabases();

        expect(dbs.sort()).toMatchInlineSnapshot(`
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
        let collections = await SERVICE.listCollections("uniform");
        let names = collections.map(coll => coll.name).sort();

        expect(names).toMatchInlineSnapshot(`
          [
            "__sql_schemas",
            "large",
            "small",
            "system.views",
            "test_view",
            "unit",
          ]
        `);
    })
})
