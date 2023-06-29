import type { Meta, StoryObj } from "@storybook/react";
import Badge from "./index";

const meta = {
  title: "Components/Badge",
  component: Badge,
} satisfies Meta<typeof Badge>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => <Badge {...args}>Hi</Badge>,
  argTypes: {
    variant: {
      description: "Badge Variant",
      control: "select",
      options: [undefined, "secondary", "outline", "destructive", "success"],
      type: { name: "string", required: false },
    },
  },
};

export const BadgeVariants = () => (
  <div className="flex space-x-2">
    <Badge>default</Badge>
    <Badge variant="secondary">secondary</Badge>
    <Badge variant="outline">outline</Badge>
    <Badge variant="destructive">destructive</Badge>
    <Badge variant="success">success</Badge>
  </div>
);
