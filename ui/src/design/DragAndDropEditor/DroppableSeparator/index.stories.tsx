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
    blockPath: [0],
    id: "1",
    position: "before",
  },
  argTypes: {
    id: {
      description:
        "The ID is needed for the drag and drop action, it is converted from the blockPath",
      control: "text",
      type: { name: "string", required: true },
    },
  },
};
