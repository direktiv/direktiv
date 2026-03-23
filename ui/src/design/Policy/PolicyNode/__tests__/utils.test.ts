import {
  containsOrGroup,
  expressionToLayoutNode,
  flattenOperator,
  toAndBranch,
} from "../utils";
import { describe, expect, test } from "vitest";
import { ExpressionType } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/types";

describe("flattenOperator", () => {
  test("flattenOperator flattens chained AND expressions in order", () => {
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

  test("flattenOperator flattens chained OR expressions in order", () => {
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

  test("expressionToLayoutNode flattens chained AND expressions", () => {
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
    if (node.type !== "and") {
      throw new Error("Expected an and node");
    }

    expect(node.items).toHaveLength(3);
    expect(node.items).toEqual([
      {
        type: "leaf",
        expression: { Value: true },
        rows: 1,
      },
      {
        type: "leaf",
        expression: { Var: "principal" },
        rows: 1,
      },
      {
        type: "leaf",
        expression: { Value: false },
        rows: 1,
      },
    ]);
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
    if (node.type !== "or") {
      throw new Error("Expected an or node");
    }

    expect(node.branches).toHaveLength(2);
    expect(node.childSizes).toEqual([1, 1]);
    expect(node.rows).toBe(3);
  });

  test("expressionToLayoutNode flattens chained OR expressions into branches", () => {
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
    if (node.type !== "or") {
      throw new Error("Expected an or node");
    }

    expect(node.branches).toHaveLength(3);
    expect(node.branches).toEqual([
      {
        type: "and",
        items: [
          {
            type: "leaf",
            expression: { Value: true },
            rows: 1,
          },
        ],
        rows: 1,
      },
      {
        type: "and",
        items: [
          {
            type: "leaf",
            expression: { Value: false },
            rows: 1,
          },
        ],
        rows: 1,
      },
      {
        type: "and",
        items: [
          {
            type: "leaf",
            expression: { Var: "resource" },
            rows: 1,
          },
        ],
        rows: 1,
      },
    ]);
    expect(node.childSizes).toEqual([1, 1, 1]);
    expect(node.rows).toBe(4);
  });

  test("expressionToLayoutNode includes the nested OR placeholder when calculating rows", () => {
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
    if (node.type !== "and") {
      throw new Error("Expected an 'and' node");
    }

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
    if (node.type !== "and") {
      throw new Error("Expected an and node");
    }

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

  test("expressionToLayoutNode falls back to a leaf for malformed or payloads", () => {
    const expression = {
      "||": { right: { Value: false } },
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
    if (node.type !== "or") {
      throw new Error("Expected an 'or' node");
    }

    expect(node.branches).toEqual([
      {
        type: "and",
        items: [
          {
            type: "or",
            branches: [
              {
                type: "and",
                items: [
                  {
                    type: "leaf",
                    expression: { Value: true },
                    rows: 1,
                  },
                ],
                rows: 1,
              },
              {
                type: "and",
                items: [
                  {
                    type: "leaf",
                    expression: { Value: false },
                    rows: 1,
                  },
                ],
                rows: 1,
              },
            ],
            childSizes: [1, 1],
            rows: 3,
          },
          {
            type: "leaf",
            expression: { Var: "principal" },
            rows: 1,
          },
        ],
        rows: 3,
      },
      {
        type: "and",
        items: [
          {
            type: "leaf",
            expression: { Var: "resource" },
            rows: 1,
          },
        ],
        rows: 1,
      },
    ]);
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
    if (node.type !== "and") {
      throw new Error("Expected an and node");
    }

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

describe("containsOrGroup", () => {
  test("containsOrGroup detects whether any item is an OR group", () => {
    const leafBranch = toAndBranch({ Value: true });
    const leafNode = leafBranch.items[0];
    const orNode = expressionToLayoutNode({
      "||": {
        left: { Value: true },
        right: { Value: false },
      },
    });

    expect(leafNode).toBeDefined();
    if (leafNode === undefined) {
      throw new Error("Expected a leaf node");
    }

    expect(containsOrGroup([leafNode, leafNode])).toBe(false);
    expect(containsOrGroup([orNode, leafNode])).toBe(true);
    expect(containsOrGroup([leafNode, orNode])).toBe(true);
    expect(containsOrGroup([orNode, orNode])).toBe(true);
  });
});
