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
  argTypes: {
    size: {
      description: "Avatar Size",
      control: "select",
      options: ["xs", "sm", "lg", "xlg"],
      type: { name: "string", required: false },
    },
  },
};

export const AvatarSizes = () => (
  <div>
    <Avatar size="xs">AB</Avatar>
    <Avatar size="sm">CD</Avatar>
    <Avatar size="lg">EF</Avatar>
    <Avatar size="xlg">GH</Avatar>
  </div>
);
