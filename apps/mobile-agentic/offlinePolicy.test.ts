import { strict as assert } from "node:assert";
// @ts-expect-error Node's strip-types runner resolves the source extension directly.
// eslint-disable-next-line import/extensions
import { shouldQueueOffline } from "./offlinePolicy.ts";

assert.equal(shouldQueueOffline(new Error("network down")), true);
assert.equal(shouldQueueOffline({ status: 409 }), false);
assert.equal(shouldQueueOffline({ status: 422 }), false);
assert.equal(shouldQueueOffline({ status: 503 }), false);
assert.equal(shouldQueueOffline({ message: "not an HTTP response" }), true);

console.log("offlinePolicy: PASS");
