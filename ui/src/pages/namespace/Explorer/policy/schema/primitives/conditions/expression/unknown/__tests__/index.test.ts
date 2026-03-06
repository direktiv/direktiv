import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../../../utils/testutils";
import { describe, test } from "vitest";

describe("Unknown Expression schema", () => {
  test("accepts Unknown with exactly one string argument", () => {
    // Cedar: when { unknown("x") == context.lookupName };
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              "==": {
                left: { Unknown: { name: "x" } },
                right: {
                  ".": { left: { Var: "context" }, attr: "lookupName" },
                },
              },
            },
          },
        ],
      })
    );
  });

  test("rejects Unknown with multiple keys", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [{ kind: "when", body: { Unknown: { a: "x", b: "y" } } }],
      })
    );
  });

  test("rejects Unknown with non-string values", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [{ kind: "when", body: { Unknown: { a: 1 } } }],
      })
    );
  });
});
