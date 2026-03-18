import type { ExpressionType } from "../schema/primitives/conditions/expression/types";

export type PolicyLeafNode = {
  type: "leaf";
  rows: 1; // single-row leaf
  expression: ExpressionType;
};

export type PolicyAndNode = {
  type: "and";
  rows: number; // tallest child height
  items: PolicyLayoutNode[];
};

type PolicyOrNode = {
  type: "or";
  rows: number; // total stacked branch height
  branches: PolicyAndNode[];
  // childSizes tracks the row height of each OR branch so the renderer can
  // stack branches vertically with the correct amount of space.
  childSizes: number[];
};

export type PolicyLayoutNode = PolicyLeafNode | PolicyAndNode | PolicyOrNode;
