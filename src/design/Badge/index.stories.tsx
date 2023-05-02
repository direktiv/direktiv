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
      options: ["default", "secondary", "outline", "destructive"],
      type: { name: "string", required: false },
    },
  },
};

export const BadgeVariants = () => (
  <div>
    <Badge>default</Badge>
    <Badge variant="secondary">secondary</Badge>
    <Badge variant="outline">outline</Badge>
    <Badge variant="destructive">destructive</Badge>
  </div>
);
