import { describe, expect, test } from "vitest";
import { expressionToNode, shouldRenderConnector, toAndBranch } from "../utils";

import type { ExpressionType } from "~/pages/namespace/Explorer/policy/schema/primitives/conditions/expression/types";

describe("expressionToNode", () => {
  test("expressionToNode flattens chained and expressions", () => {
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

    const node = expressionToNode(expression);

    expect(node.type).toBe("and");
    if (node.type !== "and") return;

    expect(node.items).toHaveLength(3);
    expect(node.rows).toBe(1);
  });

  test("expressionToNode creates or branches with child sizes", () => {
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

    const node = expressionToNode(expression);

    expect(node.type).toBe("or");
    if (node.type !== "or") return;

    expect(node.branches).toHaveLength(2);
    expect(node.childSizes).toEqual([1, 1]);
    expect(node.rows).toBe(2);
  });
});

describe("toAndBranch", () => {
  test("toAndBranch wraps leaf expressions", () => {
    const expression: ExpressionType = { Value: true };

    const branch = toAndBranch(expression);

    expect(branch.type).toBe("and");
    expect(branch.items).toHaveLength(1);
    expect(branch.rows).toBe(1);
  });
});

describe("shouldRenderConnector", () => {
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
