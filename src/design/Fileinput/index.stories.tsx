import type { Meta, StoryObj } from "@storybook/react";

import FileInput from "./index";

const meta = {
  title: "Components/FileInput",
  component: FileInput,
} satisfies Meta<typeof FileInput>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: (args) => <FileInput placeholder="default" {...args} />,
  tags: ["autodocs"],
};
