import { AndGroup, Connector, OrGroup } from "~/design/Policy/Group";
import type { ExpressionPath, PolicyLayoutNode } from "./types";

import { Condition } from "~/design/Policy/Condition";
import type { ExpressionType } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/types";
import type { PatternElement } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/like";
import { Placeholder } from "~/design/Policy/Placeholder";
import { Separator } from "~/design/Separator";
import { containsOrGroup } from "./utils";
import { formatExpression } from "./formatExpression";

const getRootVar = (expression: ExpressionType): string | undefined => {
  if ("Var" in expression) return expression.Var;
  if ("." in expression) return getRootVar(expression["."].left);
  return undefined;
};

type ConditionParts = {
  label: string;
  operator: string;
  value: string;
};

const formatLikeValue = (pattern: PatternElement[]) =>
  pattern
    .map((element) => (element === "Wildcard" ? "*" : element.Literal))
    .join("");

const splitConditionExpression = (
  expression: ExpressionType
): ConditionParts | undefined => {
  if ("has" in expression) {
    const rootVar = getRootVar(expression.has.left);
    return {
      label: rootVar ?? formatExpression(expression.has.left),
      operator: "has",
      value: expression.has.attr,
    };
  }

  if ("in" in expression) {
    const rootVar = getRootVar(expression.in.left);
    return {
      label: rootVar ?? formatExpression(expression.in.left),
      operator: "in",
      value: formatExpression(expression.in.right),
    };
  }

  if ("==" in expression) {
    const rootVar = getRootVar(expression["=="].left);
    return {
      label: rootVar ?? formatExpression(expression["=="].left),
      operator: "==",
      value: formatExpression(expression["=="].right),
    };
  }

  if ("like" in expression) {
    const rootVar = getRootVar(expression.like.left);
    return {
      label: rootVar ?? formatExpression(expression.like.left),
      operator: "like",
      value: formatLikeValue(expression.like.pattern),
    };
  }

  if ("getTag" in expression) {
    const rootVar = getRootVar(expression.getTag.left);
    const right = expression.getTag.right;

    const value =
      "Value" in right && typeof right.Value === "string"
        ? right.Value
        : formatExpression(right);

    return {
      label: rootVar ?? formatExpression(expression.getTag.left),
      operator: "getTag",
      value,
    };
  }

  return undefined;
};

type PlaceholderAction = "add-condition" | "add-or-group";

type PolicyNodeProps = {
  node: PolicyLayoutNode;
  onConditionClick?: (path: ExpressionPath) => void;
  onPlaceholderAction?: (
    path: ExpressionPath,
    operator: "&&" | "||",
    action: PlaceholderAction
  ) => void;
};

const PolicyNode = ({
  node,
  onConditionClick,
  onPlaceholderAction,
}: PolicyNodeProps) => {
  // Condition nodes are the terminal expressions in the policy
  // tree, they render directly as a condition component.
  if (node.type === "condition") {
    const preview = formatExpression(node.expression);
    const title = JSON.stringify(node.expression, null, 2);
    const parts = splitConditionExpression(node.expression);

    return (
      <Condition
        className="font-mono"
        title={title}
        onClick={() => onConditionClick?.(node.path)}
      >
        {parts ? (
          <div className="flex size-full flex-col overflow-hidden">
            {parts.label}
            <Separator className="shrink-0" />
            {parts.operator}
            <Separator className="shrink-0" />
            {parts.value}
          </div>
        ) : (
          <span className="block w-full">{preview}</span>
        )}
      </Condition>
    );
  }

  // OR nodes render as vertical branches. The trailing
  // placeholder creates an extra empty branch
  if (node.type === "or") {
    return (
      <OrGroup childSizes={[...node.childSizes, 1]}>
        {node.branches.map((branch, index) => (
          <PolicyNode
            key={index}
            node={branch}
            onConditionClick={onConditionClick}
            onPlaceholderAction={onPlaceholderAction}
          />
        ))}
        <AndGroup>
          <Placeholder
            addCondition={() =>
              onPlaceholderAction?.(node.path, "||", "add-condition")
            }
            addOrGroup={() =>
              onPlaceholderAction?.(node.path, "||", "add-or-group")
            }
          />
        </AndGroup>
      </OrGroup>
    );
  }

  // AND nodes render children in a linear sequence. Each item expands to either
  // [node] or [node, connector], and flatMap combines those into one React
  // children array, e.g. [A, Connector, B, Connector, C]. A trailing connector
  // and placeholder may be appended afterward.
  return (
    <AndGroup>
      {node.items.flatMap((item, index) => {
        const nextItem = node.items[index + 1];
        const renderedItems = [
          <PolicyNode
            key={index}
            node={item}
            onConditionClick={onConditionClick}
            onPlaceholderAction={onPlaceholderAction}
          />,
        ];

        // Connect neighboring AND siblings unless either side is an
        // OR group. OR groups render their own branching structure.
        if (nextItem !== undefined && !containsOrGroup([item, nextItem])) {
          renderedItems.push(<Connector key={index} />);
        }

        return renderedItems;
      })}

      {node.items.length > 0 && node.items.at(-1)?.type !== "or" && (
        <Connector />
      )}

      <Placeholder
        addCondition={() =>
          onPlaceholderAction?.(node.path, "&&", "add-condition")
        }
        addOrGroup={() =>
          onPlaceholderAction?.(node.path, "&&", "add-or-group")
        }
      />
    </AndGroup>
  );
};

export default PolicyNode;
