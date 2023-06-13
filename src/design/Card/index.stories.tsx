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
    weight: {
      description: "Card has default no background",
      type: { name: "number", required: false },
    },
    noShadow: {
      description: "Card has default shadow",
      type: { name: "boolean", required: false },
    },
  },
};

export const CardBackgrounds = () => (
  <div className="flex space-x-5 bg-success-10 p-10 dark:bg-success-dark-10">
    <Card weight={0} className="flex h-64 w-64 items-center justify-center">
      no background
    </Card>
    <Card weight={1} className="flex h-64 w-64 items-center justify-center">
      with white background
    </Card>
    <Card weight={2} className="flex h-64 w-64 items-center justify-center">
      with gray-1 background
    </Card>
  </div>
);

export const NoShadow = () => (
  <div className="flex space-x-5 bg-success-10 p-10 dark:bg-success-dark-10">
    <Card noShadow className="flex h-64 w-64 items-center justify-center">
      no shadow, no background
    </Card>
    <Card
      weight={1}
      noShadow
      className="flex h-64 w-64 items-center justify-center"
    >
      no shadow, with background
    </Card>
  </div>
);
