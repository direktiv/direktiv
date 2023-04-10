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
  <div className="flex space-x-3">
    <Textarea placeholder="The placeholder text" />
  </div>
);

export const DisabledTextArea = () => (
  <div className="flex space-x-3">
    <Textarea placeholder="The placeholder text" disabled />
  </div>
);
