import { describe, expect, test } from "vitest";
import {
  expressionToLayoutNode,
  shouldRenderConnector,
  toAndBranch,
} from "../utils";

import type { ExpressionType } from "~/pages/namespace/Explorer/policy/schema/primitives/conditions/expression/types";

describe("expressionToLayoutNode", () => {
  test("expressionToLayoutNode creates leaf nodes for non-boolean expressions", () => {
    const expression: ExpressionType = { Var: "principal" };

    const node = expressionToLayoutNode(expression);

    expect(node).toEqual({
      type: "leaf",
      expression,
      rows: 1,
    });
  });

  test("expressionToLayoutNode flattens chained and expressions", () => {
    const expression: ExpressionType = {
      "&&": {
        left: {
          "&&": {
            left: { Value: true },
            right: { Var: "principal" },
          },
        },
        right: { Value: false },
      },
    };

    const node = expressionToLayoutNode(expression);

    expect(node.type).toBe("and");
    if (node.type !== "and") return;

    expect(node.items).toHaveLength(3);
    expect(node.rows).toBe(1);
  });

  test("expressionToLayoutNode creates or branches with child sizes", () => {
    const expression: ExpressionType = {
      "||": {
        left: { Value: true },
        right: {
          "&&": {
            left: { Value: false },
            right: { Var: "resource" },
          },
        },
      },
    };

    const node = expressionToLayoutNode(expression);

    expect(node.type).toBe("or");
    if (node.type !== "or") return;

    expect(node.branches).toHaveLength(2);
    expect(node.childSizes).toEqual([1, 1]);
    expect(node.rows).toBe(3);
  });

  test("expressionToLayoutNode flattens chained or expressions into branches", () => {
    const expression: ExpressionType = {
      "||": {
        left: {
          "||": {
            left: { Value: true },
            right: { Value: false },
          },
        },
        right: { Var: "resource" },
      },
    };

    const node = expressionToLayoutNode(expression);

    expect(node.type).toBe("or");
    if (node.type !== "or") return;

    expect(node.branches).toHaveLength(3);
    expect(node.childSizes).toEqual([1, 1, 1]);
    expect(node.rows).toBe(4);
  });

  test("expressionToLayoutNode includes placeholder row in nested or height", () => {
    const expression: ExpressionType = {
      "&&": {
        left: {
          "||": {
            left: { Value: true },
            right: { Value: false },
          },
        },
        right: { Var: "resource" },
      },
    };

    const node = expressionToLayoutNode(expression);

    expect(node.type).toBe("and");
    if (node.type !== "and") return;

    expect(node.rows).toBe(3);
  });
});

describe("toAndBranch", () => {
  test("toAndBranch returns existing and nodes unchanged", () => {
    const expression: ExpressionType = {
      "&&": {
        left: { Value: true },
        right: { Var: "principal" },
      },
    };

    const node = expressionToLayoutNode(expression);

    expect(node.type).toBe("and");
    if (node.type !== "and") return;

    expect(toAndBranch(expression)).toStrictEqual(node);
  });

  test("toAndBranch wraps leaf expressions", () => {
    const expression: ExpressionType = { Value: true };

    const branch = toAndBranch(expression);

    expect(branch.type).toBe("and");
    expect(branch.items).toHaveLength(1);
    expect(branch.rows).toBe(1);
  });
});

describe("shouldRenderConnector", () => {
  test("shouldRenderConnector skips connectors around or groups", () => {
    const leafBranch = toAndBranch({ Value: true });
    const leafNode = leafBranch.items[0];
    const orNode = expressionToLayoutNode({
      "||": {
        left: { Value: true },
        right: { Value: false },
      },
    });

    expect(leafNode).toBeDefined();
    if (leafNode === undefined) return;

    expect(shouldRenderConnector(leafNode, leafNode)).toBe(true);
    expect(shouldRenderConnector(orNode, leafNode)).toBe(false);
    expect(shouldRenderConnector(leafNode, orNode)).toBe(false);
  });
});
