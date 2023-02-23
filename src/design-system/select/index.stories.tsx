import type { Meta, StoryObj } from "@storybook/react";

import Select from "./index";

const meta = {
  title: "Design System/Select",
  component: Select,
} satisfies Meta<typeof Select>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => {
    return (
      <Select {...args}>
        <option value="1">Option 1</option>
        <option value="1">Option 2</option>
        <option value="1">Option 3</option>
      </Select>
    );
  },
  args: {},
  argTypes: {
    size: {
      description: "Select size",
      control: "select",
      options: ["xs", "md", "lg"],
      type: { name: "string", required: false },
    },
    border: {
      description: "with border",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
    ghost: {
      description: "almost transparent background",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
  },
};
