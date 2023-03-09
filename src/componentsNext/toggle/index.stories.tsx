import type { Meta, StoryObj } from "@storybook/react";
import { Toggle } from "./index";

const meta = {
  title: "Components (next)/Toggle",
  component: Toggle,
} satisfies Meta<typeof Toggle>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => {
    return <Toggle {...args} />;
  },
  argTypes: {},
};
