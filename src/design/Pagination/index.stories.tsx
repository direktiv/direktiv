import type { Meta, StoryObj } from "@storybook/react";
import Pagination from "./index";

const meta = {
  title: "Components/Pagination",
  component: Pagination,
} satisfies Meta<typeof Pagination>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => <Pagination {...args} />,
  argTypes: {
    total: {
      description: "total pages",
      control: {
        type: "text",
      },
      type: { name: "number", required: false },
    },
  },
};
