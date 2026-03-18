import type {
  AndExpression,
  ExpressionType,
  OrExpression,
} from "../schema/primitives/conditions/expression/types";
import type { PolicyAndNode, PolicyLayoutNode, PolicyLeafNode } from "./types";
import { BinaryExpressionSchema } from "../schema/primitives/conditions/expression/binary";
import type { BooleanOperator } from "../schema/primitives/conditions/expression/utils";
import { ExpressionSchema } from "../schema/primitives/conditions/expression";

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
  if (operator === "&&") {
    if (!isAndExpression(expression)) return [expression];

    return [
      ...flattenOperator(expression["&&"].left, operator),
      ...flattenOperator(expression["&&"].right, operator),
    ];
  }

  if (!isOrExpression(expression)) return [expression];

  return [
    ...flattenOperator(expression["||"].left, operator),
    ...flattenOperator(expression["||"].right, operator),
  ];
};

// Wraps a raw expression in a leaf layout node.
const createLeaf = (expression: ExpressionType): PolicyLeafNode => ({
  type: "leaf",
  expression,
  rows: 1,
});

// Converts a policy expression into the render-oriented layout tree.
export const expressionToLayoutNode = (
  expression: ExpressionType
): PolicyLayoutNode => {
  if (isAndExpression(expression)) {
    // Flatten chained AND expressions so the layout renders one horizontal group.
    const items = flattenOperator(expression, "&&").map(expressionToLayoutNode);

    return {
      type: "and",
      items,
      // Side-by-side items only need the tallest child height.
      rows: Math.max(1, ...items.map((item) => item.rows)),
    };
  }

  if (isOrExpression(expression)) {
    // Flatten chained OR expressions into vertically stacked branches.
    const branches = flattenOperator(expression, "||").map(toAndBranch);
    const childSizes = branches.map((branch) => branch.rows);

    return {
      type: "or",
      branches,
      childSizes,
      // Add branch heights plus one placeholder row rendered after the branches.
      rows: childSizes.reduce((sum, size) => sum + size, 1),
    };
  }

  // Everything else is a leaf condition in the layout tree.
  return createLeaf(expression);
};

// Ensures an expression can be rendered as an AND branch, wrapping non-AND nodes when needed.
export const toAndBranch = (expression: ExpressionType): PolicyAndNode => {
  const node = expressionToLayoutNode(expression);

  if (node.type === "and") return node;

  return {
    type: "and",
    items: [node],
    rows: node.rows,
  };
};

// AND connectors are skipped whenever either neighboring node is an OR group.
export const shouldRenderConnector = (
  left: PolicyLayoutNode,
  right: PolicyLayoutNode
) => left.type !== "or" && right.type !== "or";
