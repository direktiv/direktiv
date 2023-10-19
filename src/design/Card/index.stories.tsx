import type { Meta, StoryObj } from "@storybook/react";

import { Card } from "./index";

const meta = {
  title: "Components/Card",
  component: Card,
} satisfies Meta<typeof Card>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => (
    <div className="bg-info-3 p-10 dark:bg-info-dark-3">
      <Card {...args} className="h-64"></Card>
    </div>
  ),
  argTypes: {
    background: {
      description: "Card has default no background",
      control: { type: "radio" },
      type: { name: "string", required: false },
      options: ["none", "weight-1", "weight-2"],
    },
    noShadow: {
      description: "Card has default shadow",
      type: { name: "boolean", required: false },
    },
  },
};

export const CardBackgrounds = () => (
  <div className="flex space-x-5 bg-info-3 p-10 dark:bg-info-dark-3">
    <Card className="flex h-64 w-64 items-center justify-center p-5 text-center">
      no background
    </Card>
    <Card
      background="weight-1"
      className="flex h-64 w-64 items-center justify-center p-5 text-center"
    >
      background weight 1
    </Card>
    <Card
      background="weight-2"
      className="flex h-64 w-64 items-center justify-center p-5 text-center"
    >
      background weight 2
    </Card>
  </div>
);

export const NoShadow = () => (
  <div className="flex space-x-5 bg-info-3 p-10 dark:bg-info-dark-3">
    <Card
      noShadow
      className="flex h-64 w-64 items-center justify-center p-5 text-center"
    >
      no shadow, no background
    </Card>
    <Card
      background="weight-1"
      noShadow
      className="flex h-64 w-64 items-center justify-center p-5 text-center"
    >
      no shadow, background weight 1
    </Card>
    <Card
      background="weight-2"
      noShadow
      className="flex h-64 w-64 items-center justify-center p-5 text-center"
    >
      no shadow, background weight 2
    </Card>
  </div>
);
