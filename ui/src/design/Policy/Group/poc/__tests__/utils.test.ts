import { describe, expect, test } from "vitest";
import {
  expressionToNode,
  shouldRenderConnector,
  stringifyExpression,
  toAndBranch,
} from "../utils";

import type { ConditionsType } from "~/pages/namespace/Explorer/policy/schema/primitives/conditions";

describe("Policy Group POC utils", () => {
  test("stringifyExpression formats objects for display", () => {
    expect(stringifyExpression({ Value: true })).toContain('"Value": true');
  });

  test("expressionToNode flattens chained and expressions", () => {
    const clauses = [
      {
        kind: "when",
        body: {
          "&&": {
            left: {
              "&&": {
                left: { Value: true },
                right: { Var: "principal" },
              },
            },
            right: { Value: false },
          },
        },
      },
    ] satisfies ConditionsType;

    const clause = clauses[0];

    expect(clause).toBeDefined();
    if (clause === undefined) return;

    const node = expressionToNode(clause.body);

    expect(node.type).toBe("and");
    if (node.type !== "and") return;

    expect(node.items).toHaveLength(3);
    expect(node.rows).toBe(1);
  });

  test("expressionToNode creates or branches with child sizes", () => {
    const clauses = [
      {
        kind: "when",
        body: {
          "||": {
            left: { Value: true },
            right: {
              "&&": {
                left: { Value: false },
                right: { Var: "resource" },
              },
            },
          },
        },
      },
    ] satisfies ConditionsType;

    const clause = clauses[0];

    expect(clause).toBeDefined();
    if (clause === undefined) return;

    const node = expressionToNode(clause.body);

    expect(node.type).toBe("or");
    if (node.type !== "or") return;

    expect(node.branches).toHaveLength(2);
    expect(node.childSizes).toEqual([1, 1]);
    expect(node.rows).toBe(2);
  });

  test("toAndBranch wraps leaf expressions", () => {
    const clauses = [
      {
        kind: "when",
        body: { Value: true },
      },
    ] satisfies ConditionsType;

    const clause = clauses[0];

    expect(clause).toBeDefined();
    if (clause === undefined) return;

    const branch = toAndBranch(clause.body);

    expect(branch.type).toBe("and");
    expect(branch.items).toHaveLength(1);
    expect(branch.rows).toBe(1);
  });

  test("shouldRenderConnector skips connectors around or groups", () => {
    const leafBranch = toAndBranch({ Value: true });
    const leafNode = leafBranch.items[0];
    const orNode = expressionToNode({
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
  });
});
