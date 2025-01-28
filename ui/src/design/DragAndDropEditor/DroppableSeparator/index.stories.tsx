import type { Meta, StoryObj } from "@storybook/react";

import { DroppableSeparator } from "./index";

const meta = {
  title: "Components/DragAndDropEditor/DroppableSeparator",
  component: DroppableSeparator,
} satisfies Meta<typeof DroppableSeparator>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => <DroppableSeparator {...args} />,
  args: {
    id: "1",
  },
  argTypes: {
    id: {
      description: "The ID is needed for the drang and drop action",
      control: "text",
      type: { name: "string", required: true },
    },
  },
};
