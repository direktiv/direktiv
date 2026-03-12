import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../../../utils/testutils";
import { describe, test } from "vitest";

describe("IfThenElse Expression schema", () => {
  test("accepts if-then-else expression", () => {
    // Cedar: when { if context then true else false };
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              "if-then-else": {
                if: { Var: "context" },
                then: { Value: true },
                else: { Value: false },
              },
            },
          },
        ],
      })
    );
  });

  test("rejects if-then-else expression without else", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              // @ts-expect-error - if-then-else requires if/then/else
              "if-then-else": { if: { Var: "context" }, then: { Value: true } },
            },
          },
        ],
      })
    );
  });

  test("rejects if-then-else expression with additional keys", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              "if-then-else": {
                // @ts-expect-error - if-then-else is strict
                if: { Var: "context" },
                then: { Value: true },
                else: { Value: false },
                extra: { Value: false },
              },
            },
          },
        ],
      })
    );
  });
});
