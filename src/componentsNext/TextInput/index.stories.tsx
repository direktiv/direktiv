import type { Meta, StoryObj } from "@storybook/react";
import TextInput from "./index";

const meta = {
  title: "Components (next)/TextInput",
  component: TextInput,
} satisfies Meta<typeof TextInput>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: (args) => <TextInput placeholder="default" {...args} />,
  args: {},
  tags: ["autodocs"],
  argTypes: {
    size: {
      description: "select size",
      control: "select",
      options: ["xs", "sm", "md", "lg"],
      type: { name: "string", required: false },
    },
    variant: {
      description: "select variant",
      control: "select",
      options: [
        "primary",
        "secondary",
        "info",
        "ghost",
        "success",
        "warning",
        "error",
        "accent",
      ],
      type: { name: "string", required: false },
    },
    block: {
      description: "make select full width",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
  },
};

export const SelectSizes = () => {
  return (
    <div className="flex flex-wrap gap-5">
      <TextInput placeholder="size xs" size="xs" />
      <TextInput placeholder="size sm" size="sm" />
      <TextInput placeholder="default" />
      <TextInput placeholder="size lg" size="lg" />
    </div>
  );
};
export const Variants = () => {
  return (
    <div className="flex flex-col gap-5">
      <TextInput placeholder="default" />
      <TextInput placeholder="variant primary" variant="primary" />
      <TextInput placeholder="variant secondary" variant="secondary" />
      <TextInput placeholder="variant accent" variant="accent" />
      <TextInput placeholder="variant ghost" variant="ghost" />
      <TextInput placeholder="variant info" variant="info" />
      <TextInput placeholder="variant success" variant="success" />
      <TextInput placeholder="variant error" variant="error" />
    </div>
  );
};

export const Block = () => (
  <div className="p-10">
    <TextInput block />
  </div>
);
