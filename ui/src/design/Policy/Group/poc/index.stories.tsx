import { type Meta, type StoryObj } from "@storybook/react-vite";
import { PolicyConditionFlow } from "./PolicyConditionFlow";
import nestedBooleanGroups from "./policies/nestedBooleanGroups";
import nestedBooleanGroupsWithTrailingAnd from "./policies/nestedBooleanGroupsWithTrailingAnd";
import simpleWhen from "./policies/simpleWhen";

const meta = {
  title: "POC/Policy/Condition Flow",
  component: PolicyConditionFlow,
} satisfies Meta<typeof PolicyConditionFlow>;

export default meta;

type Story = StoryObj<typeof meta>;
export const SimpleWhen: Story = {
  args: { expression: simpleWhen },
};

export const NestedBooleanGroups: Story = {
  args: { expression: nestedBooleanGroups },
};

export const NestedSecondBranchWithTrailingAnd: Story = {
  args: { expression: nestedBooleanGroupsWithTrailingAnd },
};
