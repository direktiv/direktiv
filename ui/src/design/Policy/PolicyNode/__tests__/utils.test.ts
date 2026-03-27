import {
  addDefaultConditionToGroup,
  addDefaultOrGroupToAnd,
  addDefaultOrGroupToOr,
  appendToBooleanGroup,
  buildBooleanChain,
  containsOrGroup,
  expressionToLayoutNode,
  flattenOperator,
  replaceExpressionAtPath,
  toAndBranch,
} from "../utils";
import { describe, expect, test } from "vitest";
import { ExpressionType } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/types";
import nestedBooleanGroups from "../policies/nestedBooleanGroups";
import nestedBooleanGroupsWithTrailingAnd from "../policies/nestedBooleanGroupsWithTrailingAnd";

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
  test("expressionToLayoutNode creates condition nodes for non-boolean expressions", () => {
    const expression: ExpressionType = { Var: "principal" };

    const node = expressionToLayoutNode(expression);

    expect(node).toEqual({
      type: "condition",
      expression,
      rows: 1,
      path: [],
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
        type: "condition",
        expression: { Value: true },
        rows: 1,
        path: [
          { operator: "&&", side: "left" },
          { operator: "&&", side: "left" },
        ],
      },
      {
        type: "condition",
        expression: { Var: "principal" },
        rows: 1,
        path: [
          { operator: "&&", side: "left" },
          { operator: "&&", side: "right" },
        ],
      },
      {
        type: "condition",
        expression: { Value: false },
        rows: 1,
        path: [{ operator: "&&", side: "right" }],
      },
    ]);
    expect(node.rows).toBe(1);
    expect(node.path).toEqual([]);
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
    expect(node.path).toEqual([]);
    expect(node.branches[0]?.path).toEqual([{ operator: "||", side: "left" }]);
    expect(node.branches[1]?.path).toEqual([{ operator: "||", side: "right" }]);
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
            type: "condition",
            expression: { Value: true },
            rows: 1,
            path: [
              { operator: "||", side: "left" },
              { operator: "||", side: "left" },
            ],
          },
        ],
        rows: 1,
        path: [
          { operator: "||", side: "left" },
          { operator: "||", side: "left" },
        ],
      },
      {
        type: "and",
        items: [
          {
            type: "condition",
            expression: { Value: false },
            rows: 1,
            path: [
              { operator: "||", side: "left" },
              { operator: "||", side: "right" },
            ],
          },
        ],
        rows: 1,
        path: [
          { operator: "||", side: "left" },
          { operator: "||", side: "right" },
        ],
      },
      {
        type: "and",
        items: [
          {
            type: "condition",
            expression: { Var: "resource" },
            rows: 1,
            path: [{ operator: "||", side: "right" }],
          },
        ],
        rows: 1,
        path: [{ operator: "||", side: "right" }],
      },
    ]);
    expect(node.childSizes).toEqual([1, 1, 1]);
    expect(node.rows).toBe(4);
    expect(node.path).toEqual([]);
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
    expect(node.items[0]?.type).toBe("condition");
    expect(node.items[1]?.type).toBe("or");
  });

  test("expressionToLayoutNode falls back to a condition for malformed boolean payloads", () => {
    const expression = {
      "&&": { left: { Value: true } },
    } as ExpressionType;

    expect(expressionToLayoutNode(expression)).toEqual({
      type: "condition",
      expression,
      rows: 1,
      path: [],
    });
  });

  test("expressionToLayoutNode falls back to a condition for malformed or payloads", () => {
    const expression = {
      "||": { right: { Value: false } },
    } as ExpressionType;

    expect(expressionToLayoutNode(expression)).toEqual({
      type: "condition",
      expression,
      rows: 1,
      path: [],
    });
  });

  test("expressionToLayoutNode falls back to a condition for multi-key objects", () => {
    const expression = {
      Value: true,
      Var: "principal",
    } as ExpressionType;

    expect(expressionToLayoutNode(expression)).toEqual({
      type: "condition",
      expression,
      rows: 1,
      path: [],
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
                    type: "condition",
                    expression: { Value: true },
                    rows: 1,
                    path: [
                      { operator: "||", side: "left" },
                      { operator: "&&", side: "left" },
                      { operator: "||", side: "left" },
                    ],
                  },
                ],
                rows: 1,
                path: [
                  { operator: "||", side: "left" },
                  { operator: "&&", side: "left" },
                  { operator: "||", side: "left" },
                ],
              },
              {
                type: "and",
                items: [
                  {
                    type: "condition",
                    expression: { Value: false },
                    rows: 1,
                    path: [
                      { operator: "||", side: "left" },
                      { operator: "&&", side: "left" },
                      { operator: "||", side: "right" },
                    ],
                  },
                ],
                rows: 1,
                path: [
                  { operator: "||", side: "left" },
                  { operator: "&&", side: "left" },
                  { operator: "||", side: "right" },
                ],
              },
            ],
            childSizes: [1, 1],
            rows: 3,
            path: [
              { operator: "||", side: "left" },
              { operator: "&&", side: "left" },
            ],
          },
          {
            type: "condition",
            expression: { Var: "principal" },
            rows: 1,
            path: [
              { operator: "||", side: "left" },
              { operator: "&&", side: "right" },
            ],
          },
        ],
        rows: 3,
        path: [{ operator: "||", side: "left" }],
      },
      {
        type: "and",
        items: [
          {
            type: "condition",
            expression: { Var: "resource" },
            rows: 1,
            path: [{ operator: "||", side: "right" }],
          },
        ],
        rows: 1,
        path: [{ operator: "||", side: "right" }],
      },
    ]);
    expect(node.childSizes).toEqual([3, 1]);
    expect(node.rows).toBe(5);
    expect(node.path).toEqual([]);
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

  test("toAndBranch wraps condition expressions", () => {
    const expression: ExpressionType = { Value: true };

    const branch = toAndBranch(expression);

    expect(branch.type).toBe("and");
    expect(branch.items).toHaveLength(1);
    expect(branch.rows).toBe(1);
    expect(branch.path).toEqual([]);
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
    expect(branch.path).toEqual([]);
  });
});

describe("containsOrGroup", () => {
  test("containsOrGroup detects whether any item is an OR group", () => {
    const conditionBranch = toAndBranch({ Value: true });
    const conditionNode = conditionBranch.items[0];
    const orNode = expressionToLayoutNode({
      "||": {
        left: { Value: true },
        right: { Value: false },
      },
    });

    expect(conditionNode).toBeDefined();
    if (conditionNode === undefined) {
      throw new Error("Expected a condition node");
    }

    expect(containsOrGroup([conditionNode, conditionNode])).toBe(false);
    expect(containsOrGroup([orNode, conditionNode])).toBe(true);
    expect(containsOrGroup([conditionNode, orNode])).toBe(true);
    expect(containsOrGroup([orNode, orNode])).toBe(true);
  });
});

describe("replaceExpressionAtPath", () => {
  test("replaceExpressionAtPath replaces the root expression when path is empty", () => {
    const expression: ExpressionType = { Value: true };

    expect(replaceExpressionAtPath(expression, [], { Value: false })).toEqual({
      Value: false,
    });
  });

  test("replaceExpressionAtPath replaces a nested leaf expression", () => {
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

    expect(
      replaceExpressionAtPath(
        expression,
        [
          { operator: "&&", side: "left" },
          { operator: "&&", side: "right" },
        ],
        { Value: false }
      )
    ).toEqual({
      "&&": {
        left: {
          "&&": {
            left: { Value: true },
            right: { Value: false },
          },
        },
        right: { Value: false },
      },
    });
  });

  test("replaceExpressionAtPath returns the original expression for an invalid path", () => {
    const expression: ExpressionType = {
      "&&": {
        left: { Value: true },
        right: { Var: "principal" },
      },
    };

    expect(
      replaceExpressionAtPath(
        expression,
        [
          { operator: "&&", side: "left" },
          { operator: "||", side: "right" },
        ],
        { Value: false }
      )
    ).toEqual(expression);
  });
});

describe("buildBooleanChain", () => {
  test("buildBooleanChain rebuilds an AND chain from flat items", () => {
    expect(
      buildBooleanChain("&&", [
        { Value: true },
        { Var: "principal" },
        { Value: false },
      ])
    ).toEqual({
      "&&": {
        left: {
          "&&": {
            left: { Value: true },
            right: { Var: "principal" },
          },
        },
        right: { Value: false },
      },
    });
  });

  test("buildBooleanChain rebuilds an OR chain from flat items", () => {
    expect(
      buildBooleanChain("||", [
        { Value: true },
        { Var: "principal" },
        { Value: false },
      ])
    ).toEqual({
      "||": {
        left: {
          "||": {
            left: { Value: true },
            right: { Var: "principal" },
          },
        },
        right: { Value: false },
      },
    });
  });

  test("buildBooleanChain falls back to the default condition for empty items", () => {
    expect(buildBooleanChain("&&", [])).toEqual({ Value: true });
  });
});

describe("appendToBooleanGroup", () => {
  test("appendToBooleanGroup appends a condition to the root AND group", () => {
    const expression: ExpressionType = {
      "&&": {
        left: { Value: true },
        right: { Var: "principal" },
      },
    };

    expect(
      appendToBooleanGroup(expression, [], "&&", { Value: false })
    ).toEqual({
      "&&": {
        left: {
          "&&": {
            left: { Value: true },
            right: { Var: "principal" },
          },
        },
        right: { Value: false },
      },
    });
  });

  test("appendToBooleanGroup appends an OR group to a nested AND group", () => {
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

    expect(
      appendToBooleanGroup(
        expression,
        [{ operator: "&&", side: "left" }],
        "&&",
        {
          "||": {
            left: { Value: true },
            right: { Value: true },
          },
        }
      )
    ).toEqual({
      "&&": {
        left: {
          "&&": {
            left: {
              "&&": {
                left: { Value: true },
                right: { Var: "principal" },
              },
            },
            right: {
              "||": {
                left: { Value: true },
                right: { Value: true },
              },
            },
          },
        },
        right: { Value: false },
      },
    });
  });

  test("appendToBooleanGroup appends a condition branch to the root OR group", () => {
    const expression: ExpressionType = {
      "||": {
        left: { Value: true },
        right: { Var: "principal" },
      },
    };

    expect(
      appendToBooleanGroup(expression, [], "||", { Value: false })
    ).toEqual({
      "||": {
        left: {
          "||": {
            left: { Value: true },
            right: { Var: "principal" },
          },
        },
        right: { Value: false },
      },
    });
  });

  test("appendToBooleanGroup appends a nested OR branch to a nested OR group", () => {
    const expression: ExpressionType = {
      "&&": {
        left: {
          "||": {
            left: { Value: true },
            right: { Var: "principal" },
          },
        },
        right: { Value: false },
      },
    };

    expect(
      appendToBooleanGroup(
        expression,
        [{ operator: "&&", side: "left" }],
        "||",
        {
          "||": {
            left: { Value: true },
            right: { Value: true },
          },
        }
      )
    ).toEqual({
      "&&": {
        left: {
          "||": {
            left: {
              "||": {
                left: { Value: true },
                right: { Var: "principal" },
              },
            },
            right: {
              "||": {
                left: { Value: true },
                right: { Value: true },
              },
            },
          },
        },
        right: { Value: false },
      },
    });
  });

  test("appendToBooleanGroup returns the original expression for an invalid path", () => {
    const expression: ExpressionType = {
      "&&": {
        left: { Value: true },
        right: { Var: "principal" },
      },
    };

    expect(
      appendToBooleanGroup(
        expression,
        [
          { operator: "&&", side: "left" },
          { operator: "||", side: "right" },
        ],
        "&&",
        { Value: false }
      )
    ).toEqual(expression);
  });
});

describe("group demo helpers", () => {
  test("addDefaultConditionToGroup appends the default condition to an AND group", () => {
    const expression: ExpressionType = {
      "&&": {
        left: { Value: true },
        right: { Var: "principal" },
      },
    };

    expect(addDefaultConditionToGroup(expression, [], "&&")).toEqual({
      "&&": {
        left: {
          "&&": {
            left: { Value: true },
            right: { Var: "principal" },
          },
        },
        right: { Value: true },
      },
    });
  });

  test("addDefaultOrGroupToAnd appends a default OR group to an AND group", () => {
    const expression: ExpressionType = {
      "&&": {
        left: { Value: true },
        right: { Var: "principal" },
      },
    };

    expect(addDefaultOrGroupToAnd(expression, [])).toEqual({
      "&&": {
        left: {
          "&&": {
            left: { Value: true },
            right: { Var: "principal" },
          },
        },
        right: {
          "||": {
            left: { Value: true },
            right: { Value: true },
          },
        },
      },
    });
  });

  test("addDefaultOrGroupToOr appends a default OR group to an OR group", () => {
    const expression: ExpressionType = {
      "||": {
        left: { Value: true },
        right: { Var: "principal" },
      },
    };

    expect(addDefaultOrGroupToOr(expression, [])).toEqual({
      "||": {
        left: {
          "||": {
            left: { Value: true },
            right: { Var: "principal" },
          },
        },
        right: {
          "||": {
            left: { Value: true },
            right: { Value: true },
          },
        },
      },
    });
  });
});

describe("realistic fixture interactions", () => {
  test("nested boolean groups fixture stays renderable after OR branch growth", () => {
    const nextExpression = addDefaultConditionToGroup(
      nestedBooleanGroups,
      [{ operator: "&&", side: "right" }],
      "||"
    );

    const nextNode = toAndBranch(nextExpression);

    expect(nextNode.type).toBe("and");
    expect(nextNode.rows).toBeGreaterThan(1);
    expect(nextNode.items[1]?.type).toBe("or");

    const orNode = nextNode.items[1];
    expect(orNode?.type).toBe("or");
    if (orNode?.type !== "or") {
      throw new Error("Expected an OR node");
    }

    expect(orNode.branches).toHaveLength(3);
  });

  test("trailing-and fixture stays renderable after nested OR insertion", () => {
    const nextExpression = addDefaultOrGroupToOr(
      nestedBooleanGroupsWithTrailingAnd,
      [
        { operator: "&&", side: "right" },
        { operator: "&&", side: "left" },
        { operator: "||", side: "right" },
        { operator: "&&", side: "left" },
      ]
    );

    const nextNode = toAndBranch(nextExpression);

    expect(nextNode.type).toBe("and");
    expect(nextNode.rows).toBeGreaterThan(3);
    expect(nextNode.items).toHaveLength(3);
  });
});
