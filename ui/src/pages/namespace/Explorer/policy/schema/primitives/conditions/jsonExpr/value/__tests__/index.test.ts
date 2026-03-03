import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../../../utils/testutils";
import { describe, test } from "vitest";

describe("Value JsonExpr schema", () => {
  test("accepts arbitrary JSON value", () => {
    expectValidPolicy(
      createBasePolicy({
        conditions: [{ kind: "when", body: { Value: { nested: [1, null] } } }],
      })
    );
  });

  test("rejects extra top-level keys", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              Value: true,
              extra: true,
            },
          },
        ],
      })
    );
  });
});
