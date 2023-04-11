import type { Meta, StoryObj } from "@storybook/react";
import { Textarea } from "./index";

const meta = {
  title: "Components/Textarea",
  component: Textarea,
} satisfies Meta<typeof Textarea>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: (args) => <Textarea {...args} />,
  tags: ["autodocs"],
};

export const WithPlaceholder = () => (
  <Textarea placeholder="The placeholder text" />
);

export const DisabledTextArea = () => (
  <Textarea placeholder="The placeholder text" disabled />
);
