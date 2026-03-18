import type { ExpressionType } from "~/pages/namespace/Explorer/policy/schema/primitives/conditions/expression/types";

type LeafVM = {
  type: "leaf";
  preview: string;
  title: string;
  rows: 1;
};

type AndVM = {
  type: "and";
  items: NodeVM[];
  rows: number;
};

type OrVM = {
  type: "or";
  branches: AndVM[];
  childSizes: number[];
  rows: number;
};

export type NodeVM = LeafVM | AndVM | OrVM;

type BooleanOperator = "&&" | "||";
type BooleanExpressionPayload = {
  left: ExpressionType;
  right: ExpressionType;
};

export const formatExpressionTitle = (value: unknown): string =>
  JSON.stringify(value, null, 2) ?? "expression";

export const formatExpressionInline = (value: unknown): string =>
  JSON.stringify(value) ?? "expression";

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

const createLeaf = (expression: ExpressionType): LeafVM => ({
  type: "leaf",
  preview: formatExpressionInline(expression),
  title: formatExpressionTitle(expression),
  rows: 1,
});

export const expressionToNode = (expression: ExpressionType): NodeVM => {
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
      rows: childSizes.reduce((sum, size) => sum + size, 0),
    };
  }

  return createLeaf(expression);
};

export const toAndBranch = (expression: ExpressionType): AndVM => {
  const node = expressionToNode(expression);

  if (node.type === "and") return node;

  return {
    type: "and",
    items: [node],
    rows: node.rows,
  };
};

export const shouldRenderConnector = (left: NodeVM, right: NodeVM) =>
  left.type !== "or" && right.type !== "or";
