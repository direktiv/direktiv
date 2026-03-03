import { JsonExprBinaryOperators, JsonExprUnaryOperators } from "../constants";
import {
  createBasePolicy,
  expectInvalidPolicy,
  expectValidPolicy,
} from "../../../testUtils";
import { describe, test } from "vitest";

describe("Cedar policy conditions and JsonExpr", () => {
  test("accepts condition with Var JsonExpr", () => {
    // permit(principal, action, resource) when { context };
    expectValidPolicy(
      createBasePolicy({
        conditions: [{ kind: "when", body: { Var: "context" } }],
      })
    );
  });

  test("rejects invalid Var JsonExpr value", () => {
    // permit(principal, action, resource) when { actor };
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [{ kind: "when", body: { Var: "actor" } }],
      })
    );
  });

  test("accepts condition with Slot JsonExpr", () => {
    // permit(principal, action, resource) when { ?principal };
    expectValidPolicy(
      createBasePolicy({
        conditions: [{ kind: "when", body: { Slot: "?principal" } }],
      })
    );
  });

  test("rejects invalid Slot JsonExpr value", () => {
    // permit(principal, action, resource) when { ?action };
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [{ kind: "when", body: { Slot: "?action" } }],
      })
    );
  });

  test("accepts condition with Unknown JsonExpr", () => {
    // permit(principal, action, resource) when { Unknown("x") };
    expectValidPolicy(
      createBasePolicy({
        conditions: [{ kind: "when", body: { Unknown: { name: "x" } } }],
      })
    );
  });

  test("rejects Unknown JsonExpr with multiple keys", () => {
    // permit(principal, action, resource) when { Unknown("x", "y") };
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [{ kind: "when", body: { Unknown: { a: "x", b: "y" } } }],
      })
    );
  });

  test("accepts condition with unary ! JsonExpr", () => {
    // permit(principal, action, resource) when { !context };
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          { kind: "when", body: { "!": { arg: { Var: "context" } } } },
        ],
      })
    );
  });

  test("accepts condition with unary neg JsonExpr", () => {
    // permit(principal, action, resource) when { -1 };
    expectValidPolicy(
      createBasePolicy({
        conditions: [{ kind: "when", body: { neg: { arg: { Value: 1 } } } }],
      })
    );
  });

  test("rejects unary JsonExpr without arg", () => {
    // permit(principal, action, resource) when { ! };
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [{ kind: "when", body: { "!": {} } }],
      })
    );
  });

  test("accepts condition with binary == JsonExpr", () => {
    // permit(principal, action, resource) when { context == "1.3" };
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              "==": { left: { Var: "context" }, right: { Value: "1.3" } },
            },
          },
        ],
      })
    );
  });

  test("rejects binary JsonExpr without right operand", () => {
    // permit(principal, action, resource) when { context == };
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          { kind: "when", body: { "==": { left: { Var: "context" } } } },
        ],
      })
    );
  });

  test("accepts condition with attribute . JsonExpr", () => {
    // permit(principal, action, resource) when { context.tls_version };
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: { ".": { left: { Var: "context" }, attr: "tls_version" } },
          },
        ],
      })
    );
  });

  test("accepts condition with attribute has JsonExpr", () => {
    // permit(principal, action, resource) when { principal has "email" };
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

  test("rejects attribute JsonExpr without attr", () => {
    // permit(principal, action, resource) when { context. };
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          { kind: "when", body: { ".": { left: { Var: "context" } } } },
        ],
      })
    );
  });

  test("accepts condition with is JsonExpr", () => {
    // permit(principal, action, resource) when { principal is User in Group::"friends" };
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              is: {
                left: { Var: "principal" },
                entity_type: "User",
                in: { Value: { __entity: { type: "Group", id: "friends" } } },
              },
            },
          },
        ],
      })
    );
  });

  test("rejects is JsonExpr without entity_type", () => {
    // permit(principal, action, resource) when { principal is in Group::"friends" };
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              is: {
                left: { Var: "principal" },
                in: { Value: { __entity: { type: "Group", id: "friends" } } },
              },
            },
          },
        ],
      })
    );
  });

  test("accepts condition with like JsonExpr", () => {
    // permit(principal, action, resource) when { resource.email like "*@amazon.com" };
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              like: {
                left: { ".": { left: { Var: "resource" }, attr: "email" } },
                pattern: ["Wildcard", { Literal: "@amazon.com" }],
              },
            },
          },
        ],
      })
    );
  });

  test("rejects like JsonExpr with invalid pattern element", () => {
    // permit(principal, action, resource) when { resource.email like [123] };
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: { like: { left: { Var: "resource" }, pattern: [123] } },
          },
        ],
      })
    );
  });

  test("accepts condition with if-then-else JsonExpr", () => {
    // permit(principal, action, resource) when { if context then true else false };
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

  test("rejects if-then-else JsonExpr without else", () => {
    // permit(principal, action, resource) when { if context then true };
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              "if-then-else": { if: { Var: "context" }, then: { Value: true } },
            },
          },
        ],
      })
    );
  });

  test("accepts condition with Set JsonExpr", () => {
    // permit(principal, action, resource) when { [1, 2, "something"] };
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: { Set: [{ Value: 1 }, { Value: 2 }, { Value: "something" }] },
          },
        ],
      })
    );
  });

  test("rejects Set JsonExpr with invalid element", () => {
    // permit(principal, action, resource) when { [1, ???] };
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          { kind: "when", body: { Set: [{ Value: 1 }, { nope: true }] } },
        ],
      })
    );
  });

  test("accepts condition with Record JsonExpr", () => {
    // permit(principal, action, resource) when { { foo: "spam", somethingelse: false } };
    expectValidPolicy(
      createBasePolicy({
        conditions: [
          {
            kind: "when",
            body: {
              Record: {
                foo: { Value: "spam" },
                somethingelse: { Value: false },
              },
            },
          },
        ],
      })
    );
  });

  test("rejects Record JsonExpr with invalid field expr", () => {
    // permit(principal, action, resource) when { { foo: ??? } };
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [
          { kind: "when", body: { Record: { foo: { nope: true } } } },
        ],
      })
    );
  });

  test("accepts condition with extension function JsonExpr", () => {
    // permit(principal, action, resource) when { decimal("10.0") };
    expectValidPolicy(
      createBasePolicy({
        conditions: [{ kind: "when", body: { decimal: [{ Value: "10.0" }] } }],
      })
    );
  });

  test("rejects extension JsonExpr with non-array args", () => {
    // permit(principal, action, resource) when { decimal("10.0") } represented with invalid JSON shape;
    expectInvalidPolicy(
      createBasePolicy({
        conditions: [{ kind: "when", body: { decimal: { Value: "10.0" } } }],
      })
    );
  });

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
});
