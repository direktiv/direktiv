import type { Meta, StoryObj } from "@storybook/react";
import { Separator } from "./index";

const meta = {
  title: "Components/Separator",
  component: Separator,
} satisfies Meta<typeof Separator>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <div>
      <div className="space-y-1">
        <h4 className="text-sm font-medium leading-none">
          The Separator Component UI
        </h4>
        <p className="text-sm text-gray-9 dark:text-gray-dark-9">
          An open-source UI component library.
        </p>
      </div>
      <Separator className="my-4" />
      <div className="flex h-5 items-center space-x-4 text-sm">
        <div>Blog</div>
        <Separator orientation="vertical" />
        <div>Docs</div>
        <Separator orientation="vertical" />
        <div>Source</div>
      </div>
    </div>
  ),
  tags: ["autodocs"],
};
export const HorizontalSeparator = () => (
  <div>
    <div className="space-y-1">
      <h4 className="text-sm font-medium leading-none">
        The Separator Component UI
      </h4>
    </div>
    <Separator className="my-4" />
    <div className="space-y-1">
      <h4 className="text-sm font-medium leading-none">
        The Separator Component UI
      </h4>
    </div>
  </div>
);

export const VerticalSeparator = () => (
  <div className="flex h-5 items-center space-x-4 text-sm">
    <div>Blog</div>
    <Separator orientation="vertical" />
    <div>Docs</div>
    <Separator orientation="vertical" />
    <div>Source</div>
  </div>
);
