import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../../../utils/testutils";
import { describe, test } from "vitest";
import { ExpressionUnaryOperators } from "../../utils";

describe("Unary Expression schema", () => {
  type UnaryOperator = (typeof ExpressionUnaryOperators)[number];
  type UnaryArg = Record<string, unknown>;

  const unaryArgsByOperator = {
    "!": { Value: true },
    neg: { Value: 1 },
    isEmpty: { Set: [] },
  } satisfies Record<UnaryOperator, UnaryArg>;

  for (const operator of ExpressionUnaryOperators) {
    test(`accepts unary operator ${operator}`, () => {
      /*
        Cedar examples by operator:
        when { !true };
        when { -1 };
        when { isEmpty([]) };
      */
      expectValidPolicy(
        createBasePolicy({
          conditions: [
            {
              kind: "when",
              body: { [operator]: { arg: unaryArgsByOperator[operator] } },
            },
          ],
        })
      );
    });
  }

  test("rejects unary expression without arg", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [{ kind: "when", body: { "!": {} } }],
      })
    );
  });

  test("rejects unary expression with additional keys", () => {
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: { "!": { arg: { Var: "context" }, other: true } },
          },
        ],
      })
    );
  });
});
