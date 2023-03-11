import type { Meta, StoryObj } from "@storybook/react";
import TextInput from "./index";

const meta = {
  title: "Components (next)/TextInput",
  component: TextInput,
} satisfies Meta<typeof TextInput>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: (args) => {
    return <TextInput placeholder="default" {...args} />;
  },
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
    ghost: {
      description: "ghost select",
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
      <TextInput placeholder="size md" size="md" />
      <TextInput placeholder="size lg" size="lg" />
    </div>
  );
};

export const Ghost = () => (
  <div className="flex bg-base-300">
    <TextInput placeholder="ghost input" ghost />
  </div>
);

export const Block = () => (
  <div className=" p-10">
    <TextInput block />
  </div>
);
