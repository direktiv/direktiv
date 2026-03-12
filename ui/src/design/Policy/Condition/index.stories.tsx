import type { Meta, StoryObj } from "@storybook/react-vite";
import { Condition } from ".";

const meta = {
  title: "Components/Policy/Condition",
  component: Condition,
} satisfies Meta<typeof Condition>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: (args) => (
    <div className="flex p-10">
      <Condition {...args} />
    </div>
  ),
  args: {
    label: "label",
    value: "value",
    operator: "equal",
  },
  argTypes: {
    label: { control: "text", description: "label" },
    value: { control: "text", description: "value" },
  },
};
