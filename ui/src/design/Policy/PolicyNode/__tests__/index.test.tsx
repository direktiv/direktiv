import { describe, expect, test } from "vitest";
import { render, screen } from "@testing-library/react";

import type { ExpressionType } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/types";
import type { PolicyLayoutNode } from "../types";
import PolicyNode from "..";

const renderCondition = (expression: ExpressionType) => {
  const node: PolicyLayoutNode = {
    type: "condition",
    rows: 1,
    expression,
    path: [],
  };

  render(<PolicyNode node={node} />);
};

describe("PolicyNode condition splitting", () => {
  test("splits like expressions into label operator and value", () => {
    renderCondition({
      like: {
        left: { ".": { left: { Var: "principal" }, attr: "email" } },
        pattern: ["Wildcard", { Literal: "@example.com" }],
      },
    });

    const condition = screen.getByLabelText("condition");
    expect(condition.textContent).toContain("principal");
    expect(condition.textContent).toContain("like");
    expect(condition.textContent).toContain("*@example.com");
  });

  test("splits getTag expressions into label operator and value", () => {
    renderCondition({
      getTag: {
        left: { Var: "resource" },
        right: { Value: "classification" },
      },
    });

    const condition = screen.getByLabelText("condition");
    expect(condition.textContent).toContain("resource");
    expect(condition.textContent).toContain("getTag");
    expect(condition.textContent).toContain("classification");
  });

  test("falls back to preview for unsupported expressions", () => {
    renderCondition({
      isEmpty: { arg: { ".": { left: { Var: "context" }, attr: "tags" } } },
    });

    const condition = screen.getByLabelText("condition");
    expect(condition.textContent).toContain("isEmpty(context.tags)");
  });
});
