import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../../../utils/testutils";
import { describe, test } from "vitest";

describe("Attribute Expression schema", () => {
  test("accepts dot accessor expression", () => {
    // Cedar: when { context.tls_version == "1.3" };
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              "==": {
                left: {
                  ".": { left: { Var: "context" }, attr: "tls_version" },
                },
                right: { Value: "1.3" },
              },
            },
          },
        ],
      })
    );
  });

  test("accepts has accessor expression", () => {
    // Cedar: when { principal has email };
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: { has: { left: { Var: "principal" }, attr: "email" } },
          },
        ],
      })
    );
  });

  test("rejects attribute expression without attr", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          // @ts-expect-error - attribute expressions require attr
          { kind: "when", body: { ".": { left: { Var: "context" } } } },
        ],
      })
    );
  });

  test("rejects non-string attr value", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            // @ts-expect-error - attr must be a string
            body: { has: { left: { Var: "principal" }, attr: 1 } },
          },
        ],
      })
    );
  });
});
