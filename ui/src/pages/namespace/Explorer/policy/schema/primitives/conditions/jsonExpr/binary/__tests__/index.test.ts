import { JsonExprBinaryOperators } from "../../utils";
import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../../../utils/testutils";
import { describe, test } from "vitest";

describe("Binary JsonExpr schema", () => {
  for (const operator of JsonExprBinaryOperators) {
    test(`accepts binary operator ${operator}`, () => {
      expectValidPolicy(
        createBasePolicy({
          conditions: [
            {
              kind: "when",
              body: {
                [operator]: { left: { Value: 1 }, right: { Value: 2 } },
              },
            },
          ],
        })
      );
    });
  }

  test("rejects binary expression without right operand", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          { kind: "when", body: { "==": { left: { Var: "context" } } } },
        ],
      })
    );
  });

  test("rejects binary expression with additional keys", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              "==": {
                left: { Var: "context" },
                right: { Value: "1.3" },
                extra: { Value: true },
              },
            },
          },
        ],
      })
    );
  });
});
