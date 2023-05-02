import type { Meta, StoryObj } from "@storybook/react";

import { Card } from "./index";

const meta = {
  title: "Components/Card",
  component: Card,
} satisfies Meta<typeof Card>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => <Card {...args} className="h-64"></Card>,
  argTypes: {
    withBackground: {
      description: "Card has default gray background",
      type: { name: "boolean", required: false },
    },
    noShadow: {
      description: "Card has default shadow",
      type: { name: "boolean", required: false },
    },
  },
};

export const CardBackgrounds = () => (
  <div className="flex space-x-5">
    <Card className="flex h-64 w-64 items-center justify-center">
      no background
    </Card>
    <Card withBackground className="flex h-64 w-64 items-center justify-center">
      with background
    </Card>
  </div>
);
export const NoShadow = () => (
  <div className="flex space-x-5">
    <Card noShadow className="flex h-64 w-64 items-center justify-center">
      no shadow, no background
    </Card>
    <Card
      noShadow
      withBackground
      className="flex h-64 w-64 items-center justify-center"
    >
      no shadow, with background
    </Card>
  </div>
);
