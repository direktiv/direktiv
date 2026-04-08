import type { Meta, StoryObj } from "@storybook/react-vite";
import { Condition } from ".";

const meta = {
  title: "Components/Policy/Condition",
  component: Condition,
} satisfies Meta<typeof Condition>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <div className="flex p-10">
      <Condition>
        <div>label</div>
        <div>equal</div>
        <div>value</div>
      </Condition>
    </div>
  ),
};
