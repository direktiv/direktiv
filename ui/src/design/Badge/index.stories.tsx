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

export const BadgeWithIcon = () => (
  <div className="flex flex-col gap-y-5">
    <div className="flex space-x-2">
      <Badge icon="complete">default complete icon</Badge>
      <Badge icon="complete" variant="secondary">
        secondary complete icon
      </Badge>
      <Badge icon="complete" variant="outline">
        outline complete icon
      </Badge>
      <Badge icon="complete" variant="destructive">
        destructive complete icon
      </Badge>
      <Badge icon="complete" variant="success">
        success complete icon
      </Badge>
    </div>
    <div className="flex space-x-2">
      <Badge icon="pending">default pending icon</Badge>
      <Badge icon="pending" variant="secondary">
        secondary pending icon
      </Badge>
      <Badge icon="pending" variant="outline">
        outline pending icon
      </Badge>
      <Badge icon="pending" variant="destructive">
        destructive pending icon
      </Badge>
      <Badge icon="pending" variant="success">
        success pending icon
      </Badge>
    </div>
    <div className="flex space-x-2">
      <Badge icon="failed">default failed icon</Badge>
      <Badge icon="failed" variant="secondary">
        secondary failed icon
      </Badge>
      <Badge icon="failed" variant="outline">
        outline failed icon
      </Badge>
      <Badge icon="failed" variant="destructive">
        destructive failed icon
      </Badge>
      <Badge icon="failed" variant="success">
        success failed icon
      </Badge>
    </div>{" "}
    <div className="flex space-x-2">
      <Badge icon="crashed">default crashed icon</Badge>
      <Badge icon="crashed" variant="secondary">
        secondary crashed icon
      </Badge>
      <Badge icon="crashed" variant="outline">
        outline crashed icon
      </Badge>
      <Badge icon="crashed" variant="destructive">
        destructive crashed icon
      </Badge>
      <Badge icon="crashed" variant="success">
        success crashed icon
      </Badge>
    </div>
  </div>
);
