import type { Meta, StoryObj } from "@storybook/react";
import { RxFontItalic } from "react-icons/rx";
import { Toggle } from "./index";

const meta = {
  title: "Components/Toggle",
  component: Toggle,
} satisfies Meta<typeof Toggle>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => (
    <Toggle aria-label="Toggle italic" {...args}>
      <RxFontItalic className="h-4 w-4" />
    </Toggle>
  ),
};

export const OutlineToggle = () => (
  <Toggle aria-label="Toggle italic" outline>
    <RxFontItalic className="h-4 w-4" />
  </Toggle>
);

export const ToggleSize = () => (
  <div className="flex flex-row gap-2">
    <Toggle aria-label="Toggle italic" size="sm">
      <RxFontItalic className="h-4 w-4" />
    </Toggle>
    <Toggle aria-label="Toggle italic">
      <RxFontItalic className="h-4 w-4" />
    </Toggle>
    <Toggle aria-label="Toggle italic" size="lg">
      <RxFontItalic className="h-4 w-4" />
    </Toggle>
  </div>
);

export const DefaultPressed = () => (
  <Toggle aria-label="Toggle italic" defaultPressed>
    <RxFontItalic className="h-4 w-4" />
  </Toggle>
);
