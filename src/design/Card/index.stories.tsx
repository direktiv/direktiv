import type { Meta, StoryObj } from "@storybook/react";

import { Card } from "./index";

const meta = {
  title: "Components/Card",
  component: Card,
} satisfies Meta<typeof Card>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => <Card {...args} className="w-100 h-64"></Card>,
  argTypes: {
    withBackground: {
      description: "Card has default gray background",
      type: { name: "boolean", required: false },
    },
  },
};
export const CardWithBackground = () => (
  <Card withBackground className="w-100 h-64"></Card>
);
