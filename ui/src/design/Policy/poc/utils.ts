import { BooleanOperator } from "~/pages/namespace/Explorer/policy/schema/primitives/conditions/expression/utils";
import type { ExpressionType } from "~/pages/namespace/Explorer/policy/schema/primitives/conditions/expression/types";

type PolicyLeafNode = {
  type: "leaf";
  preview: string;
  title: string;
  // rows is the vertical space this node occupies in the layout grid.
  // A leaf always renders as a single row.
  rows: 1;
};

type PolicyAndNode = {
  type: "and";
  items: PolicyLayoutNode[];
  rows: number;
};

type PolicyOrNode = {
  type: "or";
  branches: PolicyAndNode[];
  childSizes: number[];
  rows: number;
};

export type PolicyLayoutNode = PolicyLeafNode | PolicyAndNode | PolicyOrNode;

type BooleanExpressionPayload = {
  left: ExpressionType;
  right: ExpressionType;
};

const getSingleEntry = (
  expression: ExpressionType
): [string, unknown] | undefined => {
  const entries = Object.entries(expression);

  if (entries.length !== 1) return undefined;

  const [key, value] = entries[0] ?? [];

  if (key === undefined) return undefined;

  return [key, value];
};

const isExpressionInput = (value: unknown): value is ExpressionType =>
  typeof value === "object" && value !== null && !Array.isArray(value);

const isBooleanExpressionPayload = (
  value: unknown
): value is BooleanExpressionPayload => {
  if (!isExpressionInput(value)) return false;

  const recordValue = value as Record<string, unknown>;

  return (
    isExpressionInput(recordValue.left) && isExpressionInput(recordValue.right)
  );
};

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

const createLeaf = (expression: ExpressionType): PolicyLeafNode => ({
  type: "leaf",
  preview: JSON.stringify(expression),
  title: JSON.stringify(expression, null, 2),
  rows: 1,
});

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

export const toAndBranch = (expression: ExpressionType): PolicyAndNode => {
  const node = expressionToNode(expression);

  if (node.type === "and") return node;

  return {
    type: "and",
    items: [node],
    rows: node.rows,
  };
};

export const shouldRenderConnector = (
  left: PolicyLayoutNode,
  right: PolicyLayoutNode
) => left.type !== "or" && right.type !== "or";
