import {
  AndExpression,
  ExpressionType,
  OrExpression,
} from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/types";
import type {
  ExpressionPath,
  PolicyAndNode,
  PolicyConditionNode,
  PolicyLayoutNode,
} from "./types";
import { BinaryExpressionSchema } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/binary";
import { BooleanOperator } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/utils";
import { ExpressionSchema } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression";

const defaultConditionExpression: ExpressionType = { Value: true };
const defaultOrGroupExpression: ExpressionType = {
  "||": {
    left: { Value: true },
    right: { Value: true },
  },
};

const isAndExpression = (expression: unknown): expression is AndExpression => {
  const parsed = BinaryExpressionSchema(ExpressionSchema).safeParse(expression);

  return parsed.success && "&&" in parsed.data;
};

const isOrExpression = (expression: unknown): expression is OrExpression => {
  const parsed = BinaryExpressionSchema(ExpressionSchema).safeParse(expression);

  return parsed.success && "||" in parsed.data;
};

// Converts nested binary boolean trees like a && (b && c) or
// (a || b) || c into a flat list the layout layer can render as a
// single AND row or OR branch stack. It stops when the operator changes
// so mixed expressions like a && (b || c) keep their nested structure.
export const flattenOperator = (
  expression: ExpressionType,
  operator: BooleanOperator
): ExpressionType[] => {
  const children = getBinaryChildren(expression, operator);

  if (!children) return [expression];

  return [
    ...flattenOperator(children.left, operator),
    ...flattenOperator(children.right, operator),
  ];
};

type FlattenedExpressionWithPath = {
  expression: ExpressionType;
  path: ExpressionPath;
};

// Returns the child pair for the requested boolean operator, or null for leaf/malformed nodes.
const getBinaryChildren = (
  expression: ExpressionType,
  operator: BooleanOperator
) => {
  if (operator === "&&") {
    if (!isAndExpression(expression)) return null;

    return expression["&&"];
  }

  if (!isOrExpression(expression)) return null;

  return expression["||"];
};

// Walks the expression tree by layout path and returns the deepest matching node.
export const getExpressionAtPath = (
  expression: ExpressionType,
  path: ExpressionPath
): ExpressionType =>
  path.reduce<ExpressionType>((currentExpression, segment) => {
    const children = getBinaryChildren(currentExpression, segment.operator);

    if (!children) {
      return currentExpression;
    }

    return children[segment.side];
  }, expression);

const flattenOperatorWithPaths = (
  expression: ExpressionType,
  operator: BooleanOperator,
  basePath: ExpressionPath = []
): FlattenedExpressionWithPath[] => {
  const children = getBinaryChildren(expression, operator);

  if (!children) {
    return [{ expression, path: basePath }];
  }

  return [
    ...flattenOperatorWithPaths(children.left, operator, [
      ...basePath,
      { operator, side: "left" },
    ]),
    ...flattenOperatorWithPaths(children.right, operator, [
      ...basePath,
      { operator, side: "right" },
    ]),
  ];
};

// Converts a policy expression into the render-oriented layout tree.
export const expressionToLayoutNode = (
  expression: ExpressionType,
  path: ExpressionPath = []
): PolicyLayoutNode => {
  if (isAndExpression(expression)) {
    // Flatten chained AND expressions so the layout renders one horizontal group.
    const items = flattenOperatorWithPaths(expression, "&&", path).map(
      ({ expression: item, path: itemPath }) =>
        expressionToLayoutNode(item, itemPath)
    );

    return {
      type: "and",
      items,
      // Side-by-side items only need the tallest child height.
      rows: Math.max(1, ...items.map((item) => item.rows)),
      path,
    };
  }

  if (isOrExpression(expression)) {
    // Flatten chained OR expressions into vertically stacked branches.
    const branches = flattenOperatorWithPaths(expression, "||", path).map(
      ({ expression: branch, path: branchPath }) =>
        toAndBranch(branch, branchPath)
    );
    const childSizes = branches.map((branch) => branch.rows);

    return {
      type: "or",
      branches,
      childSizes,
      // Add branch heights plus one placeholder row rendered after the branches.
      rows: childSizes.reduce((sum, size) => sum + size, 1),
      path,
    };
  }

  // Everything else is a terminal condition.
  return {
    type: "condition",
    expression,
    rows: 1,
    path,
  } satisfies PolicyConditionNode;
};

// Ensures an expression can be rendered as an AND branch, wrapping non-AND nodes when needed.
export const toAndBranch = (
  expression: ExpressionType,
  path: ExpressionPath = []
): PolicyAndNode => {
  const node = expressionToLayoutNode(expression, path);

  if (node.type === "and") return node;

  return {
    type: "and",
    items: [node],
    rows: node.rows,
    path,
  };
};

export const replaceExpressionAtPath = (
  expression: ExpressionType,
  path: ExpressionPath,
  nextExpression: ExpressionType
): ExpressionType => {
  if (path.length === 0) {
    return nextExpression;
  }

  const [segment, ...rest] = path;

  if (segment === undefined) {
    return expression;
  }

  const children = getBinaryChildren(expression, segment.operator);

  if (!children) {
    return expression;
  }

  return {
    [segment.operator]: {
      ...children,
      [segment.side]: replaceExpressionAtPath(
        children[segment.side],
        rest,
        nextExpression
      ),
    },
  };
};

export const buildBooleanChain = (
  operator: BooleanOperator,
  items: ExpressionType[]
): ExpressionType => {
  const [firstItem, ...restItems] = items;

  if (firstItem === undefined) {
    return defaultConditionExpression;
  }

  return restItems.reduce<ExpressionType>((currentExpression, item) => {
    if (operator === "&&") {
      return {
        "&&": {
          left: currentExpression,
          right: item,
        },
      };
    }

    return {
      "||": {
        left: currentExpression,
        right: item,
      },
    };
  }, firstItem);
};

export const appendToBooleanGroup = (
  expression: ExpressionType,
  path: ExpressionPath,
  operator: BooleanOperator,
  nextItem: ExpressionType
): ExpressionType => {
  const currentGroup = getExpressionAtPath(expression, path);

  const nextGroup = buildBooleanChain(operator, [
    ...flattenOperator(currentGroup, operator),
    nextItem,
  ]);

  return replaceExpressionAtPath(expression, path, nextGroup);
};

export const addDefaultConditionToGroup = (
  expression: ExpressionType,
  path: ExpressionPath,
  operator: BooleanOperator
) =>
  appendToBooleanGroup(expression, path, operator, defaultConditionExpression);

export const addDefaultOrGroupToAnd = (
  expression: ExpressionType,
  path: ExpressionPath
) => appendToBooleanGroup(expression, path, "&&", defaultOrGroupExpression);

export const addDefaultOrGroupToOr = (
  expression: ExpressionType,
  path: ExpressionPath
) => appendToBooleanGroup(expression, path, "||", defaultOrGroupExpression);

// Returns true when any item in the list is an OR group.
export const containsOrGroup = (items: PolicyLayoutNode[]) =>
  items.some((item) => item.type === "or");
