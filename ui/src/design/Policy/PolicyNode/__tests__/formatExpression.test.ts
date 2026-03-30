import { describe, expect, test } from "vitest";

import type { ExpressionType } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/types";
import { formatExpression } from "../formatExpression";

describe("formatExpression", () => {
  test("formats has expressions as readable Cedar text", () => {
    const expression: ExpressionType = {
      has: { left: { Var: "principal" }, attr: "email" },
    };

    expect(formatExpression(expression)).toBe("principal has email");
  });

  test("formats like expressions with wildcard patterns", () => {
    const expression: ExpressionType = {
      like: {
        left: { ".": { left: { Var: "principal" }, attr: "email" } },
        pattern: ["Wildcard", { Literal: "@example.com" }],
      },
    };

    expect(formatExpression(expression)).toBe(
      "principal.email like *@example.com"
    );
  });

  test("formats getTag expressions with unquoted string values", () => {
    const expression: ExpressionType = {
      getTag: {
        left: { Var: "resource" },
        right: { Value: "classification" },
      },
    };

    expect(formatExpression(expression)).toBe("resource getTag classification");
  });

  test("formats binary expressions", () => {
    const expression: ExpressionType = {
      "==": {
        left: { ".": { left: { Var: "context" }, attr: "region" } },
        right: { Value: "eu-west-1" },
      },
    };

    expect(formatExpression(expression)).toBe('context.region == "eu-west-1"');
  });

  test("formats record-like object values in binary expressions", () => {
    const expression: ExpressionType = {
      "==": {
        left: { Var: "context" },
        right: { Value: { region: "eu-west-1" } },
      },
    };

    expect(formatExpression(expression)).toBe("context == region: eu-west-1");
  });

  test("formats entity values in binary in expressions", () => {
    const expression: ExpressionType = {
      in: {
        left: { Var: "resource" },
        right: {
          Value: { __entity: { type: "Folder", id: "Public" } },
        },
      },
    };

    expect(formatExpression(expression)).toBe(
      "resource in type: folder id: public"
    );
  });

  test("formats set and record expressions", () => {
    const setExpression: ExpressionType = {
      Set: [{ Value: 1 }, { Value: "read" }],
    };
    const recordExpression: ExpressionType = {
      Record: {
        region: { Value: "eu-west-1" },
        secure: { Value: true },
        folder: { Value: { __entity: { type: "Folder", id: "Public" } } },
      },
    };

    expect(formatExpression(setExpression)).toBe('[1, "read"]');
    expect(formatExpression(recordExpression)).toBe(
      '{"region": "eu-west-1", "secure": true, "folder": type: folder id: public}'
    );
  });

  test("formats unary expressions", () => {
    const expression: ExpressionType = {
      isEmpty: { arg: { ".": { left: { Var: "context" }, attr: "tags" } } },
    };

    expect(formatExpression(expression)).toBe("isEmpty(context.tags)");
  });

  test("falls back to JSON for malformed expressions", () => {
    const malformedExpression = {
      unexpected: true,
    } as ExpressionType;

    expect(formatExpression(malformedExpression)).toBe(
      JSON.stringify(malformedExpression)
    );
  });
});
