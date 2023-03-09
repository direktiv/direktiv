import type { Meta, StoryObj } from "@storybook/react";
import Avatar from "./index";

const meta = {
  title: "Components (next)/Avatar",
  component: Avatar,
} satisfies Meta<typeof Avatar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => {
    return <Avatar {...args} />;
  },
  argTypes: {
    size: {
      description: "Avatar Size",
      control: "select",
      options: ["xs", "sm", "lg", "xlg"],
      type: { name: "string", required: false },
    },
    placeholder: {
      description: "Placeholder",
      control: {
        type: "text",
      },
      type: { name: "string", required: false },
    },
    src: {
      description: "Source url",
      control: {
        type: "text",
        default:
          "https://daisyui.com/images/stock/photo-1534528741775-53994a69daeb.jpg",
      },
      type: { name: "string", required: false },
    },
  },
};

export const AvatarSizes = () => {
  return (
    <div>
      <Avatar size="xs"></Avatar>
      <Avatar size="sm"></Avatar>
      <Avatar size="lg"></Avatar>
      <Avatar size="xlg"></Avatar>
    </div>
  );
};

export const AvatarPlaceholder = () => {
  return (
    <div>
      <Avatar size="xlg" placeholder="CD"></Avatar>
    </div>
  );
};
