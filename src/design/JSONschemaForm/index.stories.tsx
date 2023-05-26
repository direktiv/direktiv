import type { Meta, StoryObj } from "@storybook/react";
import { JSONschemaForm } from "../JSONschemaForm";

const meta = {
  title: "Components/JSONschemaForm",
  component: JSONschemaForm,
} satisfies Meta<typeof JSONschemaForm>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: (args) => <JSONschemaForm {...args} />,
  tags: ["autodocs"],
};
