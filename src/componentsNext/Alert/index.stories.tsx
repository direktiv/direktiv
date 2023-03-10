import type { Meta, StoryObj } from "@storybook/react";
import Alert from "./index";

const meta = {
  title: "Components (next)/Alert",
  component: Alert,
} satisfies Meta<typeof Alert>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => {
    return <Alert {...args} />;
  },
  argTypes: {
    variant: {
      description: "Variant of the alert",
      options: ["info", "success", "warning", "error", undefined],
      control: { type: "select" },
      type: "string",
    },
    text: {
      type: "string",
      defaultValue: "Hey this is alert",
    },
  },
};
