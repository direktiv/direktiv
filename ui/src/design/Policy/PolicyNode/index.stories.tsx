import { type Meta, type StoryObj } from "@storybook/react-vite";
import {
  addDefaultConditionToGroup,
  addDefaultOrGroupToAnd,
  addDefaultOrGroupToOr,
  getExpressionAtPath,
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
  const currentExpression = getExpressionAtPath(expression, path);

  const nextExpression: ExpressionType =
    "Value" in currentExpression && currentExpression.Value === true
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
              return addDefaultConditionToGroup(
                previousExpression,
                path,
                operator
              );
            }

            if (operator === "&&") {
              return addDefaultOrGroupToAnd(previousExpression, path);
            }

            return addDefaultOrGroupToOr(previousExpression, path);
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
