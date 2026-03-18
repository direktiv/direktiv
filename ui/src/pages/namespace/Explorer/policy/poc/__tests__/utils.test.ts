import { describe, expect, test } from "vitest";
import {
  expressionToLayoutNode,
  flattenOperator,
  shouldRenderConnector,
  toAndBranch,
} from "../utils";

import type { ExpressionType } from "~/pages/namespace/Explorer/policy/schema/primitives/conditions/expression/types";

describe("flattenOperator", () => {
  test("flattenOperator flattens chained and expressions in order", () => {
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

    expect(flattenOperator(expression, "&&")).toEqual([
      { Value: true },
      { Var: "principal" },
      { Value: false },
    ]);
  });

  test("flattenOperator flattens chained or expressions in order", () => {
    const expression: ExpressionType = {
      "||": {
        left: { Value: true },
        right: {
          "||": {
            left: { Value: false },
            right: { Var: "resource" },
          },
        },
      },
    };

    expect(flattenOperator(expression, "||")).toEqual([
      { Value: true },
      { Value: false },
      { Var: "resource" },
    ]);
  });

  test("flattenOperator keeps mixed boolean operators nested", () => {
    const nestedOr: ExpressionType = {
      "||": {
        left: { Value: false },
        right: { Var: "resource" },
      },
    };
    const expression: ExpressionType = {
      "&&": {
        left: { Value: true },
        right: nestedOr,
      },
    };

    expect(flattenOperator(expression, "&&")).toEqual([
      { Value: true },
      nestedOr,
    ]);
  });

  test("flattenOperator returns malformed boolean payloads unchanged", () => {
    const expression = {
      "&&": { left: { Value: true } },
    } as ExpressionType;

    expect(flattenOperator(expression, "&&")).toEqual([expression]);
  });
});

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

  test("expressionToLayoutNode keeps mixed boolean operators nested", () => {
    const expression: ExpressionType = {
      "&&": {
        left: { Value: true },
        right: {
          "||": {
            left: { Value: false },
            right: { Var: "resource" },
          },
        },
      },
    };

    const node = expressionToLayoutNode(expression);

    expect(node.type).toBe("and");
    if (node.type !== "and") return;

    expect(node.items).toHaveLength(2);
    expect(node.items[0]?.type).toBe("leaf");
    expect(node.items[1]?.type).toBe("or");
  });

  test("expressionToLayoutNode falls back to a leaf for malformed boolean payloads", () => {
    const expression = {
      "&&": { left: { Value: true } },
    } as ExpressionType;

    expect(expressionToLayoutNode(expression)).toEqual({
      type: "leaf",
      expression,
      rows: 1,
    });
  });

  test("expressionToLayoutNode falls back to a leaf for multi-key objects", () => {
    const expression = {
      Value: true,
      Var: "principal",
    } as ExpressionType;

    expect(expressionToLayoutNode(expression)).toEqual({
      type: "leaf",
      expression,
      rows: 1,
    });
  });

  test("expressionToLayoutNode preserves taller OR branch sizes", () => {
    const expression: ExpressionType = {
      "||": {
        left: {
          "&&": {
            left: {
              "||": {
                left: { Value: true },
                right: { Value: false },
              },
            },
            right: { Var: "principal" },
          },
        },
        right: { Var: "resource" },
      },
    };

    const node = expressionToLayoutNode(expression);

    expect(node.type).toBe("or");
    if (node.type !== "or") return;

    expect(node.childSizes).toEqual([3, 1]);
    expect(node.rows).toBe(5);
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

  test("toAndBranch wraps or expressions as a single branch item", () => {
    const expression: ExpressionType = {
      "||": {
        left: { Value: true },
        right: { Value: false },
      },
    };

    const branch = toAndBranch(expression);

    expect(branch.items).toHaveLength(1);
    expect(branch.items[0]?.type).toBe("or");
    expect(branch.rows).toBe(3);
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
    expect(shouldRenderConnector(orNode, orNode)).toBe(false);
  });
});
