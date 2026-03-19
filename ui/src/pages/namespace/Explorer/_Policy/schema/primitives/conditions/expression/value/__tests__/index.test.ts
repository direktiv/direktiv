import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../../../utils/testutils";
import { describe, test } from "vitest";

describe("Value Expression schema", () => {
  test("accepts arbitrary JSON value", () => {
    // Cedar: when { context.request == {"nested": [1, null]} };
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              "==": {
                left: { ".": { left: { Var: "context" }, attr: "request" } },
                right: { Value: { nested: [1, null] } },
              },
            },
          },
        ],
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
              // @ts-expect-error - value expressions are strict
              extra: true,
            },
          },
        ],
      })
    );
  });
});
