import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../../../utils/testutils";
import { describe, test } from "vitest";

describe("Set Expression schema", () => {
  test("accepts Set expression", () => {
    // Cedar: when { 1 in [1, 2, 3] };
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              in: {
                left: { Value: 1 },
                right: {
                  Set: [{ Value: 1 }, { Value: 2 }, { Value: 3 }],
                },
              },
            },
          },
        ],
      })
    );
  });

  test("rejects Set expression with invalid element", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          { kind: "when", body: { Set: [{ Value: 1 }, { nope: true }] } },
        ],
      })
    );
  });

  test("rejects Set expression with additional top-level keys", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              Set: [{ Value: 1 }],
              Record: {},
            },
          },
        ],
      })
    );
  });
});
