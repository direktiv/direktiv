import { type Meta, type StoryObj } from "@storybook/react-vite";
import { useState } from "react";
import PolicyNode from ".";
import nestedBooleanGroups from "./policies/nestedBooleanGroups";
import nestedBooleanGroupsWithTrailingAnd from "./policies/nestedBooleanGroupsWithTrailingAnd";
import simpleWhen from "./policies/simpleWhen";
import { toggleDemoConditionAtPath, toAndBranch } from "./utils";
import type { ExpressionType } from "~/pages/namespace/Explorer/Policy/schema/primitives/conditions/expression/types";

const meta = {
  title: "Components/Policy/PolicyNode",
  component: PolicyNode,
} satisfies Meta<typeof PolicyNode>;

export default meta;

type Story = StoryObj<typeof meta>;

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
