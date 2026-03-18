import { type Meta, type StoryObj } from "@storybook/react-vite";
import RenderNode from "./RenderNode";
import nestedBooleanGroups from "./policies/nestedBooleanGroups";
import nestedBooleanGroupsWithTrailingAnd from "./policies/nestedBooleanGroupsWithTrailingAnd";
import simpleWhen from "./policies/simpleWhen";
import { toAndBranch } from "./utils";

const meta = {
  title: "Components/Policy/RenderNode",
  component: RenderNode,
} satisfies Meta<typeof RenderNode>;

export default meta;

type Story = StoryObj<typeof meta>;
export const Default: Story = {
  args: { node: toAndBranch(simpleWhen) },
};

export const NestedBooleanGroups: Story = {
  render: (args) => (
    <div className="overflow-x-auto">
      <RenderNode {...args} />
    </div>
  ),
  args: { node: toAndBranch(nestedBooleanGroups) },
};

export const NestedSecondBranchWithTrailingAnd: Story = {
  render: (args) => (
    <div className="overflow-x-auto">
      <RenderNode {...args} />
    </div>
  ),
  args: { node: toAndBranch(nestedBooleanGroupsWithTrailingAnd) },
};
