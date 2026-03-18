import type {
  BooleanExpressionPayload,
  ExpressionType,
} from "~/pages/namespace/Explorer/policy/schema/primitives/conditions/expression/types";
import type { PolicyAndNode, PolicyLayoutNode, PolicyLeafNode } from "./types";
import { BooleanOperator } from "~/pages/namespace/Explorer/policy/schema/primitives/conditions/expression/utils";

// Returns the single key/value entry for an expression object when it has exactly one field.
const getSingleEntry = (
  expression: ExpressionType
): [string, unknown] | undefined => {
  const entries = Object.entries(expression);

  if (entries.length !== 1) return undefined;

  const [key, value] = entries[0] ?? [];

  if (key === undefined) return undefined;

  return [key, value];
};

// Narrows unknown values to expression-shaped objects.
const isExpressionInput = (value: unknown): value is ExpressionType =>
  typeof value === "object" && value !== null && !Array.isArray(value);

// Checks whether a value looks like a boolean expression payload with left/right expressions.
const isBooleanExpressionPayload = (
  value: unknown
): value is BooleanExpressionPayload<ExpressionType> => {
  if (!isExpressionInput(value)) return false;

  const recordValue = value as Record<string, unknown>;

  return (
    isExpressionInput(recordValue.left) && isExpressionInput(recordValue.right)
  );
};

// Recursively flattens chained uses of the same boolean operator into a linear list.
const flattenOperator = (
  expression: ExpressionType,
  operator: BooleanOperator
): ExpressionType[] => {
  const entry = getSingleEntry(expression);

  if (entry?.[0] !== operator || !isBooleanExpressionPayload(entry[1])) {
    return [expression];
  }

  return [
    ...flattenOperator(entry[1].left, operator),
    ...flattenOperator(entry[1].right, operator),
  ];
};

// Wraps a raw expression in a leaf layout node.
const createLeaf = (expression: ExpressionType): PolicyLeafNode => ({
  type: "leaf",
  expression,
  rows: 1,
});

// Converts a policy expression into the render-oriented layout tree.
export const expressionToNode = (
  expression: ExpressionType
): PolicyLayoutNode => {
  const entry = getSingleEntry(expression);

  if (entry?.[0] === "&&") {
    const items = flattenOperator(expression, "&&").map(expressionToNode);

    return {
      type: "and",
      items,
      rows: Math.max(1, ...items.map((item) => item.rows)),
    };
  }

  if (entry?.[0] === "||") {
    const branches = flattenOperator(expression, "||").map(toAndBranch);
    const childSizes = branches.map((branch) => branch.rows);

    return {
      type: "or",
      branches,
      childSizes,
      rows: childSizes.reduce((sum, size) => sum + size, 1),
    };
  }

  return createLeaf(expression);
};

// Ensures an expression can be rendered as an AND branch, wrapping non-AND nodes when needed.
export const toAndBranch = (expression: ExpressionType): PolicyAndNode => {
  const node = expressionToNode(expression);

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
