import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../../../utils/testutils";
import { describe, test } from "vitest";
import { ExpressionBinaryOperators } from "../../utils";

describe("Binary Expression schema", () => {
  type BinaryOperator = (typeof ExpressionBinaryOperators)[number];
  type BinaryOperands = {
    left: Record<string, unknown>;
    right: Record<string, unknown>;
  };

  const binaryOperandsByOperator = {
    "==": {
      left: { Var: "principal" },
      right: { Value: { __entity: { type: "User", id: "alice" } } },
    },
    "!=": {
      left: { Var: "principal" },
      right: { Value: { __entity: { type: "User", id: "bob" } } },
    },
    in: {
      left: { Var: "resource" },
      right: { Value: { __entity: { type: "Folder", id: "Public" } } },
    },
    "<": { left: { Value: 1 }, right: { Value: 2 } },
    "<=": { left: { Value: 1 }, right: { Value: 2 } },
    ">": { left: { Value: 2 }, right: { Value: 1 } },
    ">=": { left: { Value: 2 }, right: { Value: 1 } },
    "&&": { left: { Value: true }, right: { Value: true } },
    "||": { left: { Value: true }, right: { Value: false } },
    "+": { left: { Value: 1 }, right: { Value: 2 } },
    "-": { left: { Value: 3 }, right: { Value: 1 } },
    "*": { left: { Value: 2 }, right: { Value: 3 } },
    contains: {
      left: { Set: [{ Value: 1 }, { Value: 2 }, { Value: 3 }] },
      right: { Value: 2 },
    },
    containsAll: {
      left: { Set: [{ Value: 1 }, { Value: 2 }, { Value: 3 }] },
      right: { Set: [{ Value: 1 }, { Value: 2 }] },
    },
    containsAny: {
      left: { Set: [{ Value: 1 }, { Value: 2 }, { Value: 3 }] },
      right: { Set: [{ Value: 2 }, { Value: 4 }] },
    },
    hasTag: { left: { Var: "resource" }, right: { Value: "classification" } },
    getTag: { left: { Var: "resource" }, right: { Value: "classification" } },
  } satisfies Record<BinaryOperator, BinaryOperands>;

  for (const operator of ExpressionBinaryOperators) {
    test(`accepts binary operator ${operator}`, () => {
      /*
        Cedar examples by operator:
        when { principal == User::"alice" };
        when { principal != User::"bob" };
        when { resource in Folder::"Public" };
        when { 1 < 2 };
        when { 1 <= 2 };
        when { 2 > 1 };
        when { 2 >= 1 };
        when { true && true };
        when { true || false };
        when { 1 + 2 == 3 };
        when { 3 - 1 == 2 };
        when { 2 * 3 == 6 };
        when { [1, 2, 3] contains 2 };
        when { [1, 2, 3] containsAll [1, 2] };
        when { [1, 2, 3] containsAny [2, 4] };
        when { resource hasTag "classification" };
        when { resource getTag "classification" == "internal" };
      */
      expectValidPolicy(
        createBasePolicy({
          conditions: [
            {
              kind: "when",
              body: {
                [operator]: {
                  left: binaryOperandsByOperator[operator].left,
                  right: binaryOperandsByOperator[operator].right,
                },
              },
            },
          ],
        })
      );
    });
  }

  test("rejects binary expression without right operand", () => {
    // Cedar (invalid): when { context == };
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
