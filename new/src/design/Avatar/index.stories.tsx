import type { Meta, StoryObj } from "@storybook/react";
import Avatar from "./index";

const meta = {
  title: "Components/Avatar",
  component: Avatar,
} satisfies Meta<typeof Avatar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => <Avatar {...args} />,
  argTypes: {},
};

export const AvatarChild = () => (
  <div>
    <Avatar>Ad</Avatar>
  </div>
);
