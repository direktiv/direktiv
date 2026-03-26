import { ExpressionType } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/types";

export type ExpressionPathSegment = {
  operator: "&&" | "||";
  side: "left" | "right";
};

export type ExpressionPath = ExpressionPathSegment[];

export type PolicyConditionNode = {
  type: "condition";
  rows: 1; // single-row condition
  expression: ExpressionType;
  path: ExpressionPath;
};

export type PolicyAndNode = {
  type: "and";
  rows: number; // tallest child height
  items: PolicyLayoutNode[];
  path: ExpressionPath;
};

type PolicyOrNode = {
  type: "or";
  rows: number; // total stacked branch height
  branches: PolicyAndNode[];
  // childSizes tracks the row height of each OR branch so the renderer can
  // stack branches vertically with the correct amount of space.
  childSizes: number[];
  path: ExpressionPath;
};

export type PolicyLayoutNode =
  | PolicyConditionNode
  | PolicyAndNode
  | PolicyOrNode;
