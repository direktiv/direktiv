import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../../../utils/testutils";
import { describe, test } from "vitest";

describe("Extension Expression schema", () => {
  test("accepts extension function expression", () => {
    // Cedar: when { decimal("10.0") == context.maxAmount };
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              "==": {
                left: { decimal: [{ Value: "10.0" }] },
                right: { ".": { left: { Var: "context" }, attr: "maxAmount" } },
              },
            },
          },
        ],
      })
    );
  });

  test("rejects extension expression with non-array args", () => {
    expectInvalidPolicy(
      createBasePolicy({
        // @ts-expect-error - extension args must be an array
        conditions: [{ kind: "when", body: { decimal: { Value: "10.0" } } }],
      })
    );
  });

  test("rejects extension expression with multiple keys", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              decimal: [{ Value: "10.0" }],
              ip: [{ Value: "222.222.222.0/24" }],
            },
          },
        ],
      })
    );
  });

  test("rejects extension expression that uses a reserved key", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [{ kind: "when", body: { is: [{ Value: "10.0" }] } }],
      })
    );
  });
});
