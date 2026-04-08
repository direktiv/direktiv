import type { Meta, StoryObj } from "@storybook/react-vite";
import { Placeholder } from ".";
import { action } from "storybook/actions";

const meta = {
  title: "Components/Policy/Placeholder",
  component: Placeholder,
} satisfies Meta<typeof Placeholder>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: (args) => (
    <div className="flex p-10">
      <Placeholder {...args} />
    </div>
  ),
  args: {
    addCondition: action("addCondition clicked"),
    addOrGroup: action("addOrGroup clicked"),
  },
  argTypes: {
    addCondition: { action: "addCondition clicked" },
    addOrGroup: { action: "addOrGroup clicked" },
  },
};
