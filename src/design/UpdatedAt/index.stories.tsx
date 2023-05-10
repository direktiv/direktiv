import type { Meta, StoryObj } from "@storybook/react";
import UpdatedAt from "./index";

const meta = {
  title: "Components/UpdateAt",
  component: UpdatedAt,
} satisfies Meta<typeof UpdatedAt>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => <UpdatedAt {...args} />,
  argTypes: {},
};

export const UpdatedNow = () => (
  <div>
    <UpdatedAt date={new Date().toLocaleString()} />
  </div>
);
