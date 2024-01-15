import type { Meta, StoryObj } from "@storybook/react";
import { Separator } from "./index";

const meta = {
  title: "Components/Separator",
  component: Separator,
} satisfies Meta<typeof Separator>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => (
    <div className={args.vertical ? "flex h-5 items-center space-x-2" : ""}>
      <div>Some</div>
      <Separator {...args} className="my-2" />
      <div>Content</div>
    </div>
  ),
  tags: ["autodocs"],
  argTypes: {
    vertical: {
      description: "user vertical separator",
      type: {
        name: "boolean",
        required: false,
      },
    },
  },
};

export const HorizontalSeparator = () => (
  <div>
    <div>The Separator Component UI</div>
    <Separator className="my-4" />
    <div>The Separator Component UI</div>
  </div>
);

export const VerticalSeparator = () => (
  <div className="flex h-5 items-center space-x-4">
    <div>Blog</div>
    <Separator vertical />
    <div>Docs</div>
    <Separator vertical />
    <div>Source</div>
  </div>
);
