import { type Meta, type StoryObj } from "@storybook/react-vite";
import {
  addDemoConditionToGroup,
  addStarterOrGroupToAnd,
  addStarterOrGroupToOr,
  replaceExpressionAtPath,
  toAndBranch,
} from "./utils";
import type { ExpressionPath } from "./types";
import type { ExpressionType } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/types";
import PolicyNode from ".";
import nestedBooleanGroups from "./policies/nestedBooleanGroups";
import nestedBooleanGroupsWithTrailingAnd from "./policies/nestedBooleanGroupsWithTrailingAnd";
import simpleWhen from "./policies/simpleWhen";
import { useState } from "react";

const meta = {
  title: "Components/Policy/PolicyNode",
  component: PolicyNode,
} satisfies Meta<typeof PolicyNode>;

export default meta;

type Story = StoryObj<typeof meta>;

const toggleDemoConditionAtPath = (
  expression: ExpressionType,
  path: ExpressionPath
) => {
  const current = path.reduce<ExpressionType>((currentExpression, segment) => {
    if (segment.operator === "&&" && "&&" in currentExpression) {
      return currentExpression["&&"][segment.side];
    }

    if (segment.operator === "||" && "||" in currentExpression) {
      return currentExpression["||"][segment.side];
    }

    return currentExpression;
  }, expression);

  const nextExpression: ExpressionType =
    "Value" in current && current.Value === true
      ? { Value: false }
      : { Value: true };

  return replaceExpressionAtPath(expression, path, nextExpression);
};

const InteractivePolicyNode = ({
  expression,
}: {
  expression: ExpressionType;
}) => {
  const [currentExpression, setCurrentExpression] = useState(expression);

  return (
    <div className="overflow-x-auto">
      <PolicyNode
        node={toAndBranch(currentExpression)}
        onConditionClick={(path) => {
          setCurrentExpression((previousExpression) =>
            toggleDemoConditionAtPath(previousExpression, path)
          );
        }}
        onPlaceholderAction={(path, operator, action) => {
          setCurrentExpression((previousExpression) => {
            if (action === "add-condition") {
              return addDemoConditionToGroup(
                previousExpression,
                path,
                operator
              );
            }

            if (operator === "&&") {
              return addStarterOrGroupToAnd(previousExpression, path);
            }

            return addStarterOrGroupToOr(previousExpression, path);
          });
        }}
      />
    </div>
  );
};

export const Default: Story = {
  args: { node: toAndBranch(simpleWhen) },
  render: () => <InteractivePolicyNode expression={simpleWhen} />,
};

export const NestedBooleanGroups: Story = {
  args: { node: toAndBranch(nestedBooleanGroups) },
  render: () => <InteractivePolicyNode expression={nestedBooleanGroups} />,
};

export const NestedSecondBranchWithTrailingAnd: Story = {
  args: { node: toAndBranch(nestedBooleanGroupsWithTrailingAnd) },
  render: () => (
    <InteractivePolicyNode expression={nestedBooleanGroupsWithTrailingAnd} />
  ),
};
