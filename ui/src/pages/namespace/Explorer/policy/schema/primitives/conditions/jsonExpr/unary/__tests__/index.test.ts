import { JsonExprUnaryOperators } from "../../utils";
import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../../../utils/testutils";
import { describe, test } from "vitest";

describe("Unary JsonExpr schema", () => {
  for (const operator of JsonExprUnaryOperators) {
    test(`accepts unary operator ${operator}`, () => {
      expectValidPolicy(
        createBasePolicy({
          conditions: [
            {
              kind: "when",
              body: { [operator]: { arg: { Value: true } } },
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
