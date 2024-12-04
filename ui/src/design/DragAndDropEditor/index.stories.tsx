import type { Meta, StoryObj } from "@storybook/react";

import { DragAndDropEditor } from "./index";

const meta = {
  title: "Components/DragAndDropEditor",
  component: DragAndDropEditor,
} satisfies Meta<typeof DragAndDropEditor>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => <DragAndDropEditor />,
};
